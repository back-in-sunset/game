import type { ServiceConfig } from "@game/config";

export type { ServiceConfig };

export const serviceConfig: ServiceConfig = {
  imWsUrl: import.meta.env.VITE_IM_WS_URL ?? "",
  vdaGrpcUrl: import.meta.env.VITE_VDA_GRPC_URL ?? "",
  livekitUrl: import.meta.env.VITE_LIVEKIT_URL ?? "",
  imToken: import.meta.env.VITE_IM_TOKEN ?? "",
  imDomain: (import.meta.env.VITE_IM_DOMAIN ?? "platform") as "platform" | "tenant",
  imTenantId: import.meta.env.VITE_IM_TENANT_ID ?? "",
  imProjectId: import.meta.env.VITE_IM_PROJECT_ID ?? "",
  imEnvironment: import.meta.env.VITE_IM_ENVIRONMENT ?? "",
};

