import { BrowserRouter, Routes, Route } from "react-router-dom";
import { DashboardPage } from "../pages/Dashboard";
import { ChatPage } from "../pages/ChatPage";
import { VoicePage } from "../pages/VoicePage";
import { FriendsPage } from "../pages/FriendsPage";
import { HistoryPage } from "../pages/HistoryPage";
import { SettingsPage } from "../pages/SettingsPage";
import { CommentPage } from "../pages/CommentPage";
import { LoginPage } from "../pages/LoginPage";
import { DemoPage } from "../pages/DemoPage";
import { DemoRoomPage } from "../pages/DemoRoomPage";
import { useIMBootstrap } from "../hooks/useIMBootstrap";

export function App() {
  return (
    <BrowserRouter>
      <AppRoutes />
    </BrowserRouter>
  );
}

function AppRoutes() {
  useIMBootstrap();

  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/" element={<DashboardPage />} />
      <Route path="/chat/:userId" element={<ChatPage />} />
      <Route path="/voice" element={<VoicePage />} />
      <Route path="/friends" element={<FriendsPage />} />
      <Route path="/comments" element={<CommentPage />} />
      <Route path="/history" element={<HistoryPage />} />
      <Route path="/settings" element={<SettingsPage />} />
      <Route path="/demo" element={<DemoPage />} />
      <Route path="/demo/room/:roomId" element={<DemoRoomPage />} />
    </Routes>
  );
}
