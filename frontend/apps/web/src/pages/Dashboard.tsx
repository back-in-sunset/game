import { useEffect } from "react";
import { Link } from "react-router-dom";
import {
  Card, Badge, Button, Avatar, Divider, Composer,
  ConversationRow, MessageBubble, StatusRow,
} from "@game/ui";
import { useIMStore } from "../store/imStore";
import { useVoiceStore } from "../store/voiceStore";
import { useAuthStore } from "../store/authStore";
import { conversations, callHighlights } from "@game/shared";
import { serviceConfig } from "../config";

export function DashboardPage() {
  const {
    status, error, messages, conversations: storeConvs, activeConversationId,
    connect, disconnect, loadConversations, setActiveConversation, sendMessage, sendVoiceAction,
  } = useIMStore();
  const voice = useVoiceStore();
  const { isLoggedIn, nickname, logout } = useAuthStore();

  useEffect(() => {
    loadConversations(conversations);
    if (serviceConfig.imWsUrl) {
      connect(serviceConfig.imWsUrl, {
      token: serviceConfig.imToken,
      domain: serviceConfig.imDomain,
      scope: {
        tenant_id: serviceConfig.imTenantId,
        project_id: serviceConfig.imProjectId,
        environment: serviceConfig.imEnvironment,
      },
    });
    }
    return () => disconnect();
  }, []);

  const active = storeConvs.find((c) => c.id === activeConversationId) ?? storeConvs[0];

  if (!active) {
    return (
      <div className="shell">
        <main className="main" style={{ display: "flex", justifyContent: "center", alignItems: "center", minHeight: "100vh" }}>
          <p>Loading...</p>
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
            <p>Web · Mobile · Voice</p>
          </div>
        </div>

        <Card title="会话列表" subtitle="统一承接消息与通话">
          <div className="conversationList">
            {storeConvs.map((item) => (
              <Link key={item.id} to={`/chat/${item.userId}`} style={{ textDecoration: "none", color: "inherit" }}>
                <ConversationRow
                  conversation={item}
                  active={item.id === activeConversationId}
                  onClick={() => setActiveConversation(item.id)}
                />
              </Link>
            ))}
          </div>
        </Card>

        <Card title="在线状态" subtitle="当前服务面板">
          <div className="stack">
            <StatusRow label="IM WebSocket" value={status} tone={status === "connected" ? "success" : "warning"} />
            <StatusRow label="VDA gRPC" value={serviceConfig.vdaGrpcUrl || "not configured"} tone={serviceConfig.vdaGrpcUrl ? "success" : "warning"} />
            <StatusRow label="LiveKit" value={serviceConfig.livekitUrl || "not configured"} tone={serviceConfig.livekitUrl ? "success" : "warning"} />
          </div>
          {error ? <p className="errorText">{error}</p> : null}
        </Card>

        <Card title="导航">
          <div className="stack" style={{ marginTop: 12 }}>
            {isLoggedIn() ? (
              <Badge tone="success">{nickname || "已登录"}</Badge>
            ) : (
              <Link to="/login"><Button variant="primary">登录</Button></Link>
            )}
            <Link to="/"><Button variant="secondary">仪表盘</Button></Link>
            <Link to="/voice"><Button variant="secondary">语音通话</Button></Link>
            <Link to="/friends"><Button variant="secondary">好友</Button></Link>
            <Link to="/history"><Button variant="secondary">历史记录</Button></Link>
            <Link to="/settings"><Button variant="secondary">设置</Button></Link>
            {isLoggedIn() ? (
              <Button variant="ghost" onClick={() => { logout(); disconnect(); }}>退出</Button>
            ) : null}
          </div>
        </Card>
      </aside>

      <main className="main">
        <header className="hero">
          <div>
            <Badge>Product UI · light theme</Badge>
            <h2>聊天和语音通话统一工作台</h2>
            <p>先把 IM 和 VDA 放进同一个前端骨架，后面可以无缝扩展到 React Native。</p>
          </div>
          <div className="heroStats">
            {callHighlights.map((item) => (
              <Card key={item.label} title={item.label} subtitle={item.help}>
                <strong>{item.value}</strong>
              </Card>
            ))}
          </div>
        </header>

        <section className="contentGrid">
          <Card title={active.title} subtitle={active.subtitle}>
            <div className="chatFeed">
              {messages.map((msg) => (
                <MessageBubble key={msg.id} message={msg} />
              ))}
            </div>
            <Divider />
            <Composer
              placeholder={`给 ${active.title} 发送消息`}
              onSend={(text) => sendMessage(active.userId, text)}
            />
          </Card>

          <Card title="通话控制" subtitle="VDA 面板">
            <div className="callPanel">
              <Avatar name={active.title} />
              <div>
                <strong>{active.title}</strong>
                <p>{active.statusText}</p>
              </div>
            </div>
            <div className="buttonGrid">
              <Button variant="primary" onClick={() => sendVoiceAction("call_invite", { callee: active.userId })}>
                发起通话
              </Button>
              <Button variant="secondary" onClick={() => sendVoiceAction("call_accept", { call_id: voice.callId || "call_pending" })}>
                接听
              </Button>
              <Button variant="ghost" onClick={() => sendVoiceAction("call_end", { call_id: voice.callId || "call_pending" })}>
                挂断
              </Button>
            </div>
          </Card>
        </section>
      </main>
    </div>
  );
}
