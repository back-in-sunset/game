import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Card, Badge, Button } from "@game/ui";
import { getFriendAPI } from "../services/http";
import type { FriendItem } from "@game/api";

export function FriendsPage() {
  const [friends, setFriends] = useState<FriendItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    getFriendAPI().list()
      .then(setFriends)
      .catch((e) => setError(e instanceof Error ? e.message : "failed to load friends"))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">V</div>
          <div><h1>好友</h1></div>
        </div>
        <Link to="/">← 返回</Link>
      </aside>
      <main className="main">
        <Card title="好友列表" subtitle={loading ? "加载中..." : `共 ${friends.length} 位好友`}>
          {error ? <p className="errorText">{error}</p> : null}
          <div className="stack">
            {friends.map((f) => (
              <div key={f.userId} className="statusRow">
                <span>{f.nickname}</span>
                <Badge tone={f.online ? "success" : "neutral"}>
                  {f.online ? "在线" : "离线"}
                </Badge>
              </div>
            ))}
            {!loading && !error && friends.length === 0 ? (
              <p style={{ color: "#64748b", fontSize: 13 }}>暂无好友，去添加好友吧</p>
            ) : null}
          </div>
        </Card>
      </main>
    </div>
  );
}
