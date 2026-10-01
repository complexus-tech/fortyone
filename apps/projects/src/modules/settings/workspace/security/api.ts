import type { z } from "zod";
import { createApiClient } from "api-client";
import type { WorkspaceCtx } from "@/lib/http";
import { getApiUrl } from "@/lib/api-url";
import { get, post, put, remove } from "@/lib/http";
import type { AuditFilters, PolicyUpdate } from "./types";
import { auditPageSchema, policySchema, sessionListSchema } from "./types";

const decode =
  <T>(schema: z.ZodType<T>) =>
  (response: unknown): T => {
    if (
      typeof response !== "object" ||
      response === null ||
      !("data" in response)
    ) {
      throw new Error("Missing API data");
    }
    return schema.parse(response.data);
  };
export const auditSearch = (filters: AuditFilters) => {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(filters))
    if (value) params.set(key, value);
  return params;
};
export const getPolicy = (ctx: WorkspaceCtx) =>
  get("security/policy", ctx, undefined, decode(policySchema));
export const savePolicy = (ctx: WorkspaceCtx, input: PolicyUpdate) =>
  put("security/policy", input, ctx, { retry: 0 }, decode(policySchema));
export const listSessions = (
  ctx: WorkspaceCtx,
  userId: string,
  includeRevoked: boolean,
) => {
  const query = new URLSearchParams({ includeRevoked: String(includeRevoked) });
  if (userId) query.set("userId", userId);
  return get(
    `security/sessions?${query.toString()}`,
    ctx,
    undefined,
    decode(sessionListSchema),
  );
};
export const revokeSession = (ctx: WorkspaceCtx, id: string, reason: string) =>
  remove(`security/sessions/${id}`, ctx, { json: { reason }, retry: 0 });
export const revokeMemberSessions = (
  ctx: WorkspaceCtx,
  id: string,
  reason: string,
) =>
  post(`security/members/${id}/revoke-sessions`, { reason }, ctx, { retry: 0 });
export const listAudit = (
  ctx: WorkspaceCtx,
  filters: AuditFilters,
  cursor: string,
) => {
  const query = auditSearch(filters);
  query.set("limit", "50");
  if (cursor) query.set("cursor", cursor);
  return get(
    `security/audit?${query.toString()}`,
    ctx,
    undefined,
    decode(auditPageSchema),
  );
};
export const downloadAudit = async (
  ctx: WorkspaceCtx,
  filters: AuditFilters,
) => {
  const client = createApiClient(
    `${getApiUrl()}/workspaces/${encodeURIComponent(ctx.workspaceSlug)}`,
  );
  const response = await client.get(
    `security/audit/export?${auditSearch(filters).toString()}`,
    { retry: 0 },
  );
  if (!response.headers.get("Content-Type")?.startsWith("text/csv")) {
    throw new Error("The audit export could not be downloaded.");
  }
  const blob = await response.blob();
  const url = URL.createObjectURL(blob);
  try {
    const link = document.createElement("a");
    link.href = url;
    link.download = "workspace-audit.csv";
    document.body.appendChild(link);
    link.click();
    link.remove();
  } finally {
    setTimeout(() => {
      URL.revokeObjectURL(url);
    }, 1_000);
  }
};
