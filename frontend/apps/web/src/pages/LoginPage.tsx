import { useState } from "react";
import { useNavigate, Navigate } from "react-router-dom";
import { Card, Button } from "@game/ui";
import { useAuthStore } from "../store/authStore";
import { WebShell } from "../components/WebShell";

export function LoginPage() {
  const { login, isLoggedIn } = useAuthStore();
  const navigate = useNavigate();
  const [baseUrl, setBaseUrl] = useState("http://localhost:8080");
  const [token, setToken] = useState("");
  const [error, setError] = useState("");

  if (isLoggedIn()) return <Navigate to="/" replace />;

  function doLogin() {
    if (!baseUrl.trim() || !token.trim()) {
      setError("请填写 API 地址和 Token");
      return;
    }
    try {
      login(baseUrl.trim(), token.trim());
      navigate("/", { replace: true });
    } catch {
      setError("登录失败");
    }
  }

  return (
    <WebShell title="登录" subtitle="进入广场、好友管理和聊天框" activeTab="square">
      <section className="feed" style={{ maxWidth: 520 }}>
        <Card title="登录" subtitle="输入 API 地址和 JWT Token">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              doLogin();
            }}
            className="stack"
          >
            <div>
              <label style={{ display: "block", marginBottom: 4, fontSize: 13, color: "#667085" }}>API 地址</label>
              <input
                className="searchInput"
                value={baseUrl}
                onChange={(e) => setBaseUrl(e.target.value)}
                placeholder="http://localhost:8080"
              />
            </div>
            <div>
              <label style={{ display: "block", marginBottom: 4, fontSize: 13, color: "#667085" }}>JWT Token</label>
              <input
                className="searchInput"
                value={token}
                onChange={(e) => setToken(e.target.value)}
                placeholder="输入平台 Token"
                type="password"
              />
            </div>
            {error ? <p className="errorText">{error}</p> : null}
            <Button variant="primary" onClick={doLogin}>
              登录
            </Button>
          </form>
        </Card>
      </section>
    </WebShell>
  );
}
