import { useMemo } from "react";
import { Link } from "react-router-dom";
import { Avatar, Badge, Card, Composer, ConversationRow, Divider, MessageBubble } from "@game/ui";
import { conversations } from "@game/shared";
import { useIMStore } from "../store/imStore";
import { WebShell } from "../components/WebShell";
import { useIMBootstrap } from "../hooks/useIMBootstrap";

const posts = [
  {
    id: "p1",
    author: "Mina",
    role: "产品经理",
    time: "2 分钟前",
    tag: "广场",
    title: "今晚把新 UI 的三栏结构过一遍",
    body: "把广场、好友、聊天框放在同一套空间里，消息层级和导航都要比现在更明确。",
    likes: 23,
    comments: 6,
  },
  {
    id: "p2",
    author: "Jasper",
    role: "设计",
    time: "18 分钟前",
    tag: "视觉",
    title: "浅色 Discord 风格更适合当前产品",
    body: "我们保留浅色底，但把左侧入口、频道和内容区的层级做得更清晰，避免“工具台”感太重。",
    likes: 41,
    comments: 9,
  },
  {
    id: "p3",
    author: "Echo",
    role: "工程",
    time: "1 小时前",
    tag: "状态",
    title: "IM 连接和聊天框先统一到一个 shell",
    body: "不用让每个页面都单独处理连接状态，减少跳页后体验断裂。",
    likes: 15,
    comments: 2,
  },
];

const quickLinks = [
  { label: "好友管理", to: "/friends", help: "查看在线和离线联系人" },
  { label: "聊天框", to: "/history", help: "查看和发送消息" },
  { label: "语音面板", to: "/voice", help: "进入通话控制" },
];

const channelHighlights = [
  { name: "广场", hint: "帖子流", active: true },
  { name: "公告", hint: "系统发布", active: false },
  { name: "创意", hint: "设计讨论", active: false },
  { name: "开发", hint: "实现细节", active: false },
];

export function DashboardPage() {
  const { status, error, messages, conversations: storeConvs, activeConversationId, setActiveConversation, sendMessage } =
    useIMStore();
  useIMBootstrap();

  const active = useMemo(
    () => storeConvs.find((c) => c.id === activeConversationId) ?? storeConvs[0] ?? conversations[0],
    [activeConversationId, storeConvs],
  );

  return (
    <WebShell
      title="广场"
      subtitle="帖子流 / 推荐内容 / 快捷会话"
      activeTab="square"
      rightRail={
        <div className="stack">
          <div className="sectionHeader">
            <div className="titleBlock">
              <h3>推荐频道</h3>
              <p>当前最活跃的讨论板块</p>
            </div>
          </div>
          <div className="stack">
            {channelHighlights.map((item) => (
              <div key={item.name} className="statusRow">
                <span>{item.name}</span>
                <Badge tone={item.active ? "success" : "neutral"}>{item.hint}</Badge>
              </div>
            ))}
          </div>
          <Divider />
          <div className="sectionHeader">
            <div className="titleBlock">
              <h3>连接状态</h3>
              <p>广场内容与会话状态统一展示</p>
            </div>
          </div>
          <Badge tone={status === "connected" ? "success" : "warning"}>{status}</Badge>
          {error ? <p className="errorText">{error}</p> : null}
        </div>
      }
    >
      <section className="feed">
        <Card title="广场帖子" subtitle="把讨论、通知和项目动态放在同一个流里">
          <div className="postStream">
            {posts.map((post) => (
              <article key={post.id} className="postCard">
                <div className="postHeader">
                  <div className="postAuthor">
                    <Avatar name={post.author} />
                    <div>
                      <h4>{post.author}</h4>
                      <p>{post.role} · {post.time}</p>
                    </div>
                  </div>
                  <Badge tone="neutral">{post.tag}</Badge>
                </div>
                <h3 style={{ margin: 0, fontSize: 18 }}>{post.title}</h3>
                <p className="postBody" style={{ marginTop: 10 }}>{post.body}</p>
                <div className="postActions">
                  <span className="pillButton">♥ {post.likes}</span>
                  <span className="pillButton">✎ {post.comments}</span>
                  <span className="pillButton">↗ 分享</span>
                </div>
              </article>
            ))}
          </div>
        </Card>

        <div className="workspace">
          <Card title="快捷入口" subtitle="从广场直接跳到具体模块">
            <div className="stack">
              {quickLinks.map((item) => (
                <Link key={item.to} to={item.to} className="linkRow">
                  <span>↗</span>
                  <span className="channelMeta">
                    <strong>{item.label}</strong>
                    <span>{item.help}</span>
                  </span>
                </Link>
              ))}
            </div>
          </Card>

          <Card title={active.title} subtitle={active.subtitle}>
            <div className="chatFeed">
              {messages.length > 0 ? (
                messages.map((msg) => <MessageBubble key={msg.id} message={msg} />)
              ) : (
                <div className="emptyState">还没有消息，先从右侧会话里发一条试试。</div>
              )}
            </div>
            <Divider />
            <Composer
              placeholder={`给 ${active.title} 发送消息`}
              onSend={(text) => sendMessage(active.userId, text)}
            />
          </Card>
        </div>
      </section>

      <aside className="stack">
        <Card title="好友预览" subtitle="最近互动的人">
          <div className="conversationList">
            {storeConvs.map((item) => (
              <ConversationRow
                key={item.id}
                conversation={item}
                active={item.id === activeConversationId}
                onClick={() => setActiveConversation(item.id)}
              />
            ))}
          </div>
        </Card>
      </aside>
    </WebShell>
  );
}
