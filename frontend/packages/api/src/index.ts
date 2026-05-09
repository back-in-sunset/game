export type IMDomain = "platform" | "tenant";

export type IMScope = {
  tenant_id: string;
  project_id: string;
  environment: string;
};

export type IMAuthRequest = {
  token: string;
  domain: IMDomain;
  scope: IMScope;
};

export type IMAuthReply = {
  type?: "login_ok";
  user_id?: number;
  domain?: IMDomain;
  scope?: IMScope;
};

export type IMSendAck = {
  type?: "send_ack";
  seq?: number;
  online_recipients?: number;
  stored_offline?: boolean;
};

export type IMVoiceAck = {
  type?: string;
  call_id?: string;
  state?: string;
  livekit_url?: string;
  livekit_token?: string;
  livekit_room?: string;
  room_id?: string;
  participants?: number[];
  muted?: boolean;
  push_targets?: number[];
};

export type IMConversation = {
  id: number;
  title: string;
  preview: string;
  online: boolean;
};

export type IMMessage = {
  id: string;
  conversationId: number;
  text: string;
  mine: boolean;
  time: string;
};

export type IMEnvelope = {
  action: string;
  data?: Record<string, unknown>;
};

export type IMClientStatus = "idle" | "connecting" | "connected" | "reconnecting" | "error";

export type IMPacket = {
  op: number;
  body: string;
};

export {
  HttpClient,
  HttpError,
  createUserAPI,
  createFriendAPI,
  createCommentAPI,
  createHistoryAPI,
  createPlatformAPI,
} from "./http";
export type {
  UserProfile,
  UserAPI,
  FriendItem,
  FriendAPI,
  CommentItem,
  CommentAPI,
  HistoryItem,
  HistoryAPI,
  PlatformConfig,
  PlatformAPI,
} from "./http";

// IM binary frame codec — 16-byte header + JSON payload
const OP_AUTH = 7;
const OP_SEND = 4;
const OP_HEARTBEAT = 2;

export function encodeFrame(op: number, seq: number, body: unknown): Uint8Array<ArrayBuffer> {
  const payload = new TextEncoder().encode(JSON.stringify(body));
  const frame = new Uint8Array(16 + payload.byteLength) as Uint8Array<ArrayBuffer>;
  const view = new DataView(frame.buffer);
  view.setUint32(0, frame.byteLength);
  view.setUint16(4, 16);
  view.setUint16(6, 1);
  view.setUint32(8, op);
  view.setUint32(12, seq);
  frame.set(payload, 16);
  return frame;
}

export function encodeAuthFrame(seq: number, body: unknown): Uint8Array<ArrayBuffer> {
  return encodeFrame(OP_AUTH, seq, body);
}

export function encodeSendFrame(seq: number, body: unknown): Uint8Array<ArrayBuffer> {
  return encodeFrame(OP_SEND, seq, body);
}

export function encodeHeartbeat(seq: number): Uint8Array<ArrayBuffer> {
  const frame = new Uint8Array(16) as Uint8Array<ArrayBuffer>;
  const view = new DataView(frame.buffer);
  view.setUint32(0, 16);
  view.setUint16(4, 16);
  view.setUint16(6, 1);
  view.setUint32(8, OP_HEARTBEAT);
  view.setUint32(12, seq);
  return frame;
}

export function decodeFrame(buffer: ArrayBuffer): IMPacket {
  const view = new DataView(buffer);
  return {
    op: view.getUint32(8),
    body: new TextDecoder().decode(buffer.slice(16)),
  };
}

