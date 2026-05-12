import React from "react";
import ReactDOM from "react-dom/client";
import { App } from "./index";
import { useAuthStore } from "./store/authStore";
import { useSettingsStore } from "./store/settingsStore";
import "./styles.css";

// dev helper: set token from browser console
// usage: __setDevToken("eyJ...")
(window as any).__setDevToken = (t: string) => {
  useAuthStore.getState().login("http://localhost:8080", t);
  const imWsUrl = useSettingsStore.getState().imWsUrl;
  console.log("[dev] token set, len:", t.length, "imWsUrl:", imWsUrl);
  console.log("[dev] go to /login then back to / to connect, or just refresh");
};

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
