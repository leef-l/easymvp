#!/bin/bash
# 腾讯云 Coding Plan 模型能力测试
API_KEY="sk-sp-ErlXWrTOnFVKB4kNZkJcyN6mHnLBn5d9nyVS8e0QmS4eoYih"
API_URL="https://api.lkeap.cloud.tencent.com/coding/v3/chat/completions"

MODELS=("tc-code-latest" "hunyuan-2.0-instruct" "hunyuan-2.0-thinking" "hunyuan-t1" "hunyuan-turbos" "glm-5" "kimi-k2.5" "minimax-m2.5")

call_model() {
  local model=$1
  local prompt=$2
  local max_tokens=${3:-500}

  START=$(date +%s%N)
  RESP=$(curl -s --max-time 120 "$API_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $API_KEY" \
    -d "{
      \"model\": \"$model\",
      \"messages\": [{\"role\": \"user\", \"content\": $(echo "$prompt" | jq -Rs .)}],
      \"max_tokens\": $max_tokens,
      \"temperature\": 0.7
    }" 2>&1)
  END=$(date +%s%N)
  DURATION=$(( (END - START) / 1000000 ))

  # 提取内容
  ERROR=$(echo "$RESP" | jq -r '.error.message // empty' 2>/dev/null)
  if [ -n "$ERROR" ]; then
    echo "❌ ERROR: $ERROR (${DURATION}ms)"
    return 1
  fi

  CONTENT=$(echo "$RESP" | jq -r '.choices[0].message.content // empty' 2>/dev/null)
  REASONING=$(echo "$RESP" | jq -r '.choices[0].message.reasoning_content // empty' 2>/dev/null)
  TOKENS_IN=$(echo "$RESP" | jq -r '.usage.prompt_tokens // 0' 2>/dev/null)
  TOKENS_OUT=$(echo "$RESP" | jq -r '.usage.completion_tokens // 0' 2>/dev/null)
  FINISH=$(echo "$RESP" | jq -r '.choices[0].finish_reason // empty' 2>/dev/null)

  echo "⏱ ${DURATION}ms | tokens: ${TOKENS_IN}→${TOKENS_OUT} | finish: ${FINISH}"
  if [ -n "$REASONING" ] && [ "$REASONING" != "null" ]; then
    echo "🧠 思考: $(echo "$REASONING" | head -c 200)..."
  fi
  if [ -n "$CONTENT" ]; then
    echo "$CONTENT"
  else
    echo "(无内容输出，仅有reasoning)"
  fi
  return 0
}

echo "╔══════════════════════════════════════════════════════════════╗"
echo "║       腾讯云 Coding Plan 模型能力对比测试                    ║"
echo "║       测试时间: $(date '+%Y-%m-%d %H:%M:%S')                     ║"
echo "╚══════════════════════════════════════════════════════════════╝"
echo ""

# ═══════════════════════════════════════════
# 测试1: 身份识别
# ═══════════════════════════════════════════
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "【测试1】身份识别 — 你是什么模型?"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for model in "${MODELS[@]}"; do
  echo ""
  echo "▶ $model"
  call_model "$model" "你是什么模型？一句话回答你的真实名称和版本。" 100
done

# ═══════════════════════════════════════════
# 测试2: 代码生成
# ═══════════════════════════════════════════
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "【测试2】代码生成 — Go并发LRU缓存"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
CODE_PROMPT="用Go写一个并发安全的LRU缓存，支持Get(key)和Put(key,value)，容量为n。只输出代码，不要解释。"
for model in "${MODELS[@]}"; do
  echo ""
  echo "▶ $model"
  call_model "$model" "$CODE_PROMPT" 800
done

# ═══════════════════════════════════════════
# 测试3: 逻辑推理
# ═══════════════════════════════════════════
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "【测试3】逻辑推理 — 经典灯泡问题"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
LOGIC_PROMPT="一个房间有3个开关控制隔壁房间3盏灯。你只能去隔壁房间一次。如何确定哪个开关对应哪盏灯？简洁回答。"
for model in "${MODELS[@]}"; do
  echo ""
  echo "▶ $model"
  call_model "$model" "$LOGIC_PROMPT" 300
done

# ═══════════════════════════════════════════
# 测试4: 架构设计
# ═══════════════════════════════════════════
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "【测试4】架构设计 — AI任务调度系统"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
ARCH_PROMPT="设计一个AI任务并行调度系统，需要支持：1)批次调度 2)依赖检查 3)资源冲突检测 4)失败重试。给出核心数据结构和调度算法，100字以内。"
for model in "${MODELS[@]}"; do
  echo ""
  echo "▶ $model"
  call_model "$model" "$ARCH_PROMPT" 500
done

echo ""
echo "╔══════════════════════════════════════════════════════════════╗"
echo "║                    测试完成                                  ║"
echo "╚══════════════════════════════════════════════════════════════╝"
