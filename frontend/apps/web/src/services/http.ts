import { HttpClient } from "@game/api";
import type { UserAPI } from "@game/api";
import type { FriendAPI } from "@game/api";
import type { CommentAPI } from "@game/api";
import type { HistoryAPI } from "@game/api";
import { createUserAPI, createFriendAPI, createCommentAPI, createHistoryAPI } from "@game/api";
import { useAuthStore } from "../store/authStore";

let _http: HttpClient | null = null;
let _user: UserAPI | null = null;
let _friend: FriendAPI | null = null;
let _comment: CommentAPI | null = null;
let _history: HistoryAPI | null = null;

function getClient(): HttpClient {
  const { baseUrl, token } = useAuthStore.getState();
  if (!_http || (_http as unknown as { _token?: string })._token !== token) {
    _http = new HttpClient(baseUrl, token);
  }
  return _http;
}

export function getHttp(): HttpClient {
  return getClient();
}

export function getUserAPI(): UserAPI {
  if (!_user) _user = createUserAPI(getClient());
  return _user;
}

export function getFriendAPI(): FriendAPI {
  if (!_friend) _friend = createFriendAPI(getClient());
  return _friend;
}

export function getCommentAPI(): CommentAPI {
  if (!_comment) _comment = createCommentAPI(getClient());
  return _comment;
}

export function getHistoryAPI(): HistoryAPI {
  if (!_history) _history = createHistoryAPI(getClient());
  return _history;
}

export function resetAPIs(): void {
  _http = null;
  _user = null;
  _friend = null;
  _comment = null;
  _history = null;
}
