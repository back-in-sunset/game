import { describe, it, expect, vi, beforeEach } from "vitest";
import { HttpClient, HttpError } from "../http/client";

describe("HttpClient", () => {
  beforeEach(() => {
    vi.restoreAllMocks();
  });

  describe("get", () => {
    it("calls fetch with correct URL and headers", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response(JSON.stringify({ id: 1 }), { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080");
      const result = await client.get("/user/1");

      expect(result).toEqual({ id: 1 });
      expect(fetch).toHaveBeenCalledWith(
        new URL("http://localhost:8080/user/1"),
        { headers: { "Content-Type": "application/json" } },
      );
    });

    it("strips trailing slash from baseUrl", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response("{}", { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080/");
      await client.get("/path");

      expect(fetch).toHaveBeenCalledWith(
        new URL("http://localhost:8080/path"),
        expect.anything(),
      );
    });

    it("sends Authorization header when token is set", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response("{}", { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080", "my-jwt");
      await client.get("/secure");

      expect(fetch).toHaveBeenCalledWith(
        expect.anything(),
        {
          headers: {
            "Content-Type": "application/json",
            Authorization: "Bearer my-jwt",
          },
        },
      );
    });

    it("throws HttpError on non-ok response", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response("not found", { status: 404 }),
      );

      const client = new HttpClient("http://localhost:8080");
      let err: unknown;
      try {
        await client.get("/missing");
      } catch (e) {
        err = e;
      }
      expect(err).toBeInstanceOf(HttpError);
      expect((err as HttpError).status).toBe(404);
    });
  });

  describe("post", () => {
    it("sends JSON body with POST method", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response('{"ok":true}', { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080");
      const result = await client.post("/create", { name: "test" });

      expect(result).toEqual({ ok: true });
      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:8080/create",
        {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ name: "test" }),
        },
      );
    });
  });

  describe("delete", () => {
    it("sends DELETE request with correct method", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response("{}", { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080");
      await client.delete("/item/5");

      expect(fetch).toHaveBeenCalledWith(
        "http://localhost:8080/item/5",
        {
          method: "DELETE",
          headers: { "Content-Type": "application/json" },
        },
      );
    });
  });

  describe("setToken", () => {
    it("updates Authorization header on subsequent requests", async () => {
      vi.spyOn(globalThis, "fetch").mockResolvedValue(
        new Response("{}", { status: 200, headers: { "Content-Type": "application/json" } }),
      );

      const client = new HttpClient("http://localhost:8080");
      client.setToken("new-token");
      await client.get("/test");

      expect(fetch).toHaveBeenCalledWith(
        expect.anything(),
        {
          headers: {
            "Content-Type": "application/json",
            Authorization: "Bearer new-token",
          },
        },
      );
    });
  });
});

describe("HttpError", () => {
  it("is an Error with status", () => {
    const err = new HttpError(500, "server error");
    expect(err).toBeInstanceOf(Error);
    expect(err.name).toBe("HttpError");
    expect(err.status).toBe(500);
    expect(err.message).toBe("server error");
  });
});
