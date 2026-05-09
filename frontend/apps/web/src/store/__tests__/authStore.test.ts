import { describe, it, expect, beforeEach } from "vitest";
import { useAuthStore } from "../authStore";

function reset() {
  useAuthStore.setState({
    baseUrl: "",
    token: "",
    userId: null,
    nickname: "",
  });
}

describe("authStore", () => {
  beforeEach(() => {
    reset();
  });

  describe("login", () => {
    it("sets baseUrl and token", () => {
      useAuthStore.getState().login("http://example.com", "jwt123");
      const s = useAuthStore.getState();
      expect(s.baseUrl).toBe("http://example.com");
      expect(s.token).toBe("jwt123");
    });
  });

  describe("logout", () => {
    it("clears auth state", () => {
      useAuthStore.getState().login("http://example.com", "jwt123");
      useAuthStore.getState().setUserId(42);
      useAuthStore.getState().setNickname("Alice");

      useAuthStore.getState().logout();

      const s = useAuthStore.getState();
      expect(s.baseUrl).toBe("");
      expect(s.token).toBe("");
      expect(s.userId).toBeNull();
      expect(s.nickname).toBe("");
    });
  });

  describe("isLoggedIn", () => {
    it("returns false when token is empty", () => {
      expect(useAuthStore.getState().isLoggedIn()).toBe(false);
    });

    it("returns true when token is set", () => {
      useAuthStore.getState().login("http://example.com", "has-value");
      expect(useAuthStore.getState().isLoggedIn()).toBe(true);
    });
  });

  describe("setUserId / setNickname", () => {
    it("updates userId and nickname independently", () => {
      useAuthStore.getState().setUserId(99);
      useAuthStore.getState().setNickname("Bob");
      const s = useAuthStore.getState();
      expect(s.userId).toBe(99);
      expect(s.nickname).toBe("Bob");
    });
  });
});
