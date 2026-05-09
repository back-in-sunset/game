import {
  Room,
  RoomEvent,
  type ConnectionState,
} from "livekit-client";

export type LiveKitStatus = ConnectionState | "idle";

export class LiveKitService {
  private room: Room;

  constructor() {
    this.room = new Room();
  }

  get state(): LiveKitStatus {
    return this.room.state;
  }

  get isMuted(): boolean {
    return !this.room.localParticipant.isMicrophoneEnabled;
  }

  onStateChange(fn: (state: LiveKitStatus) => void): () => void {
    const handler = (state: ConnectionState) => fn(state);
    this.room.on(RoomEvent.ConnectionStateChanged, handler);
    return () => { this.room.off(RoomEvent.ConnectionStateChanged, handler); };
  }

  async connect(url: string, token: string): Promise<void> {
    await this.room.connect(url, token);
  }

  disconnect(): void {
    this.room.disconnect();
  }

  dispose(): void {
    this.room.disconnect();
    this.room.removeAllListeners();
  }

  async setMute(muted: boolean): Promise<void> {
    await this.room.localParticipant.setMicrophoneEnabled(!muted);
  }
}
