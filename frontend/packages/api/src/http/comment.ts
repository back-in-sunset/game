import type { HttpClient } from "./client";

export const COMMENT_SORT_CREATED_TIME = 0 as const;
export const COMMENT_SORT_LIKE_COUNT = 1 as const;

export type CommentSortType = typeof COMMENT_SORT_CREATED_TIME | typeof COMMENT_SORT_LIKE_COUNT;

type RawCommentItem = {
  id: number;
  obj_id: number;
  obj_type: number;
  member_id: number;
  comment_id: number;
  at_member_ids: string;
  ip: string;
  platform: number;
  device: string;
  message: string;
  meta: string;
  reply_id: number;
  state: number;
  root_id: number;
  created_at: number;
  floor: number;
  like_count: number;
  hate_count: number;
  count: number;
};

type RawCommentListResponse = {
  list: RawCommentItem[];
  is_end: boolean;
  cursor: number;
  last_id: number;
};

type RawCommentActionResponse = {
  success: boolean;
  message: string;
};

export type CommentItem = {
  id: number;
  objId: number;
  objType: number;
  memberId: number;
  commentId: number;
  atMemberIds: string;
  ip: string;
  platform: number;
  device: string;
  message: string;
  meta: string;
  replyId: number;
  state: number;
  rootId: number;
  createdAt: number;
  floor: number;
  likeCount: number;
  hateCount: number;
  count: number;
};

export type CommentListRequest = {
  objId: number;
  objType: number;
  memberId?: number;
  cursor?: number;
  pageSize?: number;
  sortType?: CommentSortType;
  commentId?: number;
  rootId?: number;
  replyId?: number;
  sortedField?: number;
};

export type CommentListResponse = {
  list: CommentItem[];
  isEnd: boolean;
  cursor: number;
  lastId: number;
};

export type CommentCreateRequest = {
  objId: number;
  objType: number;
  memberId: number;
  message: string;
  commentId?: number;
  atMemberIds?: string;
  ip?: string;
  platform?: number;
  device?: string;
  meta?: string;
  replyId?: number;
  state?: number;
  rootId?: number;
};

export type CommentMutationRequest = {
  objId: number;
  objType: number;
  commentId: number;
  memberId: number;
};

export type CommentActionResponse = {
  success: boolean;
  message: string;
};

function toQuery(params: Record<string, string | number | boolean | null | undefined>): Record<string, string> {
  const query: Record<string, string> = {};
  Object.entries(params).forEach(([key, value]) => {
    if (value !== undefined && value !== null) query[key] = String(value);
  });
  return query;
}

function normalizeComment(item: RawCommentItem): CommentItem {
  return {
    id: item.id,
    objId: item.obj_id,
    objType: item.obj_type,
    memberId: item.member_id,
    commentId: item.comment_id,
    atMemberIds: item.at_member_ids,
    ip: item.ip,
    platform: item.platform,
    device: item.device,
    message: item.message,
    meta: item.meta,
    replyId: item.reply_id,
    state: item.state,
    rootId: item.root_id,
    createdAt: item.created_at,
    floor: item.floor,
    likeCount: item.like_count,
    hateCount: item.hate_count,
    count: item.count,
  };
}

function normalizeActionResponse(resp: RawCommentActionResponse): CommentActionResponse {
  return { success: resp.success, message: resp.message };
}

function commentMutationPayload(payload: CommentMutationRequest) {
  return {
    obj_id: payload.objId,
    obj_type: payload.objType,
    comment_id: payload.commentId,
    member_id: payload.memberId,
  };
}

export function createCommentAPI(client: HttpClient) {
  const create = async (payload: CommentCreateRequest): Promise<CommentItem> => {
    const resp = await client.post<RawCommentItem>("/api/comments", {
      obj_id: payload.objId,
      obj_type: payload.objType,
      member_id: payload.memberId,
      comment_id: payload.commentId,
      at_member_ids: payload.atMemberIds,
      ip: payload.ip,
      platform: payload.platform,
      device: payload.device,
      message: payload.message,
      meta: payload.meta,
      reply_id: payload.replyId,
      state: payload.state,
      root_id: payload.rootId,
    });
    return normalizeComment(resp);
  };

  return {
    list: async (params: CommentListRequest): Promise<CommentListResponse> => {
      const resp = await client.get<RawCommentListResponse>("/api/comments", toQuery({
        obj_id: params.objId,
        obj_type: params.objType,
        member_id: params.memberId,
        cursor: params.cursor,
        page_size: params.pageSize,
        sort_type: params.sortType,
        comment_id: params.commentId,
        root_id: params.rootId,
        reply_id: params.replyId,
        sorted_field: params.sortedField,
      }));

      return {
        list: resp.list.map(normalizeComment),
        isEnd: resp.is_end,
        cursor: resp.cursor,
        lastId: resp.last_id,
      };
    },
    get: async (objId: number, commentId: number, objType: number) =>
      normalizeComment(
        await client.get<RawCommentItem>(`/api/comments/${commentId}`, toQuery({
          obj_id: objId,
          obj_type: objType,
        })),
      ),
    create,
    add: create,
    delete: async (payload: CommentMutationRequest) =>
      normalizeComment(
        await client.delete<RawCommentItem>("/api/comments", toQuery(commentMutationPayload(payload))),
      ),
    like: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.post<RawCommentActionResponse>(`/api/comments/${payload.commentId}/like`, commentMutationPayload(payload)),
      ),
    unlike: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.delete<RawCommentActionResponse>(`/api/comments/${payload.commentId}/like`, toQuery(commentMutationPayload(payload))),
      ),
    block: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.post<RawCommentActionResponse>(`/api/comments/${payload.commentId}/block`, commentMutationPayload(payload)),
      ),
    unblock: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.delete<RawCommentActionResponse>(`/api/comments/${payload.commentId}/block`, toQuery(commentMutationPayload(payload))),
      ),
    pin: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.post<RawCommentActionResponse>(`/api/comments/${payload.commentId}/pin`, commentMutationPayload(payload)),
      ),
    unpin: async (payload: CommentMutationRequest) =>
      normalizeActionResponse(
        await client.delete<RawCommentActionResponse>(`/api/comments/${payload.commentId}/pin`, toQuery(commentMutationPayload(payload))),
      ),
  };
}

export type CommentAPI = ReturnType<typeof createCommentAPI>;
