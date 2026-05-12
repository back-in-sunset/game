import { create } from "zustand";
import {
  createIMClient,
  type IMAuthRequest,
  type IMClient,
  type IMClientStatus,
} from "../services/im";
import type { Conversation, Message } from "@game/shared";
import { useVoiceStore } from "./voiceStore";

type IMState = {
  client: IMClient | null;
  status: IMClientStatus;
  error: string;
  conversations: Conversation[];
  messages: Message[];
  activeConversationId: string | null;

  connect: (url: string, auth: IMAuthRequest) => Promise<void>;
  disconnect: () => void;
  loadConversations: (convs: Conversation[]) => void;
  setActiveConversation: (id: string) => void;
  sendMessage: (receiver: number, text: string) => void;
  sendVoiceAction: (action: string, data: Record<string, unknown>) => void;
  addMessage: (msg: Message) => void;
  setStatus: (status: IMClientStatus, error?: string) => void;
};

export const useIMStore = create<IMState>((set, get) => ({
  client: null,
  status: "idle",
  error: "",
  conversations: [],
  messages: [],
  activeConversationId: null,

  loadConversations: (convs) => set({ conversations: convs }),

  connect: async (url, auth: IMAuthRequest) => {
    console.log("[imStore] connect called, url:", url, "token len:", auth.token.length, "token prefix:", auth.token.slice(0, 20) + "...");
    const client = createIMClient(url);

    client.onStatusChange((status) =>
      set({ status, error: status === "error" ? "connection lost" : "" }),
    );

    client.onPacket((packet) => {
      if (packet.op === 4) {
        try {
          const data = JSON.parse(packet.body);

          // voice call events
          if (data.type === "call_invite" && data.call_id && data.livekit_url && data.livekit_token) {
            useVoiceStore.getState().setCall({
              callId: data.call_id,
              token: data.livekit_token,
              room: data.livekit_room ?? "",
              url: data.livekit_url,
            });
            return;
          }

          if (data.type === "call_end") {
            useVoiceStore.getState().resetCall();
            return;
          }

          // presence update
          if (data.type === "presence") {
            set((s) => ({
              conversations: s.conversations.map((c) =>
                c.userId === data.user_id ? { ...c, online: Boolean(data.online) } : c,
              ),
            }));
            return;
          }

          // server-pushed message (envelope format from router)
          if (data.type === "message" && data.envelope) {
            const env = data.envelope;
            const sender = env.sender;
            const text = env.payload?.text ?? JSON.stringify(env.payload ?? {});
            const msg: Message = {
              id: `push-${Date.now()}`,
              text,
              mine: false,
              time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
            };
            set((s) => {
              const convId = s.conversations.find((c) => c.userId === sender)?.id;
              const isActive = convId === s.activeConversationId;
              return {
                conversations: isActive
                  ? s.conversations
                  : s.conversations.map((c) =>
                      c.id === convId ? { ...c, unread: c.unread + 1 } : c,
                    ),
                messages: isActive ? [...s.messages, msg] : s.messages,
              };
            });
            return;
          }

          // legacy flat format
          const msg: Message = {
            id: `push-${Date.now()}`,
            text: data.payload?.text ?? packet.body.slice(0, 200),
            mine: false,
            time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
          };
          set((s) => {
            const convId = s.conversations.find((c) => c.userId === data.sender)?.id;
            const isActive = convId === s.activeConversationId;
            return {
              conversations: isActive
                ? s.conversations
                : s.conversations.map((c) =>
                    c.id === convId ? { ...c, unread: c.unread + 1 } : c,
                  ),
              messages: isActive ? [...s.messages, msg] : s.messages,
            };
          });
        } catch {
          // ignore non-JSON frames
        }
      }

      if (packet.op === 5 || packet.op === 6) {
        try {
          const data = JSON.parse(packet.body);
          // send_ack — silently update seq tracking, don't spam chat
          if (data.type === "send_ack") return;
          set((s) => ({
            messages: [
              ...s.messages,
              {
                id: `srv-${Date.now()}`,
                text: `收到服务端响应: ${JSON.stringify(data).slice(0, 100)}`,
                mine: false,
                time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
              },
            ],
          }));
        } catch {
          // ignore non-JSON frames
        }
      }
    });

    try {
      await client.connect(auth);
      set({ client, error: "" });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : "failed to connect" });
      throw err;
    }
  },

  disconnect: () => {
    get().client?.disconnect();
    set({ client: null, status: "idle" });
  },

  setActiveConversation: (id) => {
    set((s) => {
      const convs = s.conversations.map((c) =>
        c.id === id ? { ...c, unread: 0 } : c,
      );
      const conv = s.conversations.find((c) => c.id === id);
      return {
        activeConversationId: id,
        conversations: convs,
        messages: conv?.messages ?? [],
      };
    });
  },

  sendMessage: (receiver, text) => {
    get().client?.send("send", {
      receiver,
      msg_type: "direct_message",
      payload: { text },
    });
    const msg: Message = {
      id: `me-${Date.now()}`,
      text,
      mine: true,
      time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
    };
    set((s) => ({ messages: [...s.messages, msg] }));
  },

  sendVoiceAction: (action, data) => {
    get().client?.send(action, data);
  },

  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),

  setStatus: (status, error) => set({ status, error: error ?? "" }),
}));
