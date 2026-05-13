import { useEffect, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { Avatar, Badge, Button, Card, Composer, ConversationRow, MessageBubble } from "@game/ui";
import { conversations } from "@game/shared";
import type { DemoTokenResponse } from "@game/api";
import { useIMStore } from "../store/imStore";
import { useSettingsStore } from "../store/settingsStore";
import { useVoiceStore } from "../store/voiceStore";
import { useLiveKit } from "../hooks/useLiveKit";
import { getHttp } from "../services/http";

export function DemoPage() {
  const navigate = useNavigate();
  const { livekitUrl } = useSettingsStore();
  const {
    loadConversations, setActiveConversation,
    status, error, conversations: convs, messages, activeConversationId,
    sendMessage, sendVoiceAction,
  } = useIMStore();

  const callState = useVoiceStore((s) => s.callState);
  const callId = useVoiceStore((s) => s.callId);
  const duration = useVoiceStore((s) => s.duration);
  const isMuted = useVoiceStore((s) => s.isMuted);
  const setCall = useVoiceStore((s) => s.setCall);
  const resetCall = useVoiceStore((s) => s.resetCall);
  const tickDuration = useVoiceStore((s) => s.tickDuration);
  const lk = useLiveKit();

  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const [roomSize, setRoomSize] = useState(5);

  const handleCreateRoom = () => {
    const roomId = crypto.randomUUID();
    sessionStorage.setItem(`room-size-${roomId}`, String(roomSize));
    navigate(`/demo/room/${roomId}`);
  };

  // IM is bootstrapped globally via useIMBootstrap in AppRoutes

  // call duration timer
  useEffect(() => {
    if (callState === "connected") {
      timerRef.current = setInterval(tickDuration, 1000);
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [callState, tickDuration]);

  const handleCall = async () => {
    try {
      const data = await getHttp().post<DemoTokenResponse>("/api/v1/demo/token", {});
      const [, lkToken] = data.accessToken.split("|");
      const callId = `demo-${Date.now()}`;
      const room = `demo-room`;
      setCall({ callId, token: lkToken, room, url: livekitUrl });
      sendVoiceAction("call_invite", {
        call_id: callId,
        livekit_url: livekitUrl,
        livekit_token: lkToken,
        livekit_room: room,
      });
    } catch (e) {
      console.error("Failed to get demo token:", e);
    }
  };

  const handleHangUp = () => {
    sendVoiceAction("call_end", { call_id: callId });
    lk.disconnect();
    resetCall();
  };

  const handleToggleMute = () => {
    lk.toggleMic();
  };

  const activeConv = convs.find((c) => c.id === activeConversationId);

  const callLabel =
    callState === "idle" ? "就绪"
    : callState === "ringing" ? "呼叫中..."
    : callState === "connected" ? "通话中"
    : "已结束";

  const callTone =
    callState === "connected" ? "success"
    : callState === "ringing" ? "warning"
    : "neutral";

  return (
    <div style={{ maxWidth: 960, margin: "0 auto", padding: "24px 16px", minHeight: "100vh", display: "flex", flexDirection: "column" }}>
      {/* Header */}
      <header style={{ marginBottom: 20 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700, margin: 0 }}>Game Backend Cloud</h1>
        <p style={{ fontSize: 14, color: "#667085", margin: "4px 0 0" }}>
          实时通信能力演示 — IM 消息 + WebRTC 语音
        </p>
      </header>

      {/* Connection status bar */}
      <div style={{
        display: "flex", gap: 12, alignItems: "center", padding: "8px 12px",
        background: "#f9fafb", borderRadius: 8, marginBottom: 16, fontSize: 13,
      }}>
        <Badge tone={status === "connected" ? "success" : status === "connecting" ? "warning" : "neutral"}>
          IM: {status === "connected" ? "已连接" : status === "idle" ? "空闲" : status}
        </Badge>
        <Badge tone={callTone}>语音: {callLabel}</Badge>
        {callState === "connected" ? (
          <span style={{ fontWeight: 700, color: "#166534" }}>
            {String(Math.floor(duration / 60)).padStart(2, "0")}:
            {String(duration % 60).padStart(2, "0")}
          </span>
        ) : null}
      </div>

      {/* Room Creation */}
      <div style={{
        display: "flex", gap: 12, alignItems: "center", padding: "12px 16px",
        background: "#f0fdf4", borderRadius: 12, marginBottom: 16,
        border: "1px solid #bbf7d0",
      }}>
        <div style={{ flex: 1 }}>
          <h3 style={{ fontSize: 15, fontWeight: 700, margin: 0, color: "#166534" }}>多人语音房间</h3>
          <p style={{ fontSize: 13, color: "#15803d", margin: "2px 0 0" }}>
            创建房间后分享链接，对方打开即可加入通话
          </p>
        </div>
        <select
          value={roomSize}
          onChange={(e) => setRoomSize(Number(e.target.value))}
          style={{
            padding: "6px 10px", borderRadius: 6, border: "1px solid #86efac",
            fontSize: 13, background: "#fff",
          }}
        >
          <option value={5}>5 人房间</option>
          <option value={10}>10 人房间</option>
        </select>
        <Button variant="primary" onClick={handleCreateRoom}>
          创建房间
        </Button>
      </div>

      {/* Main two-column layout */}
      <div style={{ display: "flex", gap: 16, flex: 1 }}>
        {/* Left: Conversation List */}
        <div style={{ width: 260, flexShrink: 0 }}>
          <Card title="会话" subtitle={convs.length ? `${convs.length} 个联系人` : undefined}>
            <div className="conversationList">
              {convs.map((c) => (
                <ConversationRow
                  key={c.id}
                  conversation={c}
                  active={c.id === activeConversationId}
                  onClick={() => setActiveConversation(c.id)}
                />
              ))}
            </div>
          </Card>
        </div>

        {/* Right: Messages + Voice Controls */}
        <div className="feed" style={{ flex: 1 }}>
          <Card
            title={activeConv ? `与 ${activeConv.title} 的对话` : "选择会话"}
            subtitle={activeConv && activeConv.online ? "在线" : undefined}
          >
            <div className="chatFeed" style={{ minHeight: 240 }}>
              {messages.map((m) => (
                <MessageBubble key={m.id} message={m} />
              ))}
              {messages.length === 0 && activeConv ? (
                <div className="emptyState">发送第一条消息开始对话</div>
              ) : null}
            </div>
            {activeConv ? (
              <Composer
                placeholder="输入消息..."
                onSend={(text) => {
                  if (activeConv) sendMessage(activeConv.userId, text);
                }}
              />
            ) : null}
          </Card>

          {/* Voice Call Controls */}
          <Card title="语音通话" subtitle="基于 LiveKit WebRTC">
            <div className="stack">
              <div className="buttonGrid">
                {callState !== "connected" && callState !== "ringing" ? (
                  <Button variant="primary" onClick={handleCall}>发起语音通话</Button>
                ) : null}
                {callState === "connected" || callState === "ringing" ? (
                  <>
                    <Button variant={isMuted ? "ghost" : "secondary"} onClick={handleToggleMute}>
                      {isMuted ? "取消静音" : "静音"}
                    </Button>
                    <Button variant="ghost" onClick={handleHangUp}>挂断</Button>
                  </>
                ) : null}
              </div>
              {callState === "idle" ? (
                <p style={{ fontSize: 13, color: "#667085", margin: 0 }}>
                  打开两个浏览器 Tab，点击「发起语音通话」即可体验端到端 WebRTC 通话。
                </p>
              ) : null}
            </div>
          </Card>
        </div>
      </div>

      {/* Footer */}
      <footer style={{ marginTop: 32, paddingTop: 16, borderTop: "1px solid #e5e7eb", textAlign: "center" }}>
        <a
          href="https://github.com"
          target="_blank"
          rel="noopener noreferrer"
          style={{ fontSize: 13, color: "#667085", textDecoration: "none" }}
        >
          Powered by Game Backend Cloud
        </a>
      </footer>
    </div>
  );
}
