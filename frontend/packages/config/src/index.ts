export type ServiceConfig = {
  imWsUrl: string;
  vdaGrpcUrl: string;
  livekitUrl: string;
  imToken: string;
  imDomain: "platform" | "tenant";
  imTenantId: string;
  imProjectId: string;
  imEnvironment: string;
};

export const defaultServiceConfig: ServiceConfig = {
  imWsUrl: "",
  vdaGrpcUrl: "",
  livekitUrl: "",
  imToken: "",
  imDomain: "platform",
  imTenantId: "",
  imProjectId: "",
  imEnvironment: "",
};

export function normalizeServiceConfig(input: Partial<ServiceConfig> = {}): ServiceConfig {
  return {
    ...defaultServiceConfig,
    ...input,
  };
}

