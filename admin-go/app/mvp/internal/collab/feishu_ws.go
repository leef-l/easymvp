package collab

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/frame/g"
	larkevent "github.com/larksuite/oapi-sdk-go/v3/event"
	"github.com/larksuite/oapi-sdk-go/v3/event/dispatcher"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuWSClient 飞书 WebSocket 长连接客户端。
// 底层使用官方 SDK larkws.Client，确保帧解析、握手 ACK、分包合并都正确处理。
type FeishuWSClient struct {
	appID      string
	appSecret  string
	encryptKey string

	sdkClient *larkws.Client
	running   atomic.Bool
	stopCh    chan struct{}
	mu        sync.Mutex

	// OnEvent 事件处理回调，与 Webhook 模式共用同一处理逻辑
	OnEvent func(ctx context.Context, header map[string]interface{}, event map[string]interface{})
}

// NewFeishuWSClient 创建 WebSocket 客户端。
func NewFeishuWSClient(appID, appSecret, encryptKey string) *FeishuWSClient {
	return &FeishuWSClient{
		appID:      appID,
		appSecret:  appSecret,
		encryptKey: encryptKey,
		stopCh:     make(chan struct{}),
	}
}

// Start 启动长连接（阻塞直到 Stop 被调用）。
func (c *FeishuWSClient) Start(ctx context.Context) {
	if !c.running.CompareAndSwap(false, true) {
		return
	}
	defer c.running.Store(false)

	// 构建事件分发器：注册所有需要处理的事件类型
	rawHandler := func(ctx context.Context, req *larkevent.EventReq) error {
		c.handleRawPayload(ctx, req.Body)
		return nil
	}
	noopHandler := func(ctx context.Context, req *larkevent.EventReq) error { return nil }
	d := dispatcher.NewEventDispatcher("", "").
		OnCustomizedEvent("im.message.receive_v1", rawHandler).
		OnCustomizedEvent("application.bot.menu_v6", rawHandler).
		OnCustomizedEvent("im.chat.member.bot.added_v1", rawHandler).
		OnCustomizedEvent("im.chat.member.bot.deleted_v1", rawHandler).
		OnCustomizedEvent("im.message.message_read_v1", noopHandler)

	sdkCli := larkws.NewClient(c.appID, c.appSecret,
		larkws.WithEventHandler(d),
		larkws.WithAutoReconnect(true),
	)

	c.mu.Lock()
	c.sdkClient = sdkCli
	c.mu.Unlock()

	g.Log().Infof(ctx, "[FeishuWS] 启动官方 SDK 长连接 appID=%s", c.appID)

	// Start 是阻塞的，内部自动重连
	if err := sdkCli.Start(ctx); err != nil {
		g.Log().Warningf(ctx, "[FeishuWS] 长连接退出: %v", err)
	}
}

// Stop 停止长连接。
func (c *FeishuWSClient) Stop() {
	select {
	case <-c.stopCh:
	default:
		close(c.stopCh)
	}
	// 官方 SDK 通过 ctx cancel 来停止，由 manager 取消 ctx
}

// IsRunning 返回是否正在运行。
func (c *FeishuWSClient) IsRunning() bool {
	return c.running.Load()
}

// handleRawPayload 解析官方 SDK 传来的原始事件 JSON，转发给 OnEvent 回调。
// payload 格式：{"schema":"2.0","header":{...},"event":{...}}
func (c *FeishuWSClient) handleRawPayload(ctx context.Context, body []byte) {
	if len(body) == 0 {
		return
	}

	g.Log().Infof(ctx, "[FeishuWS] 原始payload: %s", string(body))

	var raw map[string]interface{}
	if err := json.Unmarshal(body, &raw); err != nil {
		g.Log().Warningf(ctx, "[FeishuWS] payload JSON 解析失败: %v, body=%s", err, string(body))
		return
	}

	header, _ := raw["header"].(map[string]interface{})
	event, _ := raw["event"].(map[string]interface{})

	if header != nil {
		et, _ := header["event_type"].(string)
		g.Log().Infof(ctx, "[FeishuWS] 收到事件: event_type=%s", et)
	}

	if c.OnEvent != nil && (header != nil || event != nil) {
		c.OnEvent(ctx, header, event)
	}
}

// FeishuAESDecrypt 解密飞书 AES-256-CBC 加密的 base64 密文。
// key = SHA256(encryptKey)[:32]；IV = 密文前16字节。
// 用于 Webhook 模式的消息体解密（WS 模式不需要此函数）。
func FeishuAESDecrypt(cipherB64, encryptKey string) ([]byte, error) {
	cipherData, err := base64.StdEncoding.DecodeString(cipherB64)
	if err != nil {
		return nil, fmt.Errorf("base64解码失败: %w", err)
	}
	if len(cipherData) < aes.BlockSize {
		return nil, fmt.Errorf("密文太短: %d", len(cipherData))
	}

	key := feishuSHA256([]byte(encryptKey))[:32]

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("创建AES块失败: %w", err)
	}

	iv := cipherData[:aes.BlockSize]
	cipherData = cipherData[aes.BlockSize:]

	if len(cipherData)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("密文长度不是块大小的倍数")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(cipherData, cipherData)

	plain, err := pkcs7Unpad(cipherData)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func feishuSHA256(data []byte) []byte {
	h := sha256.New()
	h.Write(data)
	return h.Sum(nil)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, fmt.Errorf("空数据")
	}
	padding := int(data[length-1])
	if padding > aes.BlockSize || padding == 0 {
		return nil, fmt.Errorf("无效的 padding: %d", padding)
	}
	if length < padding {
		return nil, fmt.Errorf("数据长度小于 padding")
	}
	for _, b := range data[length-padding:] {
		if int(b) != padding {
			return nil, fmt.Errorf("padding 字节不一致")
		}
	}
	return data[:length-padding], nil
}
