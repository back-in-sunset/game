import { describe, it, expect, beforeEach } from "vitest";
import { useVoiceStore } from "../voiceStore";

describe("voiceStore", () => {
  beforeEach(() => {
    useVoiceStore.setState({
      callState: "idle",
      callId: null,
      livekitToken: null,
      livekitRoom: null,
      livekitUrl: null,
      isMuted: false,
      duration: 0,
    });
  });

  describe("setCall", () => {
    it("stores call details and transitions to ringing", () => {
      useVoiceStore.getState().setCall({
        callId: "call-1",
        token: "lk-token",
        room: "room-a",
        url: "wss://livekit.example.com",
      });

      const s = useVoiceStore.getState();
      expect(s.callId).toBe("call-1");
      expect(s.livekitToken).toBe("lk-token");
      expect(s.livekitRoom).toBe("room-a");
      expect(s.livekitUrl).toBe("wss://livekit.example.com");
      expect(s.callState).toBe("ringing");
    });
  });

  describe("toggleMute", () => {
    it("toggles isMuted", () => {
      expect(useVoiceStore.getState().isMuted).toBe(false);
      useVoiceStore.getState().toggleMute();
      expect(useVoiceStore.getState().isMuted).toBe(true);
      useVoiceStore.getState().toggleMute();
      expect(useVoiceStore.getState().isMuted).toBe(false);
    });
  });

  describe("tickDuration", () => {
    it("increments duration by 1 each call", () => {
      expect(useVoiceStore.getState().duration).toBe(0);
      useVoiceStore.getState().tickDuration();
      useVoiceStore.getState().tickDuration();
      expect(useVoiceStore.getState().duration).toBe(2);
    });
  });

  describe("resetCall", () => {
    it("resets all fields to initial state", () => {
      useVoiceStore.getState().setCall({ callId: "c", token: "t", room: "r", url: "u" });
      useVoiceStore.getState().toggleMute();
      useVoiceStore.getState().tickDuration();

      useVoiceStore.getState().resetCall();

      const s = useVoiceStore.getState();
      expect(s.callState).toBe("idle");
      expect(s.callId).toBeNull();
      expect(s.livekitToken).toBeNull();
      expect(s.isMuted).toBe(false);
      expect(s.duration).toBe(0);
    });
  });

  describe("setCallState", () => {
    it("sets call state directly", () => {
      useVoiceStore.getState().setCallState("connected");
      expect(useVoiceStore.getState().callState).toBe("connected");
      useVoiceStore.getState().setCallState("ended");
      expect(useVoiceStore.getState().callState).toBe("ended");
    });
  });
});
