import type { HttpClient } from "./client";

export type FriendItem = {
  userId: number;
  nickname: string;
  avatar: string;
  online: boolean;
  addedAt: string;
};

export function createFriendAPI(client: HttpClient) {
  return {
    list: () => client.get<FriendItem[]>("/friend/list"),
    add: (userId: number) => client.post<{ ok: boolean }>("/friend/add", { user_id: userId }),
    remove: (userId: number) => client.delete<{ ok: boolean }>(`/friend/${userId}`),
  };
}

export type FriendAPI = ReturnType<typeof createFriendAPI>;
