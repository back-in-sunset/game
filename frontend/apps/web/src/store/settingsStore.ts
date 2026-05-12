import { create } from "zustand";
import { persist } from "zustand/middleware";

type SettingsState = {
  imWsUrl: string;
  vdaGrpcUrl: string;
  livekitUrl: string;

  setImWsUrl: (url: string) => void;
  setVdaGrpcUrl: (url: string) => void;
  setLivekitUrl: (url: string) => void;
};

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      imWsUrl: "ws://localhost:8082/ws",
      vdaGrpcUrl: "127.0.0.1:9101",
      livekitUrl: "http://localhost:7880",

      setImWsUrl: (imWsUrl) => set({ imWsUrl }),
      setVdaGrpcUrl: (vdaGrpcUrl) => set({ vdaGrpcUrl }),
      setLivekitUrl: (livekitUrl) => set({ livekitUrl }),
    }),
    { name: "vda-settings-v2" },
  ),
);
