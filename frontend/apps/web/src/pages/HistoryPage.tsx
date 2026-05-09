import { useState } from "react";
import { Link } from "react-router-dom";
import { Card, MessageBubble } from "@game/ui";
import { conversations } from "@game/shared";

export function HistoryPage() {
  const [query, setQuery] = useState("");
  const allMessages = conversations.flatMap((c) => c.messages);
  const filtered = query
    ? allMessages.filter((m) => m.text.toLowerCase().includes(query.toLowerCase()))
    : allMessages;

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">V</div>
          <div><h1>历史记录</h1></div>
        </div>
        <Link to="/">← 返回</Link>
      </aside>
      <main className="main">
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
      </main>
    </div>
  );
}
