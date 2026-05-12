import {
  encodeAuthFrame,
  encodeHeartbeat,
  encodeSendFrame,
  decodeFrame,
} from "@game/api";
import type { IMAuthRequest, IMClientStatus, IMPacket } from "@game/api";
import type { IMListener, IMStatusListener } from "./types";

const HEARTBEAT_INTERVAL = 25000;

const MAX_RECONNECT_ATTEMPTS = 10;
const BASE_RECONNECT_DELAY = 1000;
const MAX_RECONNECT_DELAY = 30000;

export class IMClient {
  private socket: WebSocket | null = null;
  private seq = 1;
  private auth: IMAuthRequest | null = null;
  private intentionalClose = false;
  private reconnectAttempts = 0;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null;
  private readonly listeners = new Set<IMListener>();
  private readonly statusListeners = new Set<IMStatusListener>();
  private currentStatus: IMClientStatus = "idle";

  constructor(private readonly url: string) {}

  get status(): IMClientStatus {
    return this.currentStatus;
  }

  async connect(auth: IMAuthRequest): Promise<void> {
    if (!this.url) throw new Error("IM websocket url is required");
    this.auth = auth;
    this.intentionalClose = false;
    await this.doConnect(auth);
  }

  disconnect(): void {
    this.intentionalClose = true;
    this.clearTimers();
    this.socket?.close();
    this.socket = null;
    this.auth = null;
    this.setStatus("idle");
  }

  onPacket(listener: IMListener): () => void {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  }

  onStatusChange(listener: IMStatusListener): () => void {
    this.statusListeners.add(listener);
    listener(this.currentStatus);
    return () => this.statusListeners.delete(listener);
  }

  send(action: string, data?: Record<string, unknown>): void {
    this.sendPacket(encodeSendFrame(this.seq++, { action, data }));
  }

  private sendPacket(frame: Uint8Array<ArrayBuffer>): void {
    if (!this.socket || this.socket.readyState !== WebSocket.OPEN) {
      throw new Error("IM websocket is not connected");
    }
    this.socket.send(frame);
  }

  private setStatus(status: IMClientStatus): void {
    this.currentStatus = status;
    this.statusListeners.forEach((fn) => fn(status));
  }

  private clearTimers(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private startHeartbeat(): void {
    this.stopHeartbeat();
    this.heartbeatTimer = setInterval(() => {
      if (this.socket?.readyState === WebSocket.OPEN) {
        this.socket.send(encodeHeartbeat(this.seq++));
      }
    }, HEARTBEAT_INTERVAL);
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer);
      this.heartbeatTimer = null;
    }
  }

  private async doConnect(auth: IMAuthRequest): Promise<void> {
    this.setStatus("connecting");

    this.socket = new WebSocket(this.url);
    await new Promise<void>((resolve, reject) => {
      if (!this.socket) {
        reject(new Error("socket not created"));
        return;
      }

      const onOpen = () => {
        this.socket?.removeEventListener("error", onError);
        this.socket?.addEventListener("message", this.onMessage);
        this.socket?.addEventListener("close", this.onClose);
        console.log("[IMClient] sending auth frame, url:", this.url, "token prefix:", auth.token.slice(0, 30) + "...", "token len:", auth.token.length);
        this.socket?.send(encodeAuthFrame(this.seq++, auth));
        this.setStatus("connected");
        this.reconnectAttempts = 0;
        this.startHeartbeat();
        resolve();
      };
      const onError = () => reject(new Error("failed to open IM websocket"));

      this.socket.addEventListener("open", onOpen, { once: true });
      this.socket.addEventListener("error", onError, { once: true });
    });
  }

  private onMessage = async (event: MessageEvent): Promise<void> => {
    if (typeof event.data === "string") return;
    const buffer = await event.data.arrayBuffer();
    const frame = decodeFrame(buffer);
    console.log("[IMClient] received op:", frame.op, "body:", frame.body.slice(0, 200));

    // heartbeat reply — not forwarded to listeners
    if (frame.op === 3) return;

    this.listeners.forEach((listener) => listener(frame));
  };

  private onClose = (): void => {
    this.stopHeartbeat();
    this.socket = null;

    if (this.intentionalClose) return;

    if (this.reconnectAttempts >= MAX_RECONNECT_ATTEMPTS) {
      this.setStatus("error");
      return;
    }

    this.setStatus("reconnecting");

    const delay = Math.min(
      BASE_RECONNECT_DELAY * Math.pow(2, this.reconnectAttempts),
      MAX_RECONNECT_DELAY,
    );
    this.reconnectAttempts++;

    this.reconnectTimer = setTimeout(() => {
      if (this.auth) {
        this.doConnect(this.auth).catch(() => {
          this.onClose();
        });
      }
    }, delay);
  };
}

export function createIMClient(url: string): IMClient {
  return new IMClient(url);
}
