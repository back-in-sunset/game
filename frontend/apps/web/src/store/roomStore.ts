import { create } from "zustand";
import { LiveKitService, type ParticipantInfo } from "../services/livekit";

type RoomState = {
  roomId: string | null;
  roomSize: number;
  participants: ParticipantInfo[];
  duration: number;
  error: string;
  lk: LiveKitService | null;

  joinRoom: (roomId: string, roomSize: number, token: string, url: string) => Promise<void>;
  leaveRoom: () => void;
  toggleMute: () => Promise<void>;
  tickDuration: () => void;
  setError: (msg: string) => void;
};

export const useRoomStore = create<RoomState>((set, get) => ({
  roomId: null,
  roomSize: 0,
  participants: [],
  duration: 0,
  error: "",
  lk: null,

  joinRoom: async (roomId, roomSize, token, url) => {
    const lk = new LiveKitService();

    lk.onParticipantsChange((participants) => {
      set({ participants });
    });

    lk.onStateChange((state) => {
      if (state === "disconnected") {
        set({ error: "房间连接已断开" });
      }
    });

    try {
      await lk.connect(url, token);
      set({ roomId, roomSize, lk, participants: lk.participants, duration: 0, error: "" });
    } catch (e) {
      lk.dispose();
      throw e;
    }
  },

  leaveRoom: () => {
    const { lk } = get();
    if (lk) {
      lk.dispose();
    }
    set({ roomId: null, roomSize: 0, participants: [], duration: 0, lk: null, error: "" });
  },

  toggleMute: async () => {
    const { lk } = get();
    if (!lk) return;
    const newMuted = !lk.isMuted;
    await lk.setMute(newMuted);
  },

  tickDuration: () => set((s) => ({ duration: s.duration + 1 })),

  setError: (msg) => set({ error: msg }),
}));
