import type { HttpClient } from "./client";

export type PlatformConfig = {
  tenantId: string;
  projectId: string;
  environment: string;
};

export type DemoTokenResponse = {
  accessToken: string;
  expiresIn: number;
  livekitUrl: string;
};

export type PlatformTenant = {
  id: string;
  name: string;
  slug: string;
};

export type PlatformProject = {
  id: string;
  tenantId: string;
  name: string;
  key: string;
};

export type PlatformEnvironment = {
  id: string;
  projectId: string;
  name: string;
  displayName: string;
};

export type CreateTenantRequest = {
  name: string;
  slug: string;
};

export type UpdateTenantRequest = {
  tenantId: string;
  name: string;
  slug: string;
};

export type CreateProjectRequest = {
  tenantId: string;
  name: string;
  key: string;
};

export type CreateEnvironmentRequest = {
  projectId: string;
  name: string;
  displayName: string;
};

export type MyTenantsResponse = {
  items: PlatformTenant[];
};

export type ListProjectsResponse = {
  items: PlatformProject[];
};

export type ListEnvironmentsResponse = {
  items: PlatformEnvironment[];
};

export function createPlatformAPI(client: HttpClient) {
  return {
    getDemoToken: () => client.post<DemoTokenResponse>("/api/v1/demo/token", {}),
    createTenant: (payload: CreateTenantRequest) => client.post<PlatformTenant>("/api/platform/tenant/create", payload),
    getTenant: (tenantId: string) => client.get<PlatformTenant>(`/api/platform/tenant/${tenantId}`),
    updateTenant: (payload: UpdateTenantRequest) => client.post<PlatformTenant>("/api/platform/tenant/update", payload),
    deleteTenant: (tenantId: string) => client.post<{ success: boolean }>("/api/platform/tenant/delete", { tenantId }),
    myTenants: () => client.get<MyTenantsResponse>("/api/platform/tenant/my"),
    createProject: (payload: CreateProjectRequest) => client.post<PlatformProject>("/api/platform/project/create", payload),
    listProjects: (tenantId: string) =>
      client.get<ListProjectsResponse>("/api/platform/project/list", { tenantId }),
    createEnvironment: (payload: CreateEnvironmentRequest) =>
      client.post<PlatformEnvironment>("/api/platform/environment/create", payload),
    listEnvironments: (projectId: string) =>
      client.get<ListEnvironmentsResponse>("/api/platform/environment/list", { projectId }),
  };
}

export type PlatformAPI = ReturnType<typeof createPlatformAPI>;
