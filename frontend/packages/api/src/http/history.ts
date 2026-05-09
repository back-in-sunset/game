import type { HttpClient } from "./client";

export type HistoryItem = {
  id: string;
  peerUserId: number;
  msgType: string;
  payload: Record<string, unknown>;
  sentAt: string;
};

export function createHistoryAPI(client: HttpClient) {
  return {
    list: (peerUserId: number, limit = 20) =>
      client.get<HistoryItem[]>("/history/list", { peer_user_id: String(peerUserId), limit: String(limit) }),
  };
}

export type HistoryAPI = ReturnType<typeof createHistoryAPI>;
