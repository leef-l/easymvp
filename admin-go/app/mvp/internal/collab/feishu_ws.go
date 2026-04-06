package collab

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gorilla/websocket"
	larkws "github.com/larksuite/oapi-sdk-go/v3/ws"
)

// FeishuWSClient 飞书 WebSocket 长连接客户端。
// 文档：https://open.feishu.cn/document/server-docs/event-subscription-guide/long-connection
// 帧格式：protobuf（larkws.Frame），payload 为事件 JSON（WS 模式不加密 payload）。
// Webhook 模式下的 EncryptKey 用于 AES 解密，WS 模式下暂不使用（飞书 WS 不走加密）。
type FeishuWSClient struct {
	appID      string
	appSecret  string
	encryptKey string // 保留字段，WS 模式暂不使用

	conn    *websocket.Conn
	mu      sync.Mutex
	running atomic.Bool
	stopCh  chan struct{}

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
		return // 已在运行
	}
	defer c.running.Store(false)

	backoff := 5 * time.Second
	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		g.Log().Infof(ctx, "[FeishuWS] 尝试建立长连接...")
		if err := c.connect(ctx); err != nil {
			g.Log().Warningf(ctx, "[FeishuWS] 连接失败: %v，%v 后重试", err, backoff)
			select {
			case <-c.stopCh:
				return
			case <-time.After(backoff):
			}
			if backoff < 5*time.Minute {
				backoff *= 2
			}
			continue
		}

		g.Log().Infof(ctx, "[FeishuWS] 长连接已建立")
		backoff = 5 * time.Second // 重置退避

		c.readLoop(ctx)

		g.Log().Infof(ctx, "[FeishuWS] 连接断开，准备重连...")
	}
}

// Stop 停止长连接。
func (c *FeishuWSClient) Stop() {
	select {
	case <-c.stopCh:
	default:
		close(c.stopCh)
	}
	c.mu.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.mu.Unlock()
}

// IsRunning 返回是否正在运行。
func (c *FeishuWSClient) IsRunning() bool {
	return c.running.Load()
}

// connect 获取 WS endpoint 并建立连接。
func (c *FeishuWSClient) connect(ctx context.Context) error {
	// 1. 获取 tenant_access_token
	token, err := c.getToken(ctx)
	if err != nil {
		return fmt.Errorf("获取 token 失败: %w", err)
	}

	// 2. 获取 WebSocket URL（飞书长连接 endpoint）
	wsURL, err := c.getWSEndpoint(ctx, token)
	if err != nil {
		return fmt.Errorf("获取 WS endpoint 失败: %w", err)
	}

	// 3. 建立 WebSocket 连接
	header := http.Header{
		"Authorization": []string{"Bearer " + token},
	}
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, wsURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket 连接失败: %w", err)
	}

	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()

	// 4. 启动心跳（protobuf ping frame）
	go c.pingLoop(ctx)

	return nil
}

// readLoop 持续读取 WS 消息并分发。
func (c *FeishuWSClient) readLoop(ctx context.Context) {
	for {
		c.mu.Lock()
		conn := c.conn
		c.mu.Unlock()
		if conn == nil {
			return
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			g.Log().Warningf(ctx, "[FeishuWS] 读取消息失败: %v", err)
			return
		}

		go c.dispatch(ctx, msg)
	}
}

// dispatch 解析并分发一条 WS 消息。
// 飞书 WS 帧是 protobuf 编码（larkws.Frame），使用官方 SDK 解析。
// FrameType=0(Control): ping/pong；FrameType=1(Data): 事件。
// Payload 是事件 JSON，WS 模式下不加密。
func (c *FeishuWSClient) dispatch(ctx context.Context, msg []byte) {
	var frame larkws.Frame
	if err := frame.Unmarshal(msg); err != nil {
		g.Log().Warningf(ctx, "[FeishuWS] 帧解析失败: %v", err)
		return
	}

	headers := larkws.Headers(frame.Headers)
	msgType := larkws.MessageType(headers.GetString(larkws.HeaderType))

	switch larkws.FrameType(frame.Method) {
	case larkws.FrameTypeControl:
		// pong / 握手确认，忽略
		return
	case larkws.FrameTypeData:
		if msgType != larkws.MessageTypeEvent {
			return
		}
	default:
		return
	}

	// Payload 是事件 JSON
	payload := frame.Payload
	if len(payload) == 0 {
		return
	}

	// 解析事件
	var raw map[string]interface{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		g.Log().Warningf(ctx, "[FeishuWS] 事件 JSON 解析失败: %v", err)
		return
	}

	header, _ := raw["header"].(map[string]interface{})
	event, _ := raw["event"].(map[string]interface{})

	if c.OnEvent != nil && (header != nil || event != nil) {
		c.OnEvent(ctx, header, event)
	}
}

// pingLoop 每 30 秒发送一次心跳（protobuf ping frame）。
func (c *FeishuWSClient) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.mu.Lock()
			conn := c.conn
			c.mu.Unlock()
			if conn == nil {
				return
			}
			pingFrame := larkws.NewPingFrame(0)
			bs, err := pingFrame.Marshal()
			if err != nil {
				g.Log().Warningf(ctx, "[FeishuWS] 心跳序列化失败: %v", err)
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, bs); err != nil {
				g.Log().Warningf(ctx, "[FeishuWS] 心跳发送失败: %v", err)
				return
			}
		}
	}
}

// getToken 获取 tenant_access_token。
func (c *FeishuWSClient) getToken(ctx context.Context) (string, error) {
	resp, err := g.Client().Post(ctx,
		"https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal",
		g.Map{"app_id": c.appID, "app_secret": c.appSecret})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var result struct {
		Code              int    `json:"code"`
		Msg               string `json:"msg"`
		TenantAccessToken string `json:"tenant_access_token"`
	}
	if err := json.Unmarshal(resp.ReadAll(), &result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("code=%d msg=%s", result.Code, result.Msg)
	}
	return result.TenantAccessToken, nil
}

// getWSEndpoint 调用飞书官方接口获取 WebSocket 长连接地址。
// 接口：POST https://open.feishu.cn/callback/ws/endpoint
// 请求体：{"AppID":"...","AppSecret":"..."}  响应：{"data":{"URL":"wss://..."}}
func (c *FeishuWSClient) getWSEndpoint(ctx context.Context, _ string) (string, error) {
	resp, err := g.Client().
		SetHeaderMap(map[string]string{
			"Content-Type": "application/json",
		}).
		Post(ctx,
			"https://open.feishu.cn/callback/ws/endpoint",
			g.Map{"AppID": c.appID, "AppSecret": c.appSecret})
	if err != nil {
		return "", err
	}
	defer resp.Close()

	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			URL string `json:"URL"`
		} `json:"data"`
	}
	body := resp.ReadAll()
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析 endpoint 响应失败: %w, body=%s", err, string(body))
	}
	if result.Code != 0 {
		return "", fmt.Errorf("code=%d msg=%s", result.Code, result.Msg)
	}
	if result.Data.URL == "" {
		return "", fmt.Errorf("飞书返回的 WS URL 为空, body=%s", string(body))
	}
	return result.Data.URL, nil
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
