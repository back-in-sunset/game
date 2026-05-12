import { create } from "zustand";
import { persist } from "zustand/middleware";
import { resetAPIs } from "../services/http";

type AuthState = {
  baseUrl: string;
  token: string;
  userId: number | null;
  nickname: string;

  login: (baseUrl: string, token: string) => void;
  logout: () => void;
  setUserId: (id: number) => void;
  setNickname: (name: string) => void;
  isLoggedIn: () => boolean;
};

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      baseUrl: "http://localhost:8080",
      token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3ODEwNzk1MDUsImlhdCI6MTc3ODQ4NzUwNSwidWlkIjoxfQ.33SwLTGOs9TFGLN7NyC7TtiNbsGtAJYqKU5QcPBNifHM3UrChj6NJ8AurvuEEnY7qTZZsIozZLmOdgVWXOsMRVzh9R-wfxGVGDXXrFB8YCmqZjkDaaeHkBes0GvXt_P34SSLYYwpubCB9KVKbujeg4oFLnhTKEin_FomzLWeq_3qtis3I9zBl-j3qZbf0GgCzNosMhvtfzKWqMjXcUkpYR-SAmT-9i-U5unJCEEe4L5Y8aur1pBWHowhkxpxHNo-MLQWpreqAYdVosxvzH0r-cgSmSETSorgDegjZF6GeEdVA7bPCMRGaiU4T9Sg1hVYrvla41VoEFvE_1EHsp_6LQ",
      userId: 1,
      nickname: "dev",

      login: (baseUrl, token) => {
        resetAPIs();
        set({ baseUrl, token });
      },

      logout: () => {
        resetAPIs();
        set({ baseUrl: "", token: "", userId: null, nickname: "" });
      },

      setUserId: (userId) => set({ userId }),
      setNickname: (nickname) => set({ nickname }),

      isLoggedIn: () => get().token.length > 0,
    }),
    { name: "vda-auth-v2", partialize: (s) => ({ baseUrl: s.baseUrl, token: s.token, userId: s.userId, nickname: s.nickname }) },
  ),
);
