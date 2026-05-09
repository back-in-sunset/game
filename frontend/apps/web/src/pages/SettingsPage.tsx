import { Link } from "react-router-dom";
import { Card, StatusRow } from "@game/ui";
import { serviceConfig } from "../config";

export function SettingsPage() {
  return (
    <div className="shell">
      <aside className="sidebar">
        <div className="brand">
          <div className="brandMark">V</div>
          <div><h1>设置</h1></div>
        </div>
        <Link to="/">← 返回</Link>
      </aside>
      <main className="main">
        <Card title="服务配置">
          <div className="stack">
            <StatusRow label="IM WebSocket" value={serviceConfig.imWsUrl || "unset"} tone={serviceConfig.imWsUrl ? "success" : "warning"} />
            <StatusRow label="VDA gRPC" value={serviceConfig.vdaGrpcUrl || "unset"} tone={serviceConfig.vdaGrpcUrl ? "success" : "warning"} />
            <StatusRow label="LiveKit" value={serviceConfig.livekitUrl || "unset"} tone={serviceConfig.livekitUrl ? "success" : "warning"} />
            <StatusRow label="Domain" value={serviceConfig.imDomain} tone="success" />
            <StatusRow label="Environment" value={serviceConfig.imEnvironment || "default"} tone="success" />
          </div>
        </Card>
      </main>
    </div>
  );
}
