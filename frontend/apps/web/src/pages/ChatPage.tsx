import { useEffect } from "react";
import { useParams, Link } from "react-router-dom";
import { Card, Badge, Composer, MessageBubble, Divider } from "@game/ui";
import { useIMStore } from "../store/imStore";
import { conversations } from "@game/shared";

export function ChatPage() {
  const { userId } = useParams<{ userId: string }>();
  const { messages, sendMessage, setActiveConversation } = useIMStore();

  const conv = conversations.find((c) => c.userId === Number(userId));

  useEffect(() => {
    if (conv) setActiveConversation(conv.id);
  }, [conv?.id]);

  if (!conv) {
    return (
      <div className="shell">
        <main className="main">
          <Card title="未找到对话" subtitle={`userId: ${userId}`}>
            <Link to="/">返回仪表盘</Link>
          </Card>
        </main>
      </div>
    );
  }

  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">V</div>
          <div>
            <h1>VDA + IM Studio</h1>
          </div>
        </div>
        <Card title={conv.title} subtitle={conv.subtitle}>
          <Badge tone={conv.online ? "success" : "neutral"}>
            {conv.online ? "在线" : "离线"}
          </Badge>
        </Card>
        <Link to="/" style={{ display: "block", marginTop: 12 }}>← 返回</Link>
      </aside>
      <main className="main">
        <Card title={`与 ${conv.title} 的对话`}>
          <div className="chatFeed">
            {messages.map((msg) => (
              <MessageBubble key={msg.id} message={msg} />
            ))}
          </div>
          <Divider />
          <Composer
            placeholder={`给 ${conv.title} 发送消息`}
            onSend={(text) => sendMessage(conv.userId, text)}
          />
        </Card>
      </main>
    </div>
  );
}
