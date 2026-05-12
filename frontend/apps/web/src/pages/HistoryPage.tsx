import { useState } from "react";
import { Link } from "react-router-dom";
import { Card, MessageBubble } from "@game/ui";
import { conversations } from "@game/shared";
import { WebShell } from "../components/WebShell";

export function HistoryPage() {
  const [query, setQuery] = useState("");
  const allMessages = conversations.flatMap((c) => c.messages);
  const filtered = query
    ? allMessages.filter((m) => m.text.toLowerCase().includes(query.toLowerCase()))
    : allMessages;

  return (
    <WebShell title="聊天框" subtitle="历史消息搜索与浏览" activeTab="chat">
      <section className="feed">
        <Card title="所有消息" subtitle={`共 ${filtered.length} 条${query ? ` (搜索: "${query}")` : ""}`}>
          <input
            className="searchInput"
            placeholder="搜索消息..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
          />
          <div className="chatFeed" style={{ marginTop: 12 }}>
            {filtered.map((msg) => (
              <MessageBubble key={msg.id} message={msg} />
            ))}
          </div>
        </Card>
      </section>
      <aside className="stack">
        <Card title="筛选说明" subtitle="基于本地示例消息">
          <p className="mutedText">
            这里先做消息浏览框，后续可以接真实历史接口或频道过滤。
          </p>
          <Link to="/" className="linkRow">
            <span>◎</span>
            <span className="channelMeta">
              <strong>返回广场</strong>
              <span>继续浏览帖子流</span>
            </span>
          </Link>
        </Card>
      </aside>
    </WebShell>
  );
}
