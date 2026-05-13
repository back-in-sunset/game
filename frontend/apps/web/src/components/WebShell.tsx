import type { ReactNode } from "react";
import { Link, NavLink } from "react-router-dom";
import { Badge, Button } from "@game/ui";
import { useAuthStore } from "../store/authStore";
import { useIMStore } from "../store/imStore";
import { useSettingsStore } from "../store/settingsStore";
import { useVoiceStore } from "../store/voiceStore";

type WebShellProps = {
  title: string;
  subtitle?: string;
  activeTab?: "square" | "friends" | "chat" | "comments";
  children: ReactNode;
  rightRail?: ReactNode;
};

const channelLinks = [
  { to: "/", label: "广场", active: "square" as const, icon: "◎" },
  { to: "/comments", label: "评论", active: "comments" as const, icon: "◫" },
  { to: "/friends", label: "好友", active: "friends" as const, icon: "◌" },
  { to: "/history", label: "聊天框", active: "chat" as const, icon: "☰" },
  { to: "/voice", label: "语音", icon: "♪" },
  { to: "/settings", label: "设置", icon: "⚙" },
];

export function WebShell({ title, subtitle, activeTab, children, rightRail }: WebShellProps) {
  const { isLoggedIn, nickname, logout } = useAuthStore();
  const { disconnect, status } = useIMStore();
  const { livekitUrl, vdaGrpcUrl } = useSettingsStore();
  const { callState } = useVoiceStore();

  return (
    <div className="shell">
      <aside className="serverRail">
        <div className="serverStack">
          <button className="serverPill active" type="button" aria-label="Home">
            G
          </button>
          <div className="serverDivider" />
          <button className="serverPill" type="button" aria-label="Community">
            S
          </button>
          <button className="serverPill" type="button" aria-label="Friends">
            F
          </button>
          <button className="serverPill" type="button" aria-label="Chat">
            C
          </button>
        </div>
      </aside>

      <aside className="channelRail">
        <div className="brand">
          <div className="brandMark">G</div>
          <div>
            <h1>Game Space</h1>
            <p>Discord-style social shell</p>
          </div>
        </div>

        <div className="channelSection">
          <div className="channelGroup">
            <div className="channelGroupHeader">
              <span>导航</span>
              <span className="mutedText">{isLoggedIn() ? nickname || "已登录" : "未登录"}</span>
            </div>
            <div className="channelList">
              {channelLinks.map((item) => (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={({ isActive }) =>
                    `channelRow ${isActive || activeTab === item.active ? "active" : ""}`
                  }
                >
                  <span>{item.icon}</span>
                  <span className="channelMeta">
                    <strong>{item.label}</strong>
                    <span>
                      {item.label === "广场"
                        ? "帖子流 / 推荐内容"
                        : item.label === "评论"
                          ? "评论列表 / 高级能力"
                        : item.label === "好友"
                          ? "管理联系人"
                          : item.label === "聊天框"
                            ? "会话与消息"
                            : item.label === "语音"
                              ? "LiveKit 面板"
                              : "服务配置"}
                    </span>
                  </span>
                  {item.label === "评论" ? <Badge tone="success">New</Badge> : null}
                </NavLink>
              ))}
            </div>
          </div>

          <div className="channelGroup">
            <div className="channelGroupHeader">
              <span>状态</span>
            </div>
            <div className="stack">
              <div className="statusRow">
                <span>IM</span>
                <Badge tone={status === "connected" ? "success" : "warning"}>{status}</Badge>
              </div>
              <div className="statusRow">
                <span>VDA</span>
                <Badge tone={vdaGrpcUrl ? "success" : "warning"}>{vdaGrpcUrl ? "ready" : "unset"}</Badge>
              </div>
              <div className="statusRow">
                <span>LiveKit</span>
                <Badge tone={livekitUrl ? "success" : "warning"}>{livekitUrl ? "ready" : "unset"}</Badge>
              </div>
              <div className="statusRow">
                <span>Call</span>
                <Badge tone={callState === "connected" ? "success" : callState === "ringing" ? "warning" : "neutral"}>
                  {callState}
                </Badge>
              </div>
            </div>
          </div>

          <div className="channelGroup">
            <div className="channelGroupHeader">
              <span>账户</span>
            </div>
            <div className="stack">
              {isLoggedIn() ? (
                <Button
                  variant="ghost"
                  onClick={() => {
                    logout();
                    disconnect();
                  }}
                >
                  退出登录
                </Button>
              ) : (
                <Link to="/login" className="button primary buttonLink">
                  登录
                </Link>
              )}
              <Link to="/" className="linkRow">
                <span>⌂</span>
                <span className="channelMeta">
                  <strong>首页</strong>
                  <span>返回广场</span>
                </span>
              </Link>
            </div>
          </div>
        </div>
      </aside>

      <main className="main">
        <div className="pageHero">
          <section className="heroCard">
            <div className="toolbar">
              <div className="titleBlock">
                <h2>{title}</h2>
                {subtitle ? <p>{subtitle}</p> : null}
              </div>
              <Badge tone="neutral">Discord style</Badge>
            </div>
          </section>
          {rightRail ? <div className="heroCard">{rightRail}</div> : null}
        </div>

        <div className="workspace">{children}</div>
      </main>
    </div>
  );
}
