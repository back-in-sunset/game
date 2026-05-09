import { create } from "zustand";

type AuthState = {
  baseUrl: string;
  token: string;
  login: (baseUrl: string, token: string) => void;
  logout: () => void;
  isLoggedIn: () => boolean;
};

export const useAuthStore = create<AuthState>((set, get) => ({
  baseUrl: "",
  token: "",
  login: (baseUrl, token) => set({ baseUrl, token }),
  logout: () => set({ baseUrl: "", token: "" }),
  isLoggedIn: () => get().token.length > 0,
}));
