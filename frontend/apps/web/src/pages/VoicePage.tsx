import { useEffect, useRef } from "react";
import { Link } from "react-router-dom";
import { Card, Button, Badge } from "@game/ui";
import { useVoiceStore } from "../store/voiceStore";
import { useLiveKit } from "../hooks/useLiveKit";

export function VoicePage() {
  const callState = useVoiceStore((s) => s.callState);
  const callId = useVoiceStore((s) => s.callId);
  const duration = useVoiceStore((s) => s.duration);
  const isMuted = useVoiceStore((s) => s.isMuted);
  const toggleMute = useVoiceStore((s) => s.toggleMute);
  const resetCall = useVoiceStore((s) => s.resetCall);
  const tickDuration = useVoiceStore((s) => s.tickDuration);
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
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">V</div>
          <div><h1>语音通话</h1></div>
        </div>

        <Card title="通话状态">
          <div className="stack">
            <Badge tone={tone}>{callLabel}</Badge>
            {callId ? (
              <p style={{ fontSize: 13, color: "#64748b" }}>
                Call: {callId.slice(0, 12)}...
              </p>
            ) : null}
            {callState === "connected" ? (
              <p style={{ fontSize: 14, fontWeight: 600, color: "#166534" }}>
                {String(Math.floor(duration / 60)).padStart(2, "0")}:
                {String(duration % 60).padStart(2, "0")}
              </p>
            ) : null}
            <Badge tone={lk.status === "connected" ? "success" : "neutral"}>
              LiveKit: {lk.status}
            </Badge>
          </div>
        </Card>

        <Link to="/">← 返回</Link>
      </aside>

      <main className="main">
        <Card title="通话控制">
          <div className="buttonGrid">
            {callState !== "connected" ? (
              <Button variant="primary" onClick={() => {
                // manual connect — in real flow this comes from call_invite response
              }}>
                发起呼叫
              </Button>
            ) : null}
            <Button variant={isMuted ? "ghost" : "secondary"} onClick={() => lk.toggleMic()}>
              {isMuted ? "取消静音" : "静音"}
            </Button>
            {callState === "connected" || callState === "ringing" ? (
              <Button variant="ghost" onClick={() => { lk.disconnect(); resetCall(); }}>
                挂断
              </Button>
            ) : null}
          </div>
        </Card>

        {callState === "idle" ? (
          <Card title="使用说明" subtitle="语音通话基于 LiveKit WebRTC">
            <p style={{ fontSize: 13, color: "#64748b" }}>
              从仪表盘发起通话邀请，对方接听后自动建立 WebRTC 连接。通话中可开关麦克风。
            </p>
          </Card>
        ) : null}
      </main>
    </div>
  );
}
