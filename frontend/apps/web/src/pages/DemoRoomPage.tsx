import { useEffect, useRef, useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { Card, Button, Badge, Composer, MessageBubble } from "@game/ui";
import { generateLiveKitToken } from "@game/api";
import { useRoomStore } from "../store/roomStore";
import { useSettingsStore } from "../store/settingsStore";

const DEMO_API_KEY = "devkey";
const DEMO_API_SECRET = "this-is-a-32-character-secret-key!!";

function getDemoIdentity(): string {
  const key = "demo-identity";
  let id = localStorage.getItem(key);
  if (!id) {
    id = `user-${Math.random().toString(36).slice(2, 8)}`;
    localStorage.setItem(key, id);
  }
  return id;
}

function getRoomSize(roomId: string): number {
  const stored = sessionStorage.getItem(`room-size-${roomId}`);
  return stored ? parseInt(stored, 10) : 5;
}

type RoomMessage = {
  id: string;
  text: string;
  sender: string;
  time: string;
};

function loadMessages(roomId: string): RoomMessage[] {
  try {
    const stored = sessionStorage.getItem(`room-msgs-${roomId}`);
    return stored ? JSON.parse(stored) : [];
  } catch { return []; }
}

function saveMessages(roomId: string, msgs: RoomMessage[]): void {
  sessionStorage.setItem(`room-msgs-${roomId}`, JSON.stringify(msgs.slice(-200)));
}

export function DemoRoomPage() {
  const { roomId } = useParams<{ roomId: string }>();
  const navigate = useNavigate();
  const livekitUrl = useSettingsStore((s) => s.livekitUrl);

  const { participants, duration, error, lk, joinRoom, leaveRoom, toggleMute, tickDuration, setError } =
    useRoomStore();

  const [connecting, setConnecting] = useState(true);
  const [messages, setMessages] = useState<RoomMessage[]>(() => (roomId ? loadMessages(roomId) : []));
  const [copied, setCopied] = useState(false);
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const myIdentity = useRef(getDemoIdentity());
  const roomSizeRef = useRef(5);
  const messagesRef = useRef(messages);
  messagesRef.current = messages;

  // Connect to room on mount
  useEffect(() => {
    if (!roomId) return;

    const size = getRoomSize(roomId);
    roomSizeRef.current = size;
    const roomName = `demo-room-${roomId}`;

    generateLiveKitToken({
      roomName,
      participantId: myIdentity.current,
      apiKey: DEMO_API_KEY,
      apiSecret: DEMO_API_SECRET,
      ttl: 3600,
    })
      .then((token) => joinRoom(roomId, size, token, livekitUrl))
      .catch((e) => {
        setError(e instanceof Error ? e.message : "连接房间失败");
      })
      .finally(() => setConnecting(false));

    return () => {
      leaveRoom();
    };
  }, [roomId]); // eslint-disable-line react-hooks/exhaustive-deps

  // Duration timer
  useEffect(() => {
    if (lk) {
      timerRef.current = setInterval(tickDuration, 1000);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [lk, tickDuration]);

  const handleCopyLink = useCallback(() => {
    const url = window.location.href;
    navigator.clipboard.writeText(url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  }, []);

  const handleLeave = useCallback(() => {
    leaveRoom();
    navigate("/demo");
  }, [leaveRoom, navigate]);

  const handleSend = useCallback((text: string) => {
    setMessages((prev) => {
      const next = [
        ...prev,
        {
          id: `msg-${Date.now()}`,
          text,
          sender: myIdentity.current,
          time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
        },
      ];
      if (roomId) saveMessages(roomId, next);
      return next;
    });
  }, [roomId]);

  const isMuted = lk?.isMuted ?? false;
  const isConnected = !!lk;
  const roomSize = roomSizeRef.current;
  const roomUrl = window.location.href;

  return (
    <div style={{ maxWidth: 960, margin: "0 auto", padding: "24px 16px", minHeight: "100vh", display: "flex", flexDirection: "column" }}>
      {/* Header */}
      <header style={{ marginBottom: 20 }}>
        <h1 style={{ fontSize: 24, fontWeight: 700, margin: 0 }}>Demo Room</h1>
        <p style={{ fontSize: 14, color: "#667085", margin: "4px 0 0" }}>
          多人语音房间 — 分享链接即可加入
        </p>
      </header>

      {/* Invite bar — always visible */}
      <div style={{
        padding: "20px 24px",
        background: "linear-gradient(135deg, #1e40af 0%, #3b82f6 50%, #6366f1 100%)",
        borderRadius: 16, marginBottom: 20,
        boxShadow: "0 4px 24px rgba(59,130,246,0.25)",
      }}>
        <h3 style={{ fontSize: 17, fontWeight: 800, margin: 0, color: "#fff" }}>
          邀请其他人加入房间
        </h3>
        <p style={{ fontSize: 14, color: "rgba(255,255,255,0.85)", margin: "6px 0 12px" }}>
          复制链接发给朋友，打开浏览器即可加入通话
        </p>
        <div style={{
          display: "flex", alignItems: "center", gap: 10,
          background: "rgba(0,0,0,0.2)", padding: "4px 4px 4px 16px", borderRadius: 10,
          border: "1px solid rgba(255,255,255,0.15)",
        }}>
          <code style={{
            fontSize: 14, color: "#f0f9ff", wordBreak: "break-all", flex: 1,
            userSelect: "all", fontWeight: 600, letterSpacing: "0.02em",
          }}>
            {roomUrl}
          </code>
          <button
            onClick={handleCopyLink}
            style={{
              padding: "10px 20px", borderRadius: 8, border: "none",
              background: copied ? "#22c55e" : "#fff",
              color: copied ? "#fff" : "#1e40af",
              fontSize: 14, fontWeight: 700, cursor: "pointer",
              whiteSpace: "nowrap", transition: "all 0.15s",
              boxShadow: copied ? "none" : "0 2px 8px rgba(0,0,0,0.1)",
            }}
          >
            {copied ? "已复制 ✓" : "复制链接"}
          </button>
        </div>
      </div>

      {/* Connection status banner */}
      {connecting ? (
        <div style={{
          padding: "8px 16px", background: "#fef9c3", borderRadius: 8, marginBottom: 16,
          fontSize: 13, color: "#854d0e",
        }}>
          正在连接语音服务...
        </div>
      ) : error && !isConnected ? (
        <div style={{
          padding: "8px 16px", background: "#fee2e2", borderRadius: 8, marginBottom: 16,
          fontSize: 13, color: "#991b1b",
        }}>
          语音连接失败: {error}
        </div>
      ) : null}

      {/* Room Top Bar */}
      <div style={{
        display: "flex", alignItems: "center", gap: 16, padding: "14px 20px",
        background: "#fff", borderRadius: 12, marginBottom: 16,
        border: "2px solid #e2e8f0", boxShadow: "0 2px 8px rgba(0,0,0,0.06)",
        flexWrap: "wrap",
      }}>
        <span style={{ fontSize: 13, fontFamily: "monospace", color: "#64748b", fontWeight: 600 }}>
          房间: {roomId?.slice(0, 8)}
        </span>
        <span style={{ fontSize: 13, color: "#cbd5e1", fontWeight: 300 }}>|</span>
        <Badge tone={isConnected ? "success" : "warning"}>
          {participants.length}/{roomSize}
        </Badge>
        <span style={{ fontSize: 13, color: "#cbd5e1", fontWeight: 300 }}>|</span>
        <span style={{ fontSize: 15, fontWeight: 700, fontVariantNumeric: "tabular-nums", color: "#334155" }}>
          {String(Math.floor(duration / 60)).padStart(2, "0")}:
          {String(duration % 60).padStart(2, "0")}
        </span>
        <div style={{ flex: 1 }} />
        <button
          onClick={toggleMute}
          style={{
            padding: "8px 16px", borderRadius: 8, border: `2px solid ${isMuted ? "#f59e0b" : "#e2e8f0"}`,
            background: isMuted ? "#fffbeb" : "#f8fafc",
            color: isMuted ? "#b45309" : "#475569",
            fontSize: 13, fontWeight: 700, cursor: "pointer",
          }}
        >
          {isMuted ? "取消静音" : "静音"}
        </button>
        <button
          onClick={handleLeave}
          style={{
            padding: "8px 16px", borderRadius: 8, border: "2px solid #fecaca",
            background: "#fef2f2", color: "#dc2626",
            fontSize: 13, fontWeight: 700, cursor: "pointer",
          }}
        >
          离开
        </button>
      </div>

      {/* Two-column layout */}
      <div style={{ display: "flex", gap: 16, flex: 1 }}>
        {/* Left: Participants */}
        <div style={{ width: 260, flexShrink: 0 }}>
          <Card title="参与者" subtitle={`${participants.length} 人在线`}>
            <div className="stack">
              {participants.map((p) => (
                <div key={p.sid} style={{
                  display: "flex", alignItems: "center", gap: 8, padding: "8px 0",
                  borderBottom: "1px solid rgba(15,23,42,0.06)",
                }}>
                  <span style={{
                    width: 8, height: 8, borderRadius: "50%",
                    background: p.isMuted ? "#f59e0b" : "#22c55e",
                    flexShrink: 0,
                  }} />
                  <span style={{ fontSize: 14, fontWeight: 500 }}>
                    {p.identity}
                    {p.isLocal ? " (你)" : ""}
                  </span>
                  {p.isMuted ? (
                    <span style={{ fontSize: 12, color: "#94a3b8", marginLeft: "auto" }}>已静音</span>
                  ) : null}
                </div>
              ))}
              {participants.length === 0 && connecting ? (
                <div className="emptyState">正在连接...</div>
              ) : participants.length === 0 ? (
                <div className="emptyState">等待参与者加入...</div>
              ) : null}
            </div>
          </Card>
        </div>

        {/* Right: Chat */}
        <div className="feed" style={{ flex: 1 }}>
          <Card title="房间消息" subtitle="消息仅当前房间可见">
            <div className="chatFeed" style={{ minHeight: 240 }}>
              {messages.map((m) => (
                <MessageBubble
                  key={m.id}
                  message={{
                    id: m.id,
                    text: `${m.sender}: ${m.text}`,
                    mine: m.sender === myIdentity.current,
                    time: m.time,
                  }}
                />
              ))}
              {messages.length === 0 ? (
                <div className="emptyState">发送第一条消息开始交流</div>
              ) : null}
            </div>
            <Composer placeholder="输入消息..." onSend={handleSend} />
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
