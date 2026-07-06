import { ApiError, createApiClient } from "api-client";
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

const writeAdminApiLog = (details: Record<string, unknown>) => {
  // eslint-disable-next-line no-console -- Surface protected Admin API failures in Vercel server logs.
  console.error("admin api request failed", details);
};

const logAdminApiError = (operation: string, error: unknown) => {
  if (error instanceof ApiError) {
    writeAdminApiLog({
      operation,
      status: error.status,
      message: error.message,
    });
    return;
  }

  if (error instanceof Error) {
    writeAdminApiLog({
      operation,
      name: error.name,
      message: error.message,
    });
    return;
  }

  writeAdminApiLog({ operation, error });
};

const unwrap = async <T>(response: Promise<Response>) => {
  const payload = (await response.then((res) => res.json())) as ApiResponse<T>;
  return payload.data;
};

const adminRequest = async <T>(
  operation: string,
  response: Promise<Response>,
) => {
  try {
    return await unwrap<T>(response);
  } catch (error) {
    logAdminApiError(operation, error);
    throw error;
  }
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
  adminRequest<DashboardSummary>(
    "get_dashboard_summary",
    adminClient().get("admin/summary", { cache: "no-store" }),
  );

export const getWorkspaces = (params: AdminListParams = {}) =>
  adminRequest<ListResult<WorkspaceSummary>>(
    "list_workspaces",
    adminClient().get(`admin/workspaces${buildQuery(params)}`, {
      cache: "no-store",
    }),
  );

export const getWorkspace = (workspaceId: string) =>
  adminRequest<WorkspaceOverview>(
    "get_workspace",
    adminClient().get(`admin/workspaces/${workspaceId}`, { cache: "no-store" }),
  );

export const updateWorkspaceTrial = async (
  workspaceId: string,
  input: { trialEndsOn: string; reason: string },
) => {
  return adminRequest<WorkspaceOverview>(
    "update_workspace_trial",
    adminClient().patch(`admin/workspaces/${workspaceId}/trial`, {
      json: input,
    }),
  );
};

export const getUsers = (params: AdminListParams = {}) =>
  adminRequest<ListResult<UserSummary>>(
    "list_users",
    adminClient().get(`admin/users${buildQuery(params)}`, {
      cache: "no-store",
    }),
  );

export const getUser = (userId: string) =>
  adminRequest<UserOverview>(
    "get_user",
    adminClient().get(`admin/users/${userId}`, { cache: "no-store" }),
  );

export const getAuditLogs = (
  params: AdminListParams & { workspaceId?: string; targetType?: string } = {},
) =>
  adminRequest<ListResult<AuditLog>>(
    "list_audit_logs",
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
