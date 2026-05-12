import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { Card, Button, Badge } from "@game/ui";
import { useVoiceStore } from "../store/voiceStore";
import { useIMStore } from "../store/imStore";
import { useSettingsStore } from "../store/settingsStore";
import { useLiveKit } from "../hooks/useLiveKit";
import { WebShell } from "../components/WebShell";

export function VoicePage() {
  const callState = useVoiceStore((s) => s.callState);
  const callId = useVoiceStore((s) => s.callId);
  const duration = useVoiceStore((s) => s.duration);
  const isMuted = useVoiceStore((s) => s.isMuted);
  const setCall = useVoiceStore((s) => s.setCall);
  const resetCall = useVoiceStore((s) => s.resetCall);
  const tickDuration = useVoiceStore((s) => s.tickDuration);
  const sendVoiceAction = useIMStore((s) => s.sendVoiceAction);
  const livekitUrl = useSettingsStore((s) => s.livekitUrl);
  const lk = useLiveKit();
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (callState === "connected") {
      timerRef.current = setInterval(tickDuration, 1000);
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [callState, tickDuration]);

  const callLabel =
    callState === "idle" ? "未通话"
    : callState === "ringing" ? "呼叫中..."
    : callState === "connected" ? "通话中"
    : "已结束";

  const tone =
    callState === "connected" ? "success"
    : callState === "ringing" ? "warning"
    : "neutral";

  return (
    <WebShell
      title="语音面板"
      subtitle="通话状态 / LiveKit / 麦克风控制"
      activeTab="chat"
      rightRail={
        <div className="stack">
          <div className="sectionHeader">
            <div className="titleBlock">
              <h3>通话状态</h3>
              <p>基于 LiveKit WebRTC</p>
            </div>
          </div>
          <Badge tone={tone}>{callLabel}</Badge>
          {callId ? <p className="mutedText">Call: {callId.slice(0, 12)}...</p> : null}
          {callState === "connected" ? (
            <p style={{ fontSize: 14, fontWeight: 700, color: "#166534", margin: 0 }}>
              {String(Math.floor(duration / 60)).padStart(2, "0")}:
              {String(duration % 60).padStart(2, "0")}
            </p>
          ) : null}
          <Badge tone={lk.status === "connected" ? "success" : "neutral"}>
            LiveKit: {lk.status}
          </Badge>
        </div>
      }
    >
      <section className="feed">
        <Card title="通话控制">
          <div className="stack">
            <div className="buttonGrid">
              {callState !== "connected" ? (
                <Button variant="primary" onClick={() => {
                  const id = `call-${Date.now()}`;
                  const room = `room-${Date.now()}`;
                  const token = "dev-token";
                  setCall({ callId: id, token, room, url: livekitUrl });
                  sendVoiceAction("call_invite", { call_id: id, livekit_url: livekitUrl, livekit_token: token, livekit_room: room });
                }}>
                  发起呼叫
                </Button>
              ) : null}
              <Button variant={isMuted ? "ghost" : "secondary"} onClick={() => lk.toggleMic()}>
                {isMuted ? "取消静音" : "静音"}
              </Button>
              {callState === "connected" || callState === "ringing" ? (
                <Button variant="ghost" onClick={() => { sendVoiceAction("call_end", { call_id: callId }); lk.disconnect(); resetCall(); }}>
                  挂断
                </Button>
              ) : null}
            </div>
          </div>
        </Card>

        {callState === "idle" ? (
          <Card title="使用说明" subtitle="语音通话基于 LiveKit WebRTC">
            <p style={{ fontSize: 13, color: "#667085" }}>
              从广场或聊天框进入后发起通话邀请，对方接听后自动建立 WebRTC 连接。通话中可开关麦克风。
            </p>
            <Link to="/" className="linkRow">
              <span>◎</span>
              <span className="channelMeta">
                <strong>返回广场</strong>
                <span>继续浏览帖子</span>
              </span>
            </Link>
          </Card>
        ) : null}
      </section>
    </WebShell>
  );
}
