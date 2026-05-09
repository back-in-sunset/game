import type { HttpClient } from "./client";

export type UserProfile = {
  userId: number;
  nickname: string;
  avatar: string;
};

export function createUserAPI(client: HttpClient) {
  return {
    getProfile: (userId: number) => client.get<UserProfile>(`/user/${userId}`),
    searchUsers: (query: string) => client.get<UserProfile[]>(`/user/search`, { q: query }),
  };
}

export type UserAPI = ReturnType<typeof createUserAPI>;
