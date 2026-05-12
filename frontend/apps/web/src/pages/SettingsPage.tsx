import { Link } from "react-router-dom";
import { Card, StatusRow } from "@game/ui";
import { useSettingsStore } from "../store/settingsStore";
import { WebShell } from "../components/WebShell";

export function SettingsPage() {
  const { imWsUrl, vdaGrpcUrl, livekitUrl, setImWsUrl, setVdaGrpcUrl, setLivekitUrl } =
    useSettingsStore();

  return (
    <WebShell title="设置" subtitle="IM / VDA / LiveKit 服务配置" activeTab="square">
      <section className="feed">
        <Card title="服务配置">
          <div className="stack">
            <label style={{ display: "block" }}>
              <span style={{ display: "block", marginBottom: 4, fontSize: 13, color: "#667085" }}>
                IM WebSocket URL
              </span>
              <input
                className="searchInput"
                value={imWsUrl}
                onChange={(e) => setImWsUrl(e.target.value)}
                placeholder="ws://localhost:8082/ws"
              />
            </label>
            <label style={{ display: "block" }}>
              <span style={{ display: "block", marginBottom: 4, fontSize: 13, color: "#667085" }}>
                VDA gRPC URL
              </span>
              <input
                className="searchInput"
                value={vdaGrpcUrl}
                onChange={(e) => setVdaGrpcUrl(e.target.value)}
                placeholder="localhost:9101"
              />
            </label>
            <label style={{ display: "block" }}>
              <span style={{ display: "block", marginBottom: 4, fontSize: 13, color: "#667085" }}>
                LiveKit URL
              </span>
              <input
                className="searchInput"
                value={livekitUrl}
                onChange={(e) => setLivekitUrl(e.target.value)}
                placeholder="http://localhost:7880"
              />
            </label>
          </div>
          <div style={{ marginTop: 16 }} className="stack">
            <StatusRow
              label="IM WebSocket"
              value={imWsUrl || "未设置"}
              tone={imWsUrl ? "success" : "warning"}
            />
            <StatusRow
              label="VDA gRPC"
              value={vdaGrpcUrl || "未设置"}
              tone={vdaGrpcUrl ? "success" : "warning"}
            />
            <StatusRow
              label="LiveKit"
              value={livekitUrl || "未设置"}
              tone={livekitUrl ? "success" : "warning"}
            />
          </div>
        </Card>
      </section>
      <aside className="stack">
        <Card title="路径" subtitle="返回主界面">
          <Link to="/" className="linkRow">
            <span>◎</span>
            <span className="channelMeta">
              <strong>回到广场</strong>
              <span>帖子流首页</span>
            </span>
          </Link>
        </Card>
      </aside>
    </WebShell>
  );
}
