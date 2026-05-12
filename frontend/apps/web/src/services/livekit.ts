import {
  Room,
  RoomEvent,
  type ConnectionState,
  type RemoteParticipant,
} from "livekit-client";

export type LiveKitStatus = ConnectionState | "idle";

export type ParticipantInfo = {
  sid: string;
  identity: string;
  isMuted: boolean;
  isLocal: boolean;
};

export class LiveKitService {
  private room: Room;
  private _participants: Map<string, ParticipantInfo> = new Map();
  private _onParticipantsChange: ((participants: ParticipantInfo[]) => void) | null = null;

  constructor() {
    this.room = new Room();
    this.setupParticipantEvents();
  }

  private setupParticipantEvents(): void {
    this.room.on(RoomEvent.ParticipantConnected, (p: RemoteParticipant) => {
      this._participants.set(p.sid, {
        sid: p.sid,
        identity: p.identity || p.sid.slice(0, 8),
        isMuted: !p.isMicrophoneEnabled,
        isLocal: false,
      });
      this.notifyParticipants();
    });

    this.room.on(RoomEvent.ParticipantDisconnected, (p: RemoteParticipant) => {
      this._participants.delete(p.sid);
      this.notifyParticipants();
    });

    this.room.on(RoomEvent.LocalTrackPublished, () => {
      const local = this.room.localParticipant;
      this._participants.set(local.sid, {
        sid: local.sid,
        identity: local.identity || "You",
        isMuted: !local.isMicrophoneEnabled,
        isLocal: true,
      });
      this.notifyParticipants();
    });

    this.room.on(RoomEvent.TrackMuted, (_pub, participant) => {
      const existing = this._participants.get(participant.sid);
      if (existing) {
        this._participants.set(participant.sid, { ...existing, isMuted: true });
        this.notifyParticipants();
      }
    });

    this.room.on(RoomEvent.TrackUnmuted, (_pub, participant) => {
      const existing = this._participants.get(participant.sid);
      if (existing) {
        this._participants.set(participant.sid, { ...existing, isMuted: false });
        this.notifyParticipants();
      }
    });
  }

  private notifyParticipants(): void {
    this._onParticipantsChange?.(this.participants);
  }

  get state(): LiveKitStatus {
    return this.room.state;
  }

  get isMuted(): boolean {
    return !this.room.localParticipant.isMicrophoneEnabled;
  }

  get participants(): ParticipantInfo[] {
    return Array.from(this._participants.values());
  }

  get participantCount(): number {
    return this._participants.size;
  }

  onStateChange(fn: (state: LiveKitStatus) => void): () => void {
    const handler = (state: ConnectionState) => fn(state);
    this.room.on(RoomEvent.ConnectionStateChanged, handler);
    return () => { this.room.off(RoomEvent.ConnectionStateChanged, handler); };
  }

  onParticipantsChange(fn: (participants: ParticipantInfo[]) => void): () => void {
    this._onParticipantsChange = fn;
    return () => { this._onParticipantsChange = null; };
  }

  async connect(url: string, token: string): Promise<void> {
    await this.room.connect(url, token);
    // Register local participant
    const local = this.room.localParticipant;
    this._participants.set(local.sid, {
      sid: local.sid,
      identity: local.identity || "You",
      isMuted: !local.isMicrophoneEnabled,
      isLocal: true,
    });
    // Register existing remote participants (ParticipantConnected only fires for new joins)
    this.room.remoteParticipants.forEach((p) => {
      this._participants.set(p.sid, {
        sid: p.sid,
        identity: p.identity || p.sid.slice(0, 8),
        isMuted: !p.isMicrophoneEnabled,
        isLocal: false,
      });
    });
    this.notifyParticipants();
  }

  disconnect(): void {
    this._participants.clear();
    this.room.disconnect();
  }

  dispose(): void {
    this._participants.clear();
    this.room.disconnect();
    this.room.removeAllListeners();
  }

  async setMute(muted: boolean): Promise<void> {
    await this.room.localParticipant.setMicrophoneEnabled(!muted);
  }
}
