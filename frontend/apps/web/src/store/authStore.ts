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
      baseUrl: "",
      token: "",
      userId: null,
      nickname: "",

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
    { name: "vda-auth", partialize: (s) => ({ baseUrl: s.baseUrl, userId: s.userId, nickname: s.nickname }) },
  ),
);
