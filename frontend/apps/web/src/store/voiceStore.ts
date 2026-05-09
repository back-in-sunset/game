import { create } from "zustand";

type CallState = "idle" | "ringing" | "connected" | "ended";

type VoiceState = {
  callState: CallState;
  callId: string | null;
  livekitToken: string | null;
  livekitRoom: string | null;
  livekitUrl: string | null;
  isMuted: boolean;
  duration: number;

  setCall: (call: { callId: string; token: string; room: string; url: string }) => void;
  setCallState: (state: CallState) => void;
  toggleMute: () => void;
  resetCall: () => void;
  tickDuration: () => void;
};

export const useVoiceStore = create<VoiceState>((set) => ({
  callState: "idle",
  callId: null,
  livekitToken: null,
  livekitRoom: null,
  livekitUrl: null,
  isMuted: false,
  duration: 0,

  setCall: ({ callId, token, room, url }) =>
    set({ callId, livekitToken: token, livekitRoom: room, livekitUrl: url, callState: "ringing" }),

  setCallState: (callState) => set({ callState }),

  toggleMute: () => set((s) => ({ isMuted: !s.isMuted })),

  resetCall: () =>
    set({
      callState: "idle",
      callId: null,
      livekitToken: null,
      livekitRoom: null,
      livekitUrl: null,
      isMuted: false,
      duration: 0,
    }),

  tickDuration: () => set((s) => ({ duration: s.duration + 1 })),
}));
