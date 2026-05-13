import { create } from "zustand";
import {
  createIMClient,
  type IMAuthRequest,
  type IMClient,
  type IMClientStatus,
} from "../services/im";
import type { Conversation, Message } from "@game/shared";
import { useVoiceStore } from "./voiceStore";

export type RoomMessageData = {
  id: string;
  text: string;
  senderIdentity: string;
  displayName: string;
  roomId: string;
  time: string;
};

type IMState = {
  client: IMClient | null;
  status: IMClientStatus;
  error: string;
  conversations: Conversation[];
  messages: Message[];
  activeConversationId: string | null;
  onRoomMessage: ((msg: RoomMessageData) => void) | null;
  onRoomHistory: ((messages: RoomMessageData[]) => void) | null;

  connect: (url: string, auth: IMAuthRequest) => Promise<void>;
  disconnect: () => void;
  loadConversations: (convs: Conversation[]) => void;
  setActiveConversation: (id: string) => void;
  sendMessage: (receiver: number, text: string) => void;
  sendVoiceAction: (action: string, data: Record<string, unknown>) => void;
  sendRoomMessage: (roomId: string, text: string, identity: string, displayName: string) => void;
  loadRoomHistory: (roomId: string) => Promise<void>;
  setOnRoomMessage: (fn: ((msg: RoomMessageData) => void) | null) => void;
  setOnRoomHistory: (fn: ((msgs: RoomMessageData[]) => void) | null) => void;
  addMessage: (msg: Message) => void;
  setStatus: (status: IMClientStatus, error?: string) => void;
};

function makeRoomMessage(env: Record<string, unknown>): RoomMessageData | null {
  const payload = env.payload as Record<string, unknown> | undefined;
  if (!payload?.room_id || !payload?.text) return null;
  return {
    id: `room-${Date.now()}-${Math.random().toString(36).slice(2, 6)}`,
    text: String(payload.text),
    senderIdentity: String(payload.sender_identity ?? ""),
    displayName: String(payload.display_name ?? payload.sender_identity ?? ""),
    roomId: String(payload.room_id),
    time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
  };
}

export const useIMStore = create<IMState>((set, get) => ({
  client: null,
  status: "idle",
  error: "",
  conversations: [],
  messages: [],
  activeConversationId: null,
  onRoomMessage: null,
  onRoomHistory: null,

  loadConversations: (convs) => set({ conversations: convs }),

  connect: async (url, auth: IMAuthRequest) => {
    console.log("[imStore] connect called, url:", url, "token len:", auth.token.length);
    const client = createIMClient(url);
    set({ client }); // set early so loadRoomHistory can use it when status flips to "connected"

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

          // offline batch — drain on login, route room messages
          if (data.type === "offline_batch" && Array.isArray(data.messages)) {
            const roomMsgs: RoomMessageData[] = [];
            (data.messages as Record<string, unknown>[]).forEach((env) => {
              const rm = makeRoomMessage(env);
              if (rm) roomMsgs.push(rm);
            });
            if (roomMsgs.length > 0) {
              const onRoomHistory = get().onRoomHistory;
              if (onRoomHistory) onRoomHistory(roomMsgs);
            }
            return;
          }

          // server-pushed message (envelope format from router)
          if (data.type === "message" && data.envelope) {
            const env = data.envelope as Record<string, unknown>;

            // room message — route to room callback instead of conversation
            if (env.msg_type === "room_message") {
              const rm = makeRoomMessage(env);
              if (rm) get().onRoomMessage?.(rm);
              return;
            }

            // direct message
            const sender = env.sender as number;
            const text = (env.payload as Record<string, unknown>)?.text as string ?? JSON.stringify(env.payload ?? {});
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
            text: (data.payload as Record<string, unknown>)?.text as string ?? packet.body.slice(0, 200),
            mine: false,
            time: new Date().toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" }),
          };
          set((s) => {
            const convId = s.conversations.find((c) => c.userId === (data.sender as number))?.id;
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
      set({ error: "" });
    } catch (err) {
      // Only clear client if it's still the one we created (Strict Mode may have replaced it)
      if (get().client === client) {
        set({ client: null, error: err instanceof Error ? err.message : "failed to connect" });
      }
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

  sendRoomMessage: (roomId, text, identity, displayName) => {
    get().client?.send("send", {
      receiver: 1,
      msg_type: "room_message",
      payload: {
        text,
        room_id: roomId,
        sender_identity: identity,
        display_name: displayName,
      },
    });
  },

  loadRoomHistory: async (roomId) => {
    const client = get().client;
    if (!client) return;
    try {
      const data = await client.request(
        "list_messages",
        { peer_user_id: 1, limit: 100 },
        "message_list",
      );
      const messages = (data.messages as Record<string, unknown>[]) ?? [];
      const roomMsgs: RoomMessageData[] = [];
      messages.forEach((env) => {
        if (env.msg_type === "room_message") {
          const rm = makeRoomMessage(env);
          if (rm && rm.roomId === roomId) roomMsgs.push(rm);
        }
      });
      const onRoomHistory = get().onRoomHistory;
      if (onRoomHistory) onRoomHistory(roomMsgs);
    } catch (e) {
      console.warn("[imStore] loadRoomHistory failed:", e);
    }
  },

  setOnRoomMessage: (fn) => set({ onRoomMessage: fn }),

  setOnRoomHistory: (fn) => set({ onRoomHistory: fn }),

  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),

  setStatus: (status, error) => set({ status, error: error ?? "" }),
}));
