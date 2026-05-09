export const serviceConfig = {
  imWsUrl: process.env.EXPO_PUBLIC_IM_WS_URL ?? "",
  vdaGrpcUrl: process.env.EXPO_PUBLIC_VDA_GRPC_URL ?? "",
  livekitUrl: process.env.EXPO_PUBLIC_LIVEKIT_URL ?? "",
  imToken: process.env.EXPO_PUBLIC_IM_TOKEN ?? "",
  imDomain: (process.env.EXPO_PUBLIC_IM_DOMAIN ?? "platform") as "platform" | "tenant",
  imTenantId: process.env.EXPO_PUBLIC_IM_TENANT_ID ?? "",
  imProjectId: process.env.EXPO_PUBLIC_IM_PROJECT_ID ?? "",
  imEnvironment: process.env.EXPO_PUBLIC_IM_ENVIRONMENT ?? "",
};
