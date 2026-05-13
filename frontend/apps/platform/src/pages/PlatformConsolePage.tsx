import { useCallback, useEffect, useMemo, useState } from "react";
import type { PlatformEnvironment, PlatformProject, PlatformTenant } from "@game/api";
import { Badge, Button, Card, EmptyState, Input, Label, Separator } from "../components/ui";
import { extractDemoJwt } from "../lib/token";
import { isBlank, slugify } from "../lib/slug";
import { getPlatformAPI } from "../services/http";
import { usePlatformStore } from "../store/platformStore";

type Notice = {
  tone: "success" | "warning" | "danger" | "info";
  text: string;
} | null;

function mergeById<T extends { id: string }>(primary: T[], fallback: T[]): T[] {
  const seen = new Set<string>();
  const items: T[] = [];

  for (const item of primary) {
    if (seen.has(item.id)) continue;
    seen.add(item.id);
    items.push(item);
  }

  for (const item of fallback) {
    if (seen.has(item.id)) continue;
    seen.add(item.id);
    items.push(item);
  }

  return items;
}

function formatCount(value: number): string {
  return value > 9 ? "9+" : String(value);
}

export function PlatformConsolePage() {
  const {
    baseUrl,
    token,
    selectedTenantId,
    selectedProjectId,
    login,
    logout,
    setSelectedTenantId,
    setSelectedProjectId,
  } = usePlatformStore();

  const [baseUrlDraft, setBaseUrlDraft] = useState(baseUrl);
  const [tokenDraft, setTokenDraft] = useState(token);
  const [notice, setNotice] = useState<Notice>(null);
  const [tenantName, setTenantName] = useState("");
  const [tenantSlug, setTenantSlug] = useState("");
  const [tenantEditName, setTenantEditName] = useState("");
  const [tenantEditSlug, setTenantEditSlug] = useState("");
  const [tenantLookupId, setTenantLookupId] = useState("");
  const [projectTenantId, setProjectTenantId] = useState(selectedTenantId);
  const [projectName, setProjectName] = useState("");
  const [projectKey, setProjectKey] = useState("");
  const [environmentProjectId, setEnvironmentProjectId] = useState(selectedProjectId);
  const [environmentName, setEnvironmentName] = useState("dev");
  const [environmentDisplayName, setEnvironmentDisplayName] = useState("Development");
  const [tenantLoading, setTenantLoading] = useState(false);
  const [projectLoading, setProjectLoading] = useState(false);
  const [environmentLoading, setEnvironmentLoading] = useState(false);
  const [tenants, setTenants] = useState<PlatformTenant[]>([]);
  const [projectCache, setProjectCache] = useState<Record<string, PlatformProject[]>>({});
  const [environmentCache, setEnvironmentCache] = useState<Record<string, PlatformEnvironment[]>>({});

  useEffect(() => setBaseUrlDraft(baseUrl), [baseUrl]);
  useEffect(() => setTokenDraft(token), [token]);
  useEffect(() => setProjectTenantId(selectedTenantId), [selectedTenantId]);
  useEffect(() => setEnvironmentProjectId(selectedProjectId), [selectedProjectId]);

  const selectedTenant = useMemo(
    () => tenants.find((tenant) => tenant.id === selectedTenantId) ?? null,
    [selectedTenantId, tenants],
  );
  const selectedProjects = projectCache[selectedTenantId] ?? [];
  const selectedProject = useMemo(
    () => selectedProjects.find((project) => project.id === selectedProjectId) ?? null,
    [selectedProjectId, selectedProjects],
  );
  const selectedEnvironments = environmentCache[selectedProjectId] ?? [];

  const refreshTenants = useCallback(async () => {
    if (!token.trim()) {
      return;
    }
    setTenantLoading(true);
    try {
      const resp = await getPlatformAPI().myTenants();
      setTenants((current) => mergeById(resp.items, current));
      setNotice({ tone: "success", text: `已加载 ${resp.items.length} 个我的租户` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "加载租户失败";
      setNotice({ tone: "danger", text: message });
    } finally {
      setTenantLoading(false);
    }
  }, [token]);

  const refreshProjects = useCallback(async (tenantId: string) => {
    if (isBlank(tenantId)) {
      return;
    }
    setProjectLoading(true);
    try {
      const resp = await getPlatformAPI().listProjects(tenantId.trim());
      setProjectCache((current) => ({
        ...current,
        [tenantId]: mergeById(resp.items, current[tenantId] ?? []),
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : "加载项目失败";
      setNotice({ tone: "danger", text: message });
    } finally {
      setProjectLoading(false);
    }
  }, []);

  const refreshEnvironments = useCallback(async (projectId: string) => {
    if (isBlank(projectId)) {
      return;
    }
    setEnvironmentLoading(true);
    try {
      const resp = await getPlatformAPI().listEnvironments(projectId.trim());
      setEnvironmentCache((current) => ({
        ...current,
        [projectId]: mergeById(resp.items, current[projectId] ?? []),
      }));
    } catch (error) {
      const message = error instanceof Error ? error.message : "加载环境失败";
      setNotice({ tone: "danger", text: message });
    } finally {
      setEnvironmentLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!token.trim()) {
      setTenants([]);
      setProjectCache({});
      setEnvironmentCache({});
      setSelectedTenantId("");
      setSelectedProjectId("");
      return;
    }
    void refreshTenants();
  }, [refreshTenants, setSelectedProjectId, setSelectedTenantId, token]);

  useEffect(() => {
    if (selectedTenantId.trim()) {
      void refreshProjects(selectedTenantId);
    }
  }, [refreshProjects, selectedTenantId]);

  useEffect(() => {
    if (selectedProjectId.trim()) {
      void refreshEnvironments(selectedProjectId);
    }
  }, [refreshEnvironments, selectedProjectId]);

  useEffect(() => {
    if (!selectedTenantId && tenants[0]) {
      setSelectedTenantId(tenants[0].id);
    }
  }, [selectedTenantId, setSelectedTenantId, tenants]);

  useEffect(() => {
    if (selectedTenant) {
      setTenantEditName(selectedTenant.name);
      setTenantEditSlug(selectedTenant.slug);
      setTenantLookupId(selectedTenant.id);
    }
  }, [selectedTenant]);

  useEffect(() => {
    if (selectedTenantId && !projectTenantId) {
      setProjectTenantId(selectedTenantId);
    }
  }, [projectTenantId, selectedTenantId]);

  useEffect(() => {
    if (selectedProjectId && !environmentProjectId) {
      setEnvironmentProjectId(selectedProjectId);
    }
  }, [environmentProjectId, selectedProjectId]);

  const handleConnect = async () => {
    const nextBaseUrl = baseUrlDraft.trim() || "http://localhost:8080";
    const nextToken = extractDemoJwt(tokenDraft.trim());
    if (!nextToken) {
      setNotice({ tone: "warning", text: "请输入 JWT，或点击获取演示令牌" });
      return;
    }
    login(nextBaseUrl, nextToken);
    setNotice({ tone: "success", text: "平台已连接" });
  };

  const handleFetchDemoToken = async () => {
    const nextBaseUrl = baseUrlDraft.trim() || "http://localhost:8080";
    setNotice(null);
    try {
      login(nextBaseUrl, "");
      const resp = await getPlatformAPI().getDemoToken();
      const jwt = extractDemoJwt(resp.accessToken);
      login(nextBaseUrl, jwt);
      setNotice({ tone: "success", text: `已获取演示令牌，有效期 ${Math.floor(resp.expiresIn / 60)} 分钟` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "获取演示令牌失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleCreateTenant = async () => {
    const name = tenantName.trim();
    const slug = tenantSlug.trim() || slugify(name);
    if (isBlank(name) || isBlank(slug)) {
      setNotice({ tone: "warning", text: "租户名称和 slug 不能为空" });
      return;
    }
    try {
      const resp = await getPlatformAPI().createTenant({ name, slug });
      setTenants((current) => mergeById([resp], current));
      setSelectedTenantId(resp.id);
      setTenantName("");
      setTenantSlug("");
      setTenantEditName(resp.name);
      setTenantEditSlug(resp.slug);
      setNotice({ tone: "success", text: `租户 ${resp.name} 创建成功` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "创建租户失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleLookupTenant = async () => {
    const tenantId = tenantLookupId.trim();
    if (isBlank(tenantId)) {
      setNotice({ tone: "warning", text: "请输入 tenantId" });
      return;
    }
    try {
      const resp = await getPlatformAPI().getTenant(tenantId);
      setTenants((current) => mergeById([resp], current));
      setSelectedTenantId(resp.id);
      setNotice({ tone: "success", text: `已加载租户 ${resp.name}` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "加载租户失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleUpdateTenant = async () => {
    if (isBlank(selectedTenantId)) {
      setNotice({ tone: "warning", text: "先选择一个租户" });
      return;
    }
    const name = tenantEditName.trim();
    const slug = tenantEditSlug.trim() || slugify(name);
    if (isBlank(name) || isBlank(slug)) {
      setNotice({ tone: "warning", text: "租户名称和 slug 不能为空" });
      return;
    }
    try {
      const resp = await getPlatformAPI().updateTenant({
        tenantId: selectedTenantId,
        name,
        slug,
      });
      setTenants((current) => current.map((tenant) => (tenant.id === resp.id ? resp : tenant)));
      setTenantEditName(resp.name);
      setTenantEditSlug(resp.slug);
      setNotice({ tone: "success", text: `租户 ${resp.name} 已更新` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "更新租户失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleDeleteTenant = async () => {
    if (isBlank(selectedTenantId)) {
      setNotice({ tone: "warning", text: "先选择一个租户" });
      return;
    }
    const tenantNameLabel = selectedTenant?.name ?? selectedTenantId;
    if (!window.confirm(`确认删除租户「${tenantNameLabel}」？`)) {
      return;
    }
    try {
      await getPlatformAPI().deleteTenant(selectedTenantId);
      setTenants((current) => current.filter((tenant) => tenant.id !== selectedTenantId));
      setProjectCache((current) => {
        const next = { ...current };
        delete next[selectedTenantId];
        return next;
      });
      setEnvironmentCache({});
      setSelectedTenantId("");
      setSelectedProjectId("");
      setNotice({ tone: "success", text: `租户 ${tenantNameLabel} 已删除` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "删除租户失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleCreateProject = async () => {
    const tenantId = projectTenantId.trim() || selectedTenantId.trim();
    const name = projectName.trim();
    const key = projectKey.trim() || slugify(name);
    if (isBlank(tenantId) || isBlank(name) || isBlank(key)) {
      setNotice({ tone: "warning", text: "项目所属租户、名称和 key 不能为空" });
      return;
    }
    try {
      const resp = await getPlatformAPI().createProject({ tenantId, name, key });
      setProjectCache((current) => ({
        ...current,
        [tenantId]: mergeById([resp], current[tenantId] ?? []),
      }));
      setSelectedTenantId(tenantId);
      setSelectedProjectId(resp.id);
      setProjectName("");
      setProjectKey("");
      setNotice({ tone: "success", text: `项目 ${resp.name} 创建成功` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "创建项目失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleCreateEnvironment = async () => {
    const projectId = environmentProjectId.trim() || selectedProjectId.trim();
    const name = environmentName.trim();
    const displayName = environmentDisplayName.trim() || name;
    if (isBlank(projectId) || isBlank(name)) {
      setNotice({ tone: "warning", text: "环境所属项目和名称不能为空" });
      return;
    }
    try {
      const resp = await getPlatformAPI().createEnvironment({
        projectId,
        name,
        displayName,
      });
      setEnvironmentCache((current) => ({
        ...current,
        [projectId]: mergeById([resp], current[projectId] ?? []),
      }));
      setSelectedProjectId(projectId);
      setEnvironmentName("dev");
      setEnvironmentDisplayName("Development");
      setNotice({ tone: "success", text: `环境 ${resp.name} 创建成功` });
    } catch (error) {
      const message = error instanceof Error ? error.message : "创建环境失败";
      setNotice({ tone: "danger", text: message });
    }
  };

  const handleTenantSelect = (tenant: PlatformTenant) => {
    setSelectedTenantId(tenant.id);
    setTenantEditName(tenant.name);
    setTenantEditSlug(tenant.slug);
    setSelectedProjectId("");
  };

  const handleProjectSelect = (project: PlatformProject) => {
    setSelectedTenantId(project.tenantId);
    setSelectedProjectId(project.id);
  };

  const platformState = token ? "connected" : "offline";

  return (
    <div className="min-h-screen px-4 py-4 text-slate-900 sm:px-6 lg:px-8">
      <div className="mx-auto flex max-w-[1600px] flex-col gap-6">
        <header className="overflow-hidden rounded-[2rem] border border-slate-200/80 bg-white/75 shadow-[0_24px_80px_rgba(15,23,42,0.08)] backdrop-blur">
          <div className="grid gap-4 p-6 lg:grid-cols-[minmax(0,1.2fr)_minmax(320px,0.8fr)] lg:p-8">
            <div className="space-y-4">
              <div className="inline-flex items-center gap-2 rounded-full border border-teal-200 bg-teal-50 px-3 py-1 text-xs font-medium text-teal-700">
                Platform console
                <Badge tone={platformState === "connected" ? "success" : "warning"}>
                  {platformState}
                </Badge>
              </div>
              <div className="space-y-2">
                <h1 className="text-3xl font-semibold tracking-tight text-slate-950 sm:text-4xl">
                  前端平台管理
                </h1>
                <p className="max-w-2xl text-sm leading-6 text-slate-500 sm:text-base">
                  管理 tenant / project / environment 三层资源，直接连接后端平台 API。
                </p>
              </div>
              <div className="flex flex-wrap gap-2">
                <Badge tone="info">{formatCount(tenants.length)} tenants</Badge>
                <Badge tone="neutral">{formatCount(selectedProjects.length)} projects</Badge>
                <Badge tone="neutral">{formatCount(selectedEnvironments.length)} environments</Badge>
              </div>
            </div>

            <Card title="连接平台" description="支持手动 JWT 或演示令牌" className="bg-white/90">
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="baseUrl">API 地址</Label>
                  <Input
                    id="baseUrl"
                    value={baseUrlDraft}
                    onChange={(event) => setBaseUrlDraft(event.target.value)}
                    placeholder="http://localhost:8080"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="token">JWT / demo token</Label>
                  <Input
                    id="token"
                    value={tokenDraft}
                    onChange={(event) => setTokenDraft(event.target.value)}
                    placeholder="粘贴 JWT，或点击获取演示令牌"
                  />
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button variant="primary" onClick={handleConnect}>
                    连接平台
                  </Button>
                  <Button variant="secondary" onClick={handleFetchDemoToken}>
                    获取演示令牌
                  </Button>
                  <Button variant="ghost" onClick={() => logout()}>
                    断开
                  </Button>
                </div>
                <Separator />
                <div className="grid gap-2 text-sm text-slate-600">
                  <div className="flex items-center justify-between gap-3">
                    <span>当前状态</span>
                    <Badge tone={token ? "success" : "warning"}>{token ? "已连接" : "未连接"}</Badge>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <span>选中租户</span>
                    <span className="truncate font-medium text-slate-900">
                      {selectedTenant?.name ?? "未选择"}
                    </span>
                  </div>
                  <div className="flex items-center justify-between gap-3">
                    <span>选中项目</span>
                    <span className="truncate font-medium text-slate-900">
                      {selectedProject?.name ?? "未选择"}
                    </span>
                  </div>
                </div>
              </div>
            </Card>
          </div>

          {notice ? (
            <div
              className={
                notice.tone === "success"
                  ? "border-t border-emerald-200 bg-emerald-50 px-6 py-3 text-sm text-emerald-800"
                  : notice.tone === "warning"
                    ? "border-t border-amber-200 bg-amber-50 px-6 py-3 text-sm text-amber-800"
                    : notice.tone === "info"
                      ? "border-t border-cyan-200 bg-cyan-50 px-6 py-3 text-sm text-cyan-800"
                      : "border-t border-rose-200 bg-rose-50 px-6 py-3 text-sm text-rose-800"
              }
            >
              {notice.text}
            </div>
          ) : null}
        </header>

        <main className="grid gap-6 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_minmax(0,1fr)]">
          <Card
            title="Tenant"
            description={tenantLoading ? "正在加载我的租户" : "我的租户、创建和编辑"}
            action={
              <Button size="sm" variant="ghost" onClick={() => void refreshTenants()}>
                刷新
              </Button>
            }
          >
            <div className="space-y-4">
              <div className="space-y-3 rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <div className="space-y-2">
                  <Label htmlFor="tenantLookupId">加载已有租户</Label>
                  <div className="flex gap-2">
                    <Input
                      id="tenantLookupId"
                      value={tenantLookupId}
                      onChange={(event) => setTenantLookupId(event.target.value)}
                      placeholder="tenant-123"
                    />
                    <Button variant="secondary" onClick={handleLookupTenant}>
                      加载
                    </Button>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="tenantName">创建租户</Label>
                  <Input
                    id="tenantName"
                    value={tenantName}
                    onChange={(event) => setTenantName(event.target.value)}
                    placeholder="Acme Games"
                  />
                </div>
                <div className="space-y-2">
                  <Input
                    value={tenantSlug}
                    onChange={(event) => setTenantSlug(event.target.value)}
                    placeholder="acme-games"
                  />
                  <p className="text-xs text-slate-500">留空时会自动生成 slug。</p>
                </div>
                <Button variant="primary" onClick={handleCreateTenant}>
                  创建租户
                </Button>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm text-slate-500">
                  <span>租户列表</span>
                  <span>{tenants.length}</span>
                </div>
                <div className="space-y-2">
                  {tenants.length === 0 ? (
                    <EmptyState
                      title="暂无租户"
                      description="先连接平台，或创建一个租户开始。"
                    />
                  ) : (
                    tenants.map((tenant) => (
                      <button
                        key={tenant.id}
                        type="button"
                        onClick={() => handleTenantSelect(tenant)}
                        className={
                          tenant.id === selectedTenantId
                            ? "w-full rounded-2xl border border-teal-200 bg-teal-50 p-4 text-left shadow-sm"
                            : "w-full rounded-2xl border border-slate-200 bg-white p-4 text-left transition hover:border-slate-300 hover:bg-slate-50"
                        }
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-1">
                            <div className="text-sm font-medium text-slate-950">{tenant.name}</div>
                            <div className="text-xs text-slate-500">{tenant.slug}</div>
                          </div>
                          <Badge tone={tenant.id === selectedTenantId ? "success" : "neutral"}>{tenant.id}</Badge>
                        </div>
                      </button>
                    ))
                  )}
                </div>
              </div>

              <Separator />

              <div className="space-y-3">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-sm font-semibold text-slate-950">当前租户</h3>
                    <p className="text-xs text-slate-500">{selectedTenant?.id ?? "未选择"}</p>
                  </div>
                  <Button variant="destructive" size="sm" onClick={handleDeleteTenant} disabled={!selectedTenantId}>
                    删除
                  </Button>
                </div>
                {selectedTenant ? (
                  <div className="space-y-3 rounded-2xl border border-slate-200 bg-white p-4">
                    <div className="space-y-2">
                      <Label htmlFor="tenantEditName">名称</Label>
                      <Input
                        id="tenantEditName"
                        value={tenantEditName}
                        onChange={(event) => setTenantEditName(event.target.value)}
                      />
                    </div>
                    <div className="space-y-2">
                      <Label htmlFor="tenantEditSlug">Slug</Label>
                      <Input
                        id="tenantEditSlug"
                        value={tenantEditSlug}
                        onChange={(event) => setTenantEditSlug(event.target.value)}
                      />
                    </div>
                    <Button variant="primary" onClick={handleUpdateTenant}>
                      更新租户
                    </Button>
                  </div>
                ) : (
                  <EmptyState title="请选择一个租户" description="选中后可编辑或删除。" />
                )}
              </div>
            </div>
          </Card>

          <Card
            title="Project"
            description={projectLoading ? "正在加载项目" : "按租户管理项目"}
            action={
              <Button
                size="sm"
                variant="ghost"
                onClick={() => void refreshProjects(projectTenantId || selectedTenantId)}
              >
                刷新
              </Button>
            }
          >
            <div className="space-y-4">
              <div className="space-y-3 rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <div className="space-y-2">
                  <Label htmlFor="projectTenantId">Tenant ID</Label>
                  <Input
                    id="projectTenantId"
                    value={projectTenantId}
                    onChange={(event) => setProjectTenantId(event.target.value)}
                    placeholder="tenant-123"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="projectName">项目名称</Label>
                  <Input
                    id="projectName"
                    value={projectName}
                    onChange={(event) => setProjectName(event.target.value)}
                    placeholder="Platform Core"
                  />
                </div>
                <div className="space-y-2">
                  <Input
                    value={projectKey}
                    onChange={(event) => setProjectKey(event.target.value)}
                    placeholder="platform-core"
                  />
                  <p className="text-xs text-slate-500">留空时会自动生成 key。</p>
                </div>
                <Button variant="primary" onClick={handleCreateProject}>
                  创建项目
                </Button>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm text-slate-500">
                  <span>项目列表</span>
                  <span>{selectedProjects.length}</span>
                </div>
                <div className="space-y-2">
                  {!selectedTenantId ? (
                    <EmptyState title="先选择租户" description="项目列表依赖 tenantId。" />
                  ) : selectedProjects.length === 0 ? (
                    <EmptyState title="暂无项目" description="在上方创建第一个项目。" />
                  ) : (
                    selectedProjects.map((project) => (
                      <button
                        key={project.id}
                        type="button"
                        onClick={() => handleProjectSelect(project)}
                        className={
                          project.id === selectedProjectId
                            ? "w-full rounded-2xl border border-teal-200 bg-teal-50 p-4 text-left shadow-sm"
                            : "w-full rounded-2xl border border-slate-200 bg-white p-4 text-left transition hover:border-slate-300 hover:bg-slate-50"
                        }
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-1">
                            <div className="text-sm font-medium text-slate-950">{project.name}</div>
                            <div className="text-xs text-slate-500">{project.key}</div>
                          </div>
                          <Badge tone={project.id === selectedProjectId ? "success" : "neutral"}>
                            {project.id}
                          </Badge>
                        </div>
                      </button>
                    ))
                  )}
                </div>
              </div>
            </div>
          </Card>

          <Card
            title="Environment"
            description={environmentLoading ? "正在加载环境" : "按项目管理环境"}
            action={
              <Button
                size="sm"
                variant="ghost"
                onClick={() => void refreshEnvironments(environmentProjectId || selectedProjectId)}
              >
                刷新
              </Button>
            }
          >
            <div className="space-y-4">
              <div className="space-y-3 rounded-2xl border border-slate-200 bg-slate-50/80 p-4">
                <div className="space-y-2">
                  <Label htmlFor="environmentProjectId">Project ID</Label>
                  <Input
                    id="environmentProjectId"
                    value={environmentProjectId}
                    onChange={(event) => setEnvironmentProjectId(event.target.value)}
                    placeholder="project-123"
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="environmentName">环境名称</Label>
                  <Input
                    id="environmentName"
                    value={environmentName}
                    onChange={(event) => setEnvironmentName(event.target.value)}
                    placeholder="dev"
                  />
                </div>
                <div className="space-y-2">
                  <Input
                    value={environmentDisplayName}
                    onChange={(event) => setEnvironmentDisplayName(event.target.value)}
                    placeholder="Development"
                  />
                  <p className="text-xs text-slate-500">displayName 可为空，默认使用环境名称。</p>
                </div>
                <Button variant="primary" onClick={handleCreateEnvironment}>
                  创建环境
                </Button>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between text-sm text-slate-500">
                  <span>环境列表</span>
                  <span>{selectedEnvironments.length}</span>
                </div>
                <div className="space-y-2">
                  {!selectedProjectId ? (
                    <EmptyState title="先选择项目" description="环境列表依赖 projectId。" />
                  ) : selectedEnvironments.length === 0 ? (
                    <EmptyState title="暂无环境" description="在上方创建第一个环境。" />
                  ) : (
                    selectedEnvironments.map((environment) => (
                      <div
                        key={environment.id}
                        className="rounded-2xl border border-slate-200 bg-white p-4"
                      >
                        <div className="flex items-start justify-between gap-3">
                          <div className="space-y-1">
                            <div className="text-sm font-medium text-slate-950">{environment.displayName}</div>
                            <div className="text-xs text-slate-500">
                              {environment.name} · {environment.projectId}
                            </div>
                          </div>
                          <Badge tone="neutral">{environment.id}</Badge>
                        </div>
                      </div>
                    ))
                  )}
                </div>
              </div>
            </div>
          </Card>
        </main>
      </div>
    </div>
  );
}
