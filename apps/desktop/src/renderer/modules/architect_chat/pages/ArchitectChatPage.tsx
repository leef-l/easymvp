import { useCallback, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { useProjectState } from "@/shared/lib/project";
import { getConversation, listMessages, sendMessageStream, confirmPlan, type MessageItem } from "../api";

export function ArchitectChatPage() {
  const { t } = useTranslation();
  const { projectId } = useProjectState();
  const [messages, setMessages] = useState<MessageItem[]>([]);
  const [input, setInput] = useState("");
  const [loading, setLoading] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [status, setStatus] = useState<string>("");
  const bottomRef = useRef<HTMLDivElement>(null);
  const scrollContainerRef = useRef<HTMLDivElement>(null);
  const isAtBottomRef = useRef(true);

  const checkIsAtBottom = () => {
    const el = scrollContainerRef.current;
    if (!el) return true;
    const threshold = 60;
    return el.scrollTop + el.clientHeight >= el.scrollHeight - threshold;
  };

  const handleScroll = () => {
    isAtBottomRef.current = checkIsAtBottom();
  };

  const scrollToBottom = () => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  const load = useCallback(async () => {
    if (!projectId) return;
    try {
      const items = await listMessages(projectId);
      setMessages(items);
    } catch {
      // ignore polling errors
    }
  }, [projectId]);

  useEffect(() => {
    if (!projectId) return;
    getConversation(projectId)
      .then((conv) => setStatus(conv.status))
      .catch(() => {});
    load();
    const id = setInterval(load, 2000);
    return () => clearInterval(id);
  }, [projectId, load]);

  // Only auto-scroll when user sends a message, not when receiving stream deltas.
  const scrollOnSend = () => {
    requestAnimationFrame(() => {
      scrollToBottom();
    });
  };

  const handleSend = async () => {
    if (!projectId || !input.trim() || loading) return;
    const trimmed = input.trim();
    setInput("");
    setLoading(true);
    scrollOnSend();

    // Optimistically append user message
    const userMsg: MessageItem = {
      id: "temp-user-" + Date.now(),
      sender_role: "user",
      sender_name: "User",
      content: trimmed,
      message_kind: "chat",
      created_at: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, userMsg]);

    // Placeholder for streaming architect reply
    const placeholderId = "temp-architect-" + Date.now();
    setMessages((prev) => [
      ...prev,
      {
        id: placeholderId,
        sender_role: "architect",
        sender_name: "Architect",
        content: "",
        message_kind: "chat",
        created_at: new Date().toISOString(),
      },
    ]);

    let replyText = "";
    try {
      for await (const ev of sendMessageStream(projectId, trimmed)) {
        if (ev.type === "llm.content_delta" && ev.content) {
          replyText += ev.content;
          setMessages((prev) =>
            prev.map((m) =>
              m.id === placeholderId ? { ...m, content: replyText } : m
            )
          );
        }
        if (ev.type === "execution.done" && ev.content) {
          replyText = ev.content;
          setMessages((prev) =>
            prev.map((m) =>
              m.id === placeholderId ? { ...m, content: replyText } : m
            )
          );
        }
        if (ev.type === "error" && ev.content) {
          replyText = ev.content;
          setMessages((prev) =>
            prev.map((m) =>
              m.id === placeholderId ? { ...m, content: replyText } : m
            )
          );
          break;
        }
        if (ev.done) break;
      }
    } catch (e: any) {
      replyText = t("architectChat.sendFailed") + ": " + (e?.message ?? "");
      setMessages((prev) =>
        prev.map((m) =>
          m.id === placeholderId ? { ...m, content: replyText } : m
        )
      );
    } finally {
      setLoading(false);
      await load();
    }
  };

  const handleConfirm = async () => {
    if (!projectId || confirming) return;
    if (!window.confirm(t("architectChat.confirmPrompt"))) return;
    setConfirming(true);
    try {
      const res = await confirmPlan(projectId);
      if (res.accepted) {
        alert(t("architectChat.confirmSuccess"));
        setStatus("confirmed");
      } else {
        alert(t("architectChat.confirmRejected"));
      }
      await load();
    } catch (e: any) {
      alert(t("architectChat.confirmFailed") + ": " + (e?.message ?? ""));
    } finally {
      setConfirming(false);
    }
  };

  const roleClass = (role: string) => {
    switch (role) {
      case "user":
        return "msg-user";
      case "architect":
        return "msg-architect";
      case "reviewer":
        return "msg-reviewer";
      case "system":
        return "msg-system";
      default:
        return "msg-other";
    }
  };

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <p>{t("project.selectFirst")}</p>
      </section>
    );
  }

  return (
    <section style={{ display: "flex", flexDirection: "column", height: "100%", padding: 16, boxSizing: "border-box" }}>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 12 }}>
        <h2>{t("architectChat.title")}</h2>
        <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
          {status === "confirmed" && <span className="status-pill">{t("architectChat.statusConfirmed")}</span>}
          <button
            className="primary-button"
            onClick={handleConfirm}
            disabled={confirming || status === "confirmed"}
          >
            {confirming ? t("architectChat.confirming") : t("architectChat.confirmPlan")}
          </button>
        </div>
      </div>

      <div
        ref={scrollContainerRef}
        onScroll={handleScroll}
        style={{
          flex: 1,
          overflowY: "auto",
          maxHeight: "calc(100vh - 240px)",
          border: "1px solid var(--border-color, #e5e7eb)",
          borderRadius: 8,
          padding: 12,
          background: "var(--surface-bg, #fff)",
          marginBottom: 12,
        }}
      >
        {messages.length === 0 && (
          <p style={{ color: "#999", textAlign: "center", marginTop: 40 }}>{t("architectChat.empty")}</p>
        )}
        {messages.map((m) => (
          <div
            key={m.id}
            style={{
              marginBottom: 12,
              display: "flex",
              flexDirection: "column",
              alignItems: m.sender_role === "user" ? "flex-end" : "flex-start",
            }}
          >
            <div style={{ fontSize: 12, color: "#666", marginBottom: 4 }}>
              {m.sender_name}
            </div>
            <div
              style={{
                maxWidth: "80%",
                padding: "10px 14px",
                borderRadius: 12,
                fontSize: 14,
                lineHeight: 1.5,
                whiteSpace: "pre-wrap",
                background:
                  m.sender_role === "user"
                    ? "#dbeafe"
                    : m.sender_role === "architect"
                    ? "#f3f4f6"
                    : m.sender_role === "reviewer"
                    ? "#fef3c7"
                    : "#e5e7eb",
                color: "#111",
              }}
            >
              {m.content}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div style={{ display: "flex", gap: 8 }}>
        <textarea
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              handleSend();
            }
          }}
          placeholder={t("architectChat.inputPlaceholder")}
          disabled={loading || status === "confirmed"}
          style={{
            flex: 1,
            resize: "none",
            height: 60,
            padding: 10,
            borderRadius: 8,
            border: "1px solid #d1d5db",
            fontSize: 14,
          }}
        />
        <button
          className="primary-button"
          onClick={handleSend}
          disabled={loading || !input.trim() || status === "confirmed"}
          style={{ alignSelf: "flex-end", height: 60, padding: "0 20px" }}
        >
          {loading ? t("architectChat.sending") : t("architectChat.send")}
        </button>
      </div>
    </section>
  );
}
