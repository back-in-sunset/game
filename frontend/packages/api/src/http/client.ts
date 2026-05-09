export class HttpClient {
  private baseUrl: string;
  private token: string;

  constructor(baseUrl: string, token = "") {
    this.baseUrl = baseUrl.replace(/\/$/, "");
    this.token = token;
  }

  setToken(token: string) {
    this.token = token;
  }

  async get<T>(path: string, params?: Record<string, string>): Promise<T> {
    const url = new URL(`${this.baseUrl}${path}`);
    if (params) {
      Object.entries(params).forEach(([k, v]) => url.searchParams.set(k, v));
    }
    return this.request<T>(url, { headers: this.headers() });
  }

  async post<T>(path: string, body: unknown): Promise<T> {
    return this.request<T>(`${this.baseUrl}${path}`, {
      method: "POST",
      headers: this.headers(),
      body: JSON.stringify(body),
    });
  }

  async delete<T>(path: string): Promise<T> {
    return this.request<T>(`${this.baseUrl}${path}`, {
      method: "DELETE",
      headers: this.headers(),
    });
  }

  private async request<T>(url: string | URL, init: RequestInit): Promise<T> {
    const res = await fetch(url, init);
    const text = await res.text();
    if (!res.ok) throw new HttpError(res.status, text);
    try { return JSON.parse(text); } catch { throw new HttpError(res.status, text); }
  }

  private headers(): Record<string, string> {
    const h: Record<string, string> = { "Content-Type": "application/json" };
    if (this.token) h["Authorization"] = `Bearer ${this.token}`;
    return h;
  }
}

export class HttpError extends Error {
  constructor(public status: number, message: string) {
    super(message);
    this.name = "HttpError";
  }
}
