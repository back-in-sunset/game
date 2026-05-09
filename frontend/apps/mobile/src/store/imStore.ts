import { create } from "zustand";
import { createIMClient, type IMClient, type IMClientStatus } from "../services/im";

type Message = { id: string; text: string; mine: boolean; time: string };

type IMState = {
  client: IMClient | null;
  status: IMClientStatus;
  error: string;
  messages: Message[];

  connect: (url: string, auth: {
    token: string; domain: "platform" | "tenant";
    scope: { tenant_id: string; project_id: string; environment: string };
  }) => Promise<void>;
  disconnect: () => void;
  sendMessage: (receiver: number, text: string) => void;
};

export const useIMStore = create<IMState>((set, get) => ({
  client: null,
  status: "idle",
  error: "",
  messages: [],

  connect: async (url, auth) => {
    const client = createIMClient(url);
    client.onStatusChange((status) =>
      set({ status, error: status === "error" ? "connection lost" : "" }),
    );
    client.onPacket((packet) => {
      if (packet.op === 4 || packet.op === 5 || packet.op === 6) {
        try {
          JSON.parse(packet.body);
          set((s) => ({
            messages: [...s.messages, {
              id: `srv-${Date.now()}`,
              text: packet.body.slice(0, 200),
              mine: false,
              time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
            }],
          }));
        } catch { /* ignore */ }
      }
    });
    try {
      await client.connect(auth);
      set({ client, error: "" });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : "failed" });
      throw err;
    }
  },

  disconnect: () => {
    get().client?.disconnect();
    set({ client: null, status: "idle" });
  },

  sendMessage: (receiver, text) => {
    get().client?.send("send", { receiver, msg_type: "direct_message", payload: { text } });
    set((s) => ({
      messages: [...s.messages, {
        id: `me-${Date.now()}`, text, mine: true,
        time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
      }],
    }));
  },
}));
