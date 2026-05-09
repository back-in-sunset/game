import { useState, useRef, useEffect } from "react";
import { View, Text, TouchableOpacity, StyleSheet } from "react-native";

type Props = { onBack: () => void };

export function VoiceScreen({ onBack }: Props) {
  const [isMuted, setIsMuted] = useState(false);
  const [duration, setDuration] = useState(0);
  const [callState, setCallState] = useState<"idle" | "ringing" | "connected" | "ended">("idle");
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  useEffect(() => {
    if (callState === "connected") {
      timerRef.current = setInterval(() => setDuration((d) => d + 1), 1000);
    }
    return () => { if (timerRef.current) clearInterval(timerRef.current); };
  }, [callState]);

  function startCall() { setCallState("connected"); setDuration(0); }
  function endCall() { setCallState("ended"); }

  const statusLabel =
    callState === "idle" ? "准备就绪" : callState === "ringing" ? "呼叫中..."
    : callState === "connected" ? "通话中" : "已结束";

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={onBack}>
          <Text style={styles.backBtn}>← 返回</Text>
        </TouchableOpacity>
        <Text style={styles.title}>语音通话</Text>
      </View>

      <View style={styles.body}>
        <Text style={styles.status}>{statusLabel}</Text>
        {callState === "connected" ? (
          <Text style={styles.duration}>
            {String(Math.floor(duration / 60)).padStart(2, "0")}:{String(duration % 60).padStart(2, "0")}
          </Text>
        ) : null}

        <View style={styles.controls}>
          {callState === "idle" || callState === "ended" ? (
            <TouchableOpacity style={[styles.btnBase, { backgroundColor: "#2f6bff" }]} onPress={startCall}>
              <Text style={styles.btnText}>开始通话</Text>
            </TouchableOpacity>
          ) : null}
          {callState === "connected" ? (
            <>
              <TouchableOpacity style={[styles.btnBase, { backgroundColor: isMuted ? "#94a3b8" : "#f59e0b" }]} onPress={() => setIsMuted(!isMuted)}>
                <Text style={styles.btnText}>{isMuted ? "取消静音" : "静音"}</Text>
              </TouchableOpacity>
              <TouchableOpacity style={[styles.btnBase, { backgroundColor: "#dc2626" }]} onPress={endCall}>
                <Text style={styles.btnText}>挂断</Text>
              </TouchableOpacity>
            </>
          ) : null}
        </View>
        <Text style={styles.note}>LiveKit WebRTC 将在后续版本接入</Text>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: "#f8fafc" },
  header: { flexDirection: "row", alignItems: "center", padding: 16, paddingTop: 60, backgroundColor: "#fff", borderBottomWidth: 1, borderColor: "#e2e8f0" },
  backBtn: { fontSize: 15, color: "#2f6bff", marginRight: 12 },
  title: { fontSize: 18, fontWeight: "600", color: "#0f172a" },
  body: { flex: 1, justifyContent: "center", alignItems: "center", padding: 20 },
  status: { fontSize: 20, fontWeight: "600", color: "#0f172a", marginBottom: 8 },
  duration: { fontSize: 36, fontWeight: "700", color: "#16a34a", marginBottom: 32, fontVariant: ["tabular-nums"] },
  controls: { flexDirection: "row", gap: 12, flexWrap: "wrap", justifyContent: "center" },
  btnBase: {
    paddingHorizontal: 24, paddingVertical: 14, borderRadius: 12,
  },
  btnText: { color: "#fff", fontSize: 16, fontWeight: "600" },
  note: { marginTop: 32, fontSize: 13, color: "#94a3b8" },
});
