import type { HttpClient } from "./client";

export type PlatformConfig = {
  tenantId: string;
  projectId: string;
  environment: string;
};

export function createPlatformAPI(client: HttpClient) {
  return {
    getJWT: () => client.get<{ token: string }>("/platform/token"),
    getConfig: () => client.get<PlatformConfig>("/platform/config"),
  };
}

export type PlatformAPI = ReturnType<typeof createPlatformAPI>;
