import { useEffect, useRef, useCallback } from "react";
import { LiveKitService, type LiveKitStatus } from "../services/livekit";
import { useVoiceStore } from "../store/voiceStore";

export function useLiveKit() {
  const svc = useRef<LiveKitService | null>(null);
  const {
    livekitUrl, livekitToken, callState, isMuted,
    setCallState, toggleMute,
  } = useVoiceStore();

  const connect = useCallback(async (url: string, token: string) => {
    if (!svc.current) {
      svc.current = new LiveKitService();
    }
    svc.current.onStateChange((state) => {
      if (state === "connected") setCallState("connected");
      else if (state === "disconnected") setCallState("ended");
    });
    await svc.current.connect(url, token);
  }, [setCallState]);

  const disconnect = useCallback(() => {
    svc.current?.disconnect();
  }, []);

  const toggleMic = useCallback(async () => {
    if (!svc.current) return;
    await svc.current.setMute(!svc.current.isMuted);
    toggleMute();
  }, [toggleMute]);

  useEffect(() => {
    if (callState === "ringing" && livekitUrl && livekitToken) {
      connect(livekitUrl, livekitToken);
    }
  }, [callState, livekitUrl, livekitToken, connect]);

  useEffect(() => {
    return () => {
      svc.current?.disconnect();
      svc.current = null;
    };
  }, []);

  return {
    status: svc.current?.state ?? "idle" as LiveKitStatus,
    isMuted,
    connect,
    disconnect,
    toggleMic,
  };
}
