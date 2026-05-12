import { useCallback, useEffect, useMemo, useState } from "react";
import { Link } from "react-router-dom";
import { Avatar, Badge, Button, Card, Divider } from "@game/ui";
import { getFriendAPI } from "../services/http";
import type { FriendItem } from "@game/api";
import { WebShell } from "../components/WebShell";

export function FriendsPage() {
  const [friends, setFriends] = useState<FriendItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [query, setQuery] = useState("");

  const [showAddForm, setShowAddForm] = useState(false);
  const [newFriendId, setNewFriendId] = useState("");
  const [adding, setAdding] = useState(false);
  const [actionMsg, setActionMsg] = useState<{ text: string; tone: "success" | "error" } | null>(null);

  const loadFriends = useCallback(() => {
    setLoading(true);
    setError("");
    getFriendAPI()
      .list()
      .then(setFriends)
      .catch((e) => setError(e instanceof Error ? e.message : "failed to load friends"))
      .finally(() => setLoading(false));
  }, []);

  useEffect(() => {
    loadFriends();
  }, [loadFriends]);

  const handleAdd = async () => {
    const id = parseInt(newFriendId, 10);
    if (!id || id <= 0) {
      setActionMsg({ text: "请输入有效的用户 ID", tone: "error" });
      return;
    }
    setAdding(true);
    setActionMsg(null);
    try {
      await getFriendAPI().add(id);
      setActionMsg({ text: "好友请求已发送", tone: "success" });
      setNewFriendId("");
      setShowAddForm(false);
      loadFriends();
    } catch (e) {
      setActionMsg({ text: e instanceof Error ? e.message : "添加失败", tone: "error" });
    } finally {
      setAdding(false);
    }
  };

  const handleRemove = async (friend: FriendItem) => {
    if (!window.confirm(`确定要删除好友「${friend.nickname}」吗？`)) return;
    setActionMsg(null);
    try {
      await getFriendAPI().remove(friend.userId);
      setFriends((prev) => prev.filter((f) => f.userId !== friend.userId));
      setActionMsg({ text: `已删除好友「${friend.nickname}」`, tone: "success" });
    } catch (e) {
      setActionMsg({ text: e instanceof Error ? e.message : "删除失败", tone: "error" });
    }
  };

  const filtered = useMemo(() => {
    const text = query.trim().toLowerCase();
    if (!text) return friends;
    return friends.filter((friend) =>
      [friend.nickname, String(friend.userId)].some((item) => item.toLowerCase().includes(text)),
    );
  }, [friends, query]);

  const ManagePanel = (
    <div className="stack">
      <div className="sectionHeader">
        <div className="titleBlock">
          <h3>管理面板</h3>
          <p>添加或维护好友关系</p>
        </div>
      </div>

      {showAddForm ? (
        <div className="stack">
          <input
            className="searchInput"
            type="number"
            placeholder="输入用户 ID"
            value={newFriendId}
            onChange={(e) => setNewFriendId(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleAdd()}
          />
          <div className="buttonGrid">
            <Button variant="primary" onClick={handleAdd}>
              {adding ? "发送中..." : "发送好友请求"}
            </Button>
            <Button variant="ghost" onClick={() => { setShowAddForm(false); setNewFriendId(""); setActionMsg(null); }}>
              取消
            </Button>
          </div>
        </div>
      ) : (
        <div className="stack">
          <Button variant="primary" onClick={() => setShowAddForm(true)}>
            添加好友
          </Button>
        </div>
      )}

      {actionMsg ? (
        <p className={actionMsg.tone === "success" ? "mutedText" : "errorText"}>{actionMsg.text}</p>
      ) : null}

      <Divider />

      <div className="stack">
        <div className="statusRow">
          <span>好友总数</span>
          <Badge tone="neutral">{friends.length}</Badge>
        </div>
        <div className="statusRow">
          <span>在线</span>
          <Badge tone="success">{friends.filter((f) => f.online).length}</Badge>
        </div>
      </div>
    </div>
  );

  return (
    <WebShell
      title="好友管理"
      subtitle="查看联系人、在线状态和最近添加的人"
      activeTab="friends"
      rightRail={ManagePanel}
    >
      <section className="feed">
        <Card
          title="好友列表"
          subtitle={loading ? "加载中..." : `共 ${filtered.length} 位好友`}
        >
          <div className="channelSearch">
            <input
              className="searchInput"
              placeholder="搜索昵称或用户 ID"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>

          {error ? <p className="errorText">{error}</p> : null}

          {!loading && !error && friends.length === 0 ? (
            <div className="emptyState">
              还没有好友。点击右上角「添加好友」开始。
            </div>
          ) : (
            <div className="stack">
              {filtered.map((friend) => (
                <div key={friend.userId} className="postCard">
                  <div className="postHeader">
                    <div className="postAuthor">
                      <Avatar name={friend.nickname} />
                      <div>
                        <h4>{friend.nickname}</h4>
                        <p>ID {friend.userId} · {friend.addedAt}</p>
                      </div>
                    </div>
                    <Badge tone={friend.online ? "success" : "neutral"}>
                      {friend.online ? "在线" : "离线"}
                    </Badge>
                  </div>
                  <div className="postActions">
                    <Link to={`/chat/${friend.userId}`} className="button primary buttonLink">
                      发消息
                    </Link>
                    <Button variant="ghost" onClick={() => handleRemove(friend)}>
                      删除好友
                    </Button>
                  </div>
                </div>
              ))}
              {!loading && !error && filtered.length === 0 && friends.length > 0 ? (
                <div className="emptyState">没有匹配的好友，试试换个关键词。</div>
              ) : null}
            </div>
          )}
        </Card>
      </section>

      <aside className="stack">
        <Card title="好友概览" subtitle="基于当前列表的简单统计">
          <div className="stack">
            <div className="statusRow">
              <span>在线率</span>
              <Badge tone="success">
                {friends.length
                  ? `${Math.round((friends.filter((f) => f.online).length / friends.length) * 100)}%`
                  : "0%"}
              </Badge>
            </div>
            <div className="statusRow">
              <span>最近添加</span>
              <Badge tone="neutral">{friends[0]?.addedAt ?? "暂无"}</Badge>
            </div>
          </div>
        </Card>

        <Card title="快捷操作" subtitle="把好友管理接到聊天流里">
          <div className="stack">
            <Link to="/" className="linkRow">
              <span>◎</span>
              <span className="channelMeta">
                <strong>回到广场</strong>
                <span>查看帖子流</span>
              </span>
            </Link>
            <Link to="/history" className="linkRow">
              <span>☰</span>
              <span className="channelMeta">
                <strong>打开聊天框</strong>
                <span>进入会话页面</span>
              </span>
            </Link>
          </div>
        </Card>
      </aside>
    </WebShell>
  );
}
