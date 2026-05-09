import { useState } from "react";
import { HomeScreen } from "./src/screens/HomeScreen";
import { ChatScreen } from "./src/screens/ChatScreen";
import { VoiceScreen } from "./src/screens/VoiceScreen";
import { SettingsScreen } from "./src/screens/SettingsScreen";

type Screen =
  | { name: "home" }
  | { name: "chat"; params: { userId: number; title: string } }
  | { name: "voice" }
  | { name: "settings" };

export default function App() {
  const [screen, setScreen] = useState<Screen>({ name: "home" });

  function navigate(name: string, params?: Record<string, unknown>) {
    if (name === "chat") {
      setScreen({ name, params: params as { userId: number; title: string } });
    } else if (name === "voice") {
      setScreen({ name });
    } else if (name === "settings") {
      setScreen({ name });
    }
  }

  function goBack() {
    setScreen({ name: "home" });
  }

  switch (screen.name) {
    case "chat":
      return (
        <ChatScreen
          userId={screen.params.userId}
          title={screen.params.title}
          onBack={goBack}
        />
      );
    case "voice":
      return <VoiceScreen onBack={goBack} />;
    case "settings":
      return <SettingsScreen onBack={goBack} />;
    default:
      return <HomeScreen navigate={navigate} />;
  }
}
