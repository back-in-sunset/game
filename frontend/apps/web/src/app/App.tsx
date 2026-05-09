import { BrowserRouter, Routes, Route } from "react-router-dom";
import { DashboardPage } from "../pages/Dashboard";
import { ChatPage } from "../pages/ChatPage";
import { VoicePage } from "../pages/VoicePage";
import { FriendsPage } from "../pages/FriendsPage";
import { HistoryPage } from "../pages/HistoryPage";
import { SettingsPage } from "../pages/SettingsPage";
import { LoginPage } from "../pages/LoginPage";

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/" element={<DashboardPage />} />
        <Route path="/chat/:userId" element={<ChatPage />} />
        <Route path="/voice" element={<VoicePage />} />
        <Route path="/friends" element={<FriendsPage />} />
        <Route path="/history" element={<HistoryPage />} />
        <Route path="/settings" element={<SettingsPage />} />
      </Routes>
    </BrowserRouter>
  );
}
