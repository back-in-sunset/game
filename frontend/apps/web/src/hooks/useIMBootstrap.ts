import { useEffect } from "react";
import { conversations } from "@game/shared";
import { useAuthStore } from "../store/authStore";
import { useIMStore } from "../store/imStore";
import { useSettingsStore } from "../store/settingsStore";

export function useIMBootstrap() {
  const { token } = useAuthStore();
  const { imWsUrl } = useSettingsStore();
  const { connect, disconnect, loadConversations, setActiveConversation } = useIMStore();

  useEffect(() => {
    loadConversations(conversations);
    if (conversations[0]) {
      setActiveConversation(conversations[0].id);
    }
  }, [loadConversations, setActiveConversation]);

  useEffect(() => {
    if (!imWsUrl || !token) return;
    connect(imWsUrl, {
      token,
      domain: "platform",
      scope: { tenant_id: "", project_id: "", environment: "prod" },
    });
    return () => disconnect();
  }, [connect, disconnect, imWsUrl, token]);
}
