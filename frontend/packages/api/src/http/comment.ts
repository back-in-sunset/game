import type { HttpClient } from "./client";

export type CommentItem = {
  id: string;
  userId: number;
  nickname: string;
  content: string;
  createdAt: string;
};

export function createCommentAPI(client: HttpClient) {
  return {
    list: (targetType: string, targetId: string) =>
      client.get<CommentItem[]>("/comment/list", { target_type: targetType, target_id: targetId }),
    create: (targetType: string, targetId: string, content: string) =>
      client.post<CommentItem>("/comment/create", { target_type: targetType, target_id: targetId, content }),
  };
}

export type CommentAPI = ReturnType<typeof createCommentAPI>;
