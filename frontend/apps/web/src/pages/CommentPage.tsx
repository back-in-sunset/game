import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "react-router-dom";
import {
  Avatar,
  Badge,
  Button,
  Card,
  Divider,
} from "@game/ui";
import {
  COMMENT_SORT_CREATED_TIME,
  COMMENT_SORT_LIKE_COUNT,
  type CommentItem,
  type CommentSortType,
} from "@game/api";
import { getCommentAPI } from "../services/http";
import { useAuthStore } from "../store/authStore";
import { WebShell } from "../components/WebShell";

const DEFAULT_OBJ_ID = "1001";
const DEFAULT_OBJ_TYPE = "1";
const DEFAULT_PAGE_SIZE = "20";
const MAX_COMMENT_LENGTH = 1000;

type NoticeTone = "success" | "warning" | "error";

type Notice = {
  tone: NoticeTone;
  text: string;
} | null;

function parsePositiveInt(value: string): number | null {
  const parsed = Number.parseInt(value.trim(), 10);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : null;
}

function formatTime(ts: number): string {
  if (!ts) return "未知时间";
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  }).format(new Date(ts * 1000));
}

function formatRelativeTime(ts: number): string {
  if (!ts) return "刚刚";
  const diff = Math.max(0, Date.now() - ts * 1000);
  const minutes = Math.floor(diff / 60000);
  if (minutes < 1) return "刚刚";
  if (minutes < 60) return `${minutes} 分钟前`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} 小时前`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `${days} 天前`;
  return formatTime(ts);
}

function mergeComments(prev: CommentItem[], next: CommentItem[]): CommentItem[] {
  const seen = new Set<number>();
  const merged: CommentItem[] = [];
  for (const item of [...prev, ...next]) {
    if (seen.has(item.id)) continue;
    seen.add(item.id);
    merged.push(item);
  }
  return merged;
}

export function CommentPage() {
  const currentUserId = useAuthStore((s) => s.userId);
  const nickname = useAuthStore((s) => s.nickname);
  const isLoggedIn = Boolean(currentUserId && currentUserId > 0);
  const canModerate = currentUserId === 1;

  const [draftObjId, setDraftObjId] = useState(DEFAULT_OBJ_ID);
  const [draftObjType, setDraftObjType] = useState(DEFAULT_OBJ_TYPE);
  const [draftPageSize, setDraftPageSize] = useState(DEFAULT_PAGE_SIZE);
  const [query, setQuery] = useState({ objId: 1001, objType: 1, pageSize: 20 });
  const [sortType, setSortType] = useState<CommentSortType>(COMMENT_SORT_CREATED_TIME);

  const [comments, setComments] = useState<CommentItem[]>([]);
  const [cursor, setCursor] = useState(0);
  const [lastId, setLastId] = useState(0);
  const [isEnd, setIsEnd] = useState(false);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);

  const [message, setMessage] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [likedIds, setLikedIds] = useState<Set<number>>(new Set());

  const [moderationIdInput, setModerationIdInput] = useState("");
  const [moderationTarget, setModerationTarget] = useState<CommentItem | null>(null);
  const [moderationLoading, setModerationLoading] = useState(false);
  const listRequestSeq = useRef(0);
  const moderationRequestSeq = useRef(0);

  const messageLength = useMemo(() => Array.from(message.trim()).length, [message]);
  const myCommentCount = useMemo(
    () => comments.filter((item) => item.memberId === currentUserId).length,
    [comments, currentUserId],
  );
  const totalLikeCount = useMemo(
    () => comments.reduce((sum, item) => sum + item.likeCount, 0),
    [comments],
  );
  const sortLabel = sortType === COMMENT_SORT_CREATED_TIME ? "按时间" : "按点赞";

  const loadFirstPage = useCallback(async (preserveNotice = false) => {
    const requestId = ++listRequestSeq.current;
    setLoading(true);
    if (!preserveNotice) setNotice(null);
    setComments([]);
    setCursor(0);
    setLastId(0);
    setIsEnd(false);

    try {
      const resp = await getCommentAPI().list({
        objId: query.objId,
        objType: query.objType,
        memberId: currentUserId ?? undefined,
        cursor: 0,
        pageSize: query.pageSize,
        sortType,
        commentId: 0,
        rootId: 0,
      });
      if (requestId !== listRequestSeq.current) return;
      setComments(resp.list);
      setCursor(resp.cursor);
      setLastId(resp.lastId);
      setIsEnd(resp.isEnd || resp.list.length === 0);
      if (resp.list.length === 0 && !preserveNotice) {
        setNotice({ tone: "warning", text: "当前对象还没有评论" });
      }
    } catch (e) {
      if (requestId !== listRequestSeq.current) return;
      setNotice({ tone: "error", text: e instanceof Error ? e.message : "加载评论失败" });
    } finally {
      if (requestId === listRequestSeq.current) {
        setLoading(false);
      }
    }
  }, [currentUserId, query.objId, query.objType, query.pageSize, sortType]);

  const loadMore = useCallback(async () => {
    if (isEnd || loadingMore) return;

    const requestId = ++listRequestSeq.current;
    setLoadingMore(true);
    setNotice(null);
    try {
      const resp = await getCommentAPI().list({
        objId: query.objId,
        objType: query.objType,
        memberId: currentUserId ?? undefined,
        cursor,
        pageSize: query.pageSize,
        sortType,
        commentId: lastId,
        rootId: 0,
      });
      if (requestId !== listRequestSeq.current) return;
      setComments((prev) => mergeComments(prev, resp.list));
      setCursor(resp.cursor);
      setLastId(resp.lastId);
      setIsEnd(resp.isEnd || resp.list.length === 0);
    } catch (e) {
      if (requestId !== listRequestSeq.current) return;
      setNotice({ tone: "error", text: e instanceof Error ? e.message : "加载更多失败" });
    } finally {
      if (requestId === listRequestSeq.current) {
        setLoadingMore(false);
      }
    }
  }, [currentUserId, cursor, isEnd, lastId, loadingMore, query.objId, query.objType, query.pageSize, sortType]);

  useEffect(() => {
    void loadFirstPage();
  }, [loadFirstPage]);

  const applyQuery = useCallback(() => {
    const nextObjId = parsePositiveInt(draftObjId);
    const nextObjType = parsePositiveInt(draftObjType);
    const nextPageSize = parsePositiveInt(draftPageSize);
    if (!nextObjId || !nextObjType || !nextPageSize) {
      setNotice({ tone: "error", text: "请输入有效的 obj_id / obj_type / page_size" });
      return;
    }

    setLikedIds(new Set());
    setModerationTarget(null);
    setModerationIdInput("");
    setQuery({
      objId: nextObjId,
      objType: nextObjType,
      pageSize: nextPageSize,
    });
  }, [draftObjId, draftObjType, draftPageSize]);

  const loadModerationTarget = useCallback(
    async (commentId?: number, silent = false) => {
      if (moderationLoading) return;
      const nextId = commentId ?? parsePositiveInt(moderationIdInput);
      if (!nextId) {
        setNotice({ tone: "error", text: "请输入有效的评论 ID" });
        return;
      }

      const requestId = ++moderationRequestSeq.current;
      setModerationLoading(true);
      setNotice(null);
      try {
        const resp = await getCommentAPI().get(query.objId, nextId, query.objType);
        if (requestId !== moderationRequestSeq.current) return;
        setModerationTarget(resp);
        setModerationIdInput(String(resp.id));
        if (!silent) {
          setNotice({ tone: "success", text: `已加载评论 #${resp.id}` });
        }
      } catch (e) {
        if (requestId !== moderationRequestSeq.current) return;
        setModerationTarget(null);
        if (!silent) {
          setNotice({ tone: "error", text: e instanceof Error ? e.message : "加载评论失败" });
        }
      } finally {
        if (requestId === moderationRequestSeq.current) {
          setModerationLoading(false);
        }
      }
    },
    [moderationIdInput, moderationLoading, query.objId, query.objType],
  );

  const mutateComment = useCallback(
    async (kind: "like" | "unlike" | "delete" | "block" | "unblock" | "pin" | "unpin", targetId: number) => {
      if (!currentUserId) {
        setNotice({ tone: "warning", text: "请先登录，再执行评论操作" });
        return;
      }

      setNotice(null);
      try {
        const api = getCommentAPI();
        const payload = {
          objId: query.objId,
          objType: query.objType,
          commentId: targetId,
          memberId: currentUserId,
        };

        if (kind === "like") {
          await api.like(payload);
          setLikedIds((prev) => new Set(prev).add(targetId));
        } else if (kind === "unlike") {
          await api.unlike(payload);
          setLikedIds((prev) => {
            const next = new Set(prev);
            next.delete(targetId);
            return next;
          });
        } else if (kind === "delete") {
          await api.delete(payload);
        } else if (kind === "block") {
          await api.block(payload);
        } else if (kind === "unblock") {
          await api.unblock(payload);
        } else if (kind === "pin") {
          await api.pin(payload);
        } else if (kind === "unpin") {
          await api.unpin(payload);
        }

        await loadFirstPage(true);
        if (moderationTarget?.id === targetId) {
          await loadModerationTarget(targetId, true);
        }
        setNotice({ tone: "success", text: `已执行 ${kind} #${targetId}` });
      } catch (e) {
        setNotice({ tone: "error", text: e instanceof Error ? e.message : "操作失败" });
      }
    },
    [currentUserId, loadFirstPage, loadModerationTarget, moderationTarget?.id, query.objId, query.objType],
  );

  const handleSubmitComment = useCallback(async () => {
    if (submitting) return;
    const text = message.trim();
    if (!currentUserId) {
      setNotice({ tone: "warning", text: "登录后才能发表评论" });
      return;
    }
    if (!text) {
      setNotice({ tone: "warning", text: "评论内容不能为空" });
      return;
    }
    if (Array.from(text).length > MAX_COMMENT_LENGTH) {
      setNotice({ tone: "warning", text: "评论不能超过 1000 字" });
      return;
    }

    setSubmitting(true);
    setNotice(null);
    try {
      await getCommentAPI().create({
        objId: query.objId,
        objType: query.objType,
        memberId: currentUserId,
        message: text,
      });
      setMessage("");
      await loadFirstPage(true);
      setNotice({ tone: "success", text: "评论已发布" });
    } catch (e) {
      setNotice({ tone: "error", text: e instanceof Error ? e.message : "发布失败" });
    } finally {
      setSubmitting(false);
    }
  }, [currentUserId, loadFirstPage, message, query.objId, query.objType, submitting]);

  const currentStats = useMemo(
    () => [
      { label: "已加载", value: String(comments.length) },
      { label: "我的评论", value: String(myCommentCount) },
      { label: "总点赞", value: String(totalLikeCount) },
      { label: "分页", value: loading ? "加载中..." : isEnd ? "已到底" : "继续加载" },
    ],
    [comments.length, isEnd, loading, myCommentCount, totalLikeCount],
  );

  return (
    <WebShell
      title="评论中心"
      subtitle="评论列表 / 发布 / 点赞 / 屏蔽 / 置顶"
      activeTab="comments"
      rightRail={
        <div className="stack">
          <div className="sectionHeader">
            <div className="titleBlock">
              <h3>对象配置</h3>
              <p>按照 obj_id / obj_type 读取评论</p>
            </div>
            <Badge tone="neutral">{sortLabel}</Badge>
          </div>

          <div className="stack">
            <input
              className="searchInput"
              type="number"
              min={1}
              value={draftObjId}
              onChange={(e) => setDraftObjId(e.target.value)}
              placeholder="obj_id"
            />
            <input
              className="searchInput"
              type="number"
              min={1}
              value={draftObjType}
              onChange={(e) => setDraftObjType(e.target.value)}
              placeholder="obj_type"
            />
            <input
              className="searchInput"
              type="number"
              min={1}
              value={draftPageSize}
              onChange={(e) => setDraftPageSize(e.target.value)}
              placeholder="page_size"
            />
            <select
              className="searchInput"
              value={sortType}
              onChange={(e) => {
                const next = Number(e.target.value) as CommentSortType;
                setSortType(next);
              }}
            >
              <option value={COMMENT_SORT_CREATED_TIME}>按时间</option>
              <option value={COMMENT_SORT_LIKE_COUNT}>按点赞</option>
            </select>

            <div className="buttonGrid">
              <Button variant="primary" onClick={applyQuery}>
                应用筛选
              </Button>
              <Button variant="ghost" onClick={() => void loadFirstPage()}>
                刷新列表
              </Button>
            </div>
          </div>

          <Divider />

          <div className="stack">
            <div className="statusRow">
              <span>当前用户</span>
              <Badge tone={isLoggedIn ? "success" : "warning"}>{isLoggedIn ? `#${currentUserId}` : "未登录"}</Badge>
            </div>
            <div className="statusRow">
              <span>昵称</span>
              <Badge tone="neutral">{nickname || "未设置"}</Badge>
            </div>
            <div className="statusRow">
              <span>排序</span>
              <Badge tone="success">{sortLabel}</Badge>
            </div>
            <div className="statusRow">
              <span>cursor / last_id</span>
              <Badge tone="neutral">{cursor} / {lastId}</Badge>
            </div>
          </div>

          <Divider />

          <div className="stack">
            <div className="sectionHeader">
              <div className="titleBlock">
                <h3>管理入口</h3>
                <p>置顶 / 屏蔽 / 取消</p>
              </div>
              <Badge tone={canModerate ? "success" : "warning"}>{canModerate ? "管理员模式" : "受限"}</Badge>
            </div>

            {canModerate ? (
              <>
                <input
                  className="searchInput"
                  type="number"
                  min={1}
                  value={moderationIdInput}
                  onChange={(e) => setModerationIdInput(e.target.value)}
                  placeholder="输入评论 ID"
                />
                <div className="buttonGrid">
                  <Button variant="primary" onClick={() => void loadModerationTarget()}>
                    {moderationLoading ? "加载中..." : "加载评论"}
                  </Button>
                  <Button variant="ghost" onClick={() => { setModerationTarget(null); setModerationIdInput(""); }}>
                    清空
                  </Button>
                </div>

                {moderationTarget ? (
                  <div className="postCard commentCard target">
                    <div className="postHeader">
                      <div className="postAuthor">
                        <Avatar name={`U${moderationTarget.memberId}`} />
                        <div>
                          <h4>评论 #{moderationTarget.id}</h4>
                          <p>
                            {formatRelativeTime(moderationTarget.createdAt)} · {formatTime(moderationTarget.createdAt)}
                          </p>
                        </div>
                      </div>
                      <Badge tone={moderationTarget.state === 0 ? "success" : "warning"}>
                        {moderationTarget.state === 0 ? "正常" : "隐藏"}
                      </Badge>
                    </div>

                    <p className="commentBody">{moderationTarget.message}</p>

                    <div className="commentMetaRow">
                      <Badge tone="neutral">楼层 {moderationTarget.floor}</Badge>
                      <Badge tone="neutral">赞 {moderationTarget.likeCount}</Badge>
                      <Badge tone="neutral">踩 {moderationTarget.hateCount}</Badge>
                      <Badge tone="neutral">子评论 {moderationTarget.count}</Badge>
                    </div>

                    <div className="buttonGrid">
                      <Button variant="secondary" onClick={() => void mutateComment("block", moderationTarget.id)}>
                        屏蔽
                      </Button>
                      <Button variant="ghost" onClick={() => void mutateComment("unblock", moderationTarget.id)}>
                        取消屏蔽
                      </Button>
                      <Button variant="secondary" onClick={() => void mutateComment("pin", moderationTarget.id)}>
                        置顶
                      </Button>
                      <Button variant="ghost" onClick={() => void mutateComment("unpin", moderationTarget.id)}>
                        取消置顶
                      </Button>
                    </div>
                  </div>
                ) : (
                  <div className="emptyState">输入评论 ID，或者点击列表里的「设为目标」。</div>
                )}
              </>
            ) : (
              <div className="emptyState">当前账号未开放管理操作。默认用 `userId = 1` 演示管理入口。</div>
            )}
          </div>

          {notice ? <p className={notice.tone === "error" ? "errorText" : "mutedText"}>{notice.text}</p> : null}
        </div>
      }
    >
      <section className="feed">
        <Card title="发表评论" subtitle={`当前对象 #${query.objId} / type ${query.objType}`}>
          {isLoggedIn ? (
            <form
              className="commentComposer"
              onSubmit={(e) => {
                e.preventDefault();
                void handleSubmitComment();
              }}
            >
              <textarea
                className="commentTextarea"
                value={message}
                maxLength={MAX_COMMENT_LENGTH}
                onChange={(e) => setMessage(e.target.value)}
                placeholder="写点什么，支持普通评论"
              />
              <div className="commentComposerFooter">
                <span className="mutedText">
                  {messageLength} / {MAX_COMMENT_LENGTH}
                </span>
                <div className="buttonGrid">
                  <button className="button primary" type="submit">
                    {submitting ? "发布中..." : "发布评论"}
                  </button>
                  <Button variant="ghost" onClick={() => setMessage("")}>
                    清空
                  </Button>
                </div>
              </div>
            </form>
          ) : (
            <div className="emptyState">
              登录后才能发表评论和执行点赞操作。
              <div style={{ marginTop: 12 }}>
                <Link to="/login" className="button primary buttonLink">
                  去登录
                </Link>
              </div>
            </div>
          )}
        </Card>

        <Card
          title="评论列表"
          subtitle={loading ? "加载中..." : `共 ${comments.length} 条 · ${isEnd ? "没有更多了" : "可继续加载"}`}
        >
          <div className="commentList">
            {comments.map((comment) => {
              const isMine = comment.memberId === currentUserId;
              const isModerationTarget = moderationTarget?.id === comment.id;
              const liked = likedIds.has(comment.id);

              return (
                <article
                  key={comment.id}
                  className={`postCard commentCard ${isMine ? "mine" : ""} ${isModerationTarget ? "target" : ""}`}
                >
                  <div className="postHeader">
                    <div className="postAuthor">
                      <Avatar name={isMine ? (nickname || "我") : `U${comment.memberId}`} />
                      <div>
                        <h4>{isMine ? (nickname || "我") : `用户 #${comment.memberId}`}</h4>
                        <p>
                          ID {comment.id} · {formatRelativeTime(comment.createdAt)}
                        </p>
                      </div>
                    </div>
                    <div className="commentMetaRow">
                      <Badge tone="neutral">楼层 {comment.floor}</Badge>
                      <Badge tone={comment.likeCount > 0 ? "success" : "neutral"}>赞 {comment.likeCount}</Badge>
                    </div>
                  </div>

                  <p className="commentBody">{comment.message}</p>

                  <div className="commentMetaRow">
                    <span className="mutedText">创建于 {formatTime(comment.createdAt)}</span>
                    {comment.count > 0 ? <Badge tone="neutral">子评论 {comment.count}</Badge> : null}
                    {comment.replyId > 0 ? <Badge tone="warning">回复 #{comment.replyId}</Badge> : null}
                    {comment.rootId > 0 ? <Badge tone="neutral">根 #{comment.rootId}</Badge> : null}
                  </div>

                  <div className="buttonGrid" style={{ marginTop: 14 }}>
                    <Button
                      variant={liked ? "ghost" : "primary"}
                      onClick={() => void mutateComment(liked ? "unlike" : "like", comment.id)}
                    >
                      {liked ? "取消点赞" : "点赞"}
                    </Button>
                    {isMine ? (
                      <Button variant="ghost" onClick={() => void mutateComment("delete", comment.id)}>
                        删除
                      </Button>
                    ) : null}
                    {canModerate ? (
                      <Button
                        variant={isModerationTarget ? "primary" : "secondary"}
                        onClick={() => {
                          setModerationIdInput(String(comment.id));
                          setModerationTarget(comment);
                        }}
                      >
                        设为目标
                      </Button>
                    ) : null}
                  </div>
                </article>
              );
            })}

            {!loading && comments.length === 0 ? <div className="emptyState">还没有评论，先发布一条。</div> : null}
          </div>

          <Divider />

          <div className="toolbar">
            <div className="statusRow">
              <span>当前排序</span>
              <Badge tone="success">{sortLabel}</Badge>
            </div>
            {loading ? (
              <Badge tone="warning">加载中...</Badge>
            ) : !isEnd ? (
              <Button variant="secondary" onClick={() => void loadMore()}>
                {loadingMore ? "加载中..." : "加载更多"}
              </Button>
            ) : (
              <Badge tone="neutral">已加载完毕</Badge>
            )}
          </div>
        </Card>
      </section>

      <aside className="stack">
        <Card title="统计" subtitle="当前筛选下的快速概览">
          <div className="stack">
            {currentStats.map((item) => (
              <div key={item.label} className="statusRow">
                <span>{item.label}</span>
                <Badge tone={item.label === "分页" ? "warning" : "neutral"}>{item.value}</Badge>
              </div>
            ))}
          </div>
        </Card>

        <Card title="说明" subtitle="这页直接对接 comment service">
          <div className="stack">
            <div className="mutedText">• `obj_id / obj_type` 是真实后端 contract。</div>
            <div className="mutedText">• 点赞/取消点赞使用 `member_id`。</div>
            <div className="mutedText">• 置顶/屏蔽通过右侧管理入口操作。</div>
          </div>
        </Card>
      </aside>
    </WebShell>
  );
}
