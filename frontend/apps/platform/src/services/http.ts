import { HttpClient, createPlatformAPI, type PlatformAPI } from "@game/api";
import { usePlatformStore } from "../store/platformStore";

let cacheKey = "";
let cachedClient: HttpClient | null = null;
let cachedPlatformApi: PlatformAPI | null = null;

function getCacheKey(baseUrl: string, token: string): string {
  return `${baseUrl}::${token}`;
}

export function getHttpClient(): HttpClient {
  const { baseUrl, token } = usePlatformStore.getState();
  const key = getCacheKey(baseUrl, token);
  if (!cachedClient || cacheKey !== key) {
    cachedClient = new HttpClient(baseUrl, token);
    cachedPlatformApi = createPlatformAPI(cachedClient);
    cacheKey = key;
  }
  return cachedClient;
}

export function getPlatformAPI(): PlatformAPI {
  getHttpClient();
  if (!cachedPlatformApi) {
    throw new Error("platform api not initialized");
  }
  return cachedPlatformApi;
}

export function resetPlatformHTTP(): void {
  cacheKey = "";
  cachedClient = null;
  cachedPlatformApi = null;
}
