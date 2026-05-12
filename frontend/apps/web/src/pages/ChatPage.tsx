import { useEffect, useMemo, useState } from "react";
import { Link, useParams } from "react-router-dom";
import { Avatar, Badge, Button, Card, Composer, ConversationRow, Divider, MessageBubble } from "@game/ui";
import { conversations } from "@game/shared";
import { useIMStore } from "../store/imStore";
import { WebShell } from "../components/WebShell";

export function ChatPage() {
  const { userId } = useParams<{ userId: string }>();
  const [query, setQuery] = useState("");
  const { messages, sendMessage, setActiveConversation, conversations: storeConvs, loadConversations } = useIMStore();

  useEffect(() => {
    loadConversations(conversations);
  }, [loadConversations]);

  const conv = useMemo(
    () => storeConvs.find((item) => item.userId === Number(userId)) ?? conversations.find((item) => item.userId === Number(userId)),
    [storeConvs, userId],
  );

  useEffect(() => {
    if (conv) setActiveConversation(conv.id);
  }, [conv?.id, setActiveConversation]);

  if (!conv) {
    return (
      <WebShell title="聊天框" subtitle="未找到对话" activeTab="chat">
        <section className="feed">
          <Card title="未找到对话" subtitle={`userId: ${userId}`}>
            <Link to="/">返回广场</Link>
          </Card>
        </section>
      </WebShell>
    );
  }

  const filteredMessages = query
    ? messages.filter((msg) => msg.text.toLowerCase().includes(query.toLowerCase()))
    : messages;

  return (
    <WebShell
      title="聊天框"
      subtitle={`当前会话：${conv.title}`}
      activeTab="chat"
      rightRail={
        <div className="stack">
          <div className="sectionHeader">
            <div className="titleBlock">
              <h3>会话信息</h3>
              <p>快速查看在线、状态和常用动作</p>
            </div>
          </div>
          <div className="profileRow">
            <Avatar name={conv.title} />
            <div>
              <strong>{conv.title}</strong>
              <p>{conv.subtitle}</p>
            </div>
          </div>
          <Badge tone={conv.online ? "success" : "neutral"}>{conv.online ? "在线" : "离线"}</Badge>
          <Divider />
          <div className="stack">
            <Button variant="primary" onClick={() => sendMessage(conv.userId, "你好，刚从聊天框发出的测试消息")}>
              快速发消息
            </Button>
            <Button variant="secondary">查看资料</Button>
            <Link to="/friends" style={{ textDecoration: "none" }}>
              <Button variant="ghost">去好友管理</Button>
            </Link>
          </div>
        </div>
      }
    >
      <section className="feed">
        <Card title={conv.title} subtitle={conv.statusText}>
          <div className="channelSearch">
            <input
              className="searchInput"
              placeholder="搜索当前会话里的消息"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
            />
          </div>
          <div className="chatFeed">
            {filteredMessages.length > 0 ? (
              filteredMessages.map((msg) => <MessageBubble key={msg.id} message={msg} />)
            ) : (
              <div className="emptyState">没有匹配的消息。</div>
            )}
          </div>
          <Divider />
          <Composer
            placeholder={`给 ${conv.title} 发送消息`}
            onSend={(text) => sendMessage(conv.userId, text)}
          />
        </Card>
      </section>

      <aside className="stack">
        <Card title="会话列表" subtitle="最近互动的联系人">
          <div className="conversationList">
            {storeConvs.map((item) => (
              <ConversationRow
                key={item.id}
                conversation={item}
                active={item.id === conv.id}
                onClick={() => setActiveConversation(item.id)}
              />
            ))}
          </div>
        </Card>
      </aside>
    </WebShell>
  );
}
