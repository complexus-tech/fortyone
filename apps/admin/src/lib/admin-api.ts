import { createApiClient } from "api-client";
import { getApiUrl } from "@/lib/env";
import type {
  AdminListParams,
  AuditLog,
  DashboardSummary,
  ListResult,
  UserOverview,
  UserSummary,
  WorkspaceOverview,
  WorkspaceSummary,
} from "@/lib/types";

type ApiResponse<T> = {
  data: T;
};

const adminClient = () => createApiClient(getApiUrl());

const unwrap = async <T>(response: Promise<Response>) => {
  const payload = (await response.then((res) => res.json())) as ApiResponse<T>;
  return payload.data;
};

const buildQuery = (params: Record<string, string | number | undefined>) => {
  const searchParams = new URLSearchParams();

  Object.entries(params).forEach(([key, value]) => {
    if (value === undefined || value === "") {
      return;
    }
    searchParams.set(key, String(value));
  });

  const query = searchParams.toString();
  return query ? `?${query}` : "";
};

export const getDashboardSummary = () =>
  unwrap<DashboardSummary>(
    adminClient().get("admin/summary", { cache: "no-store" }),
  );

export const getWorkspaces = (params: AdminListParams = {}) =>
  unwrap<ListResult<WorkspaceSummary>>(
    adminClient().get(`admin/workspaces${buildQuery(params)}`, {
      cache: "no-store",
    }),
  );

export const getWorkspace = (workspaceId: string) =>
  unwrap<WorkspaceOverview>(
    adminClient().get(`admin/workspaces/${workspaceId}`, { cache: "no-store" }),
  );

export const updateWorkspaceTrial = async (
  workspaceId: string,
  input: { trialEndsOn: string; reason: string },
) => {
  return unwrap<WorkspaceOverview>(
    adminClient().patch(`admin/workspaces/${workspaceId}/trial`, {
      json: input,
    }),
  );
};

export const getUsers = (params: AdminListParams = {}) =>
  unwrap<ListResult<UserSummary>>(
    adminClient().get(`admin/users${buildQuery(params)}`, {
      cache: "no-store",
    }),
  );

export const getUser = (userId: string) =>
  unwrap<UserOverview>(
    adminClient().get(`admin/users/${userId}`, { cache: "no-store" }),
  );

export const getAuditLogs = (
  params: AdminListParams & { workspaceId?: string; targetType?: string } = {},
) =>
  unwrap<ListResult<AuditLog>>(
    adminClient().get(
      `admin/audit-logs${buildQuery({
        page: params.page,
        limit: params.limit,
        workspaceId: params.workspaceId,
        targetType: params.targetType,
      })}`,
      { cache: "no-store" },
    ),
  );
