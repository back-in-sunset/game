import { create } from "zustand";
import { persist } from "zustand/middleware";

type PlatformState = {
  baseUrl: string;
  token: string;
  selectedTenantId: string;
  selectedProjectId: string;

  login: (baseUrl: string, token: string) => void;
  logout: () => void;
  setBaseUrl: (baseUrl: string) => void;
  setToken: (token: string) => void;
  setSelectedTenantId: (tenantId: string) => void;
  setSelectedProjectId: (projectId: string) => void;
  isConnected: () => boolean;
};

export const usePlatformStore = create<PlatformState>()(
  persist(
    (set, get) => ({
      baseUrl: "http://localhost:8080",
      token: "",
      selectedTenantId: "",
      selectedProjectId: "",

      login: (baseUrl, token) =>
        set({
          baseUrl: baseUrl.trim() || "http://localhost:8080",
          token: token.trim(),
          selectedTenantId: "",
          selectedProjectId: "",
        }),

      logout: () =>
        set({
          token: "",
          selectedTenantId: "",
          selectedProjectId: "",
        }),

      setBaseUrl: (baseUrl) => set({ baseUrl }),
      setToken: (token) => set({ token }),
      setSelectedTenantId: (selectedTenantId) =>
        set((state) => ({
          selectedTenantId,
          selectedProjectId: state.selectedTenantId === selectedTenantId ? state.selectedProjectId : "",
        })),
      setSelectedProjectId: (selectedProjectId) => set({ selectedProjectId }),
      isConnected: () => get().token.trim().length > 0,
    }),
    {
      name: "game-platform-auth-v1",
      partialize: (state) => ({
        baseUrl: state.baseUrl,
        token: state.token,
        selectedTenantId: state.selectedTenantId,
        selectedProjectId: state.selectedProjectId,
      }),
    },
  ),
);
