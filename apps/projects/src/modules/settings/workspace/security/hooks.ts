"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import type { AuditFilters, PolicyUpdate } from "./types";
import {
  downloadAudit,
  getPolicy,
  listAudit,
  listSessions,
  revokeMemberSessions,
  revokeSession,
  savePolicy,
} from "./api";

const useSecurityContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { userRole } = useUserRole();
  return {
    ctx: { session, workspaceSlug },
    enabled: Boolean(session && workspaceSlug && userRole === "admin"),
    key: ["workspace-security", workspaceSlug] as const,
  };
};
export const useSecurityPolicy = () => {
  const { ctx, enabled, key } = useSecurityContext();
  return useQuery({
    queryKey: [...key, "policy"],
    queryFn: () => getPolicy(ctx),
    enabled,
    refetchOnWindowFocus: false,
  });
};
export const useSaveSecurityPolicy = () => {
  const { ctx, key } = useSecurityContext();
  const client = useQueryClient();
  return useMutation({
    mutationFn: (input: PolicyUpdate) => savePolicy(ctx, input),
    onSuccess: async (policy) => {
      client.setQueryData([...key, "policy"], policy);
      await client.invalidateQueries({ queryKey: [...key, "audit"] });
    },
  });
};
export const useSecuritySessions = (
  userId: string,
  includeRevoked: boolean,
) => {
  const { ctx, enabled, key } = useSecurityContext();
  return useQuery({
    queryKey: [...key, "sessions", userId, includeRevoked],
    queryFn: () => listSessions(ctx, userId, includeRevoked),
    enabled,
    refetchInterval: 60_000,
  });
};
export const useRevokeSessions = () => {
  const { ctx, key } = useSecurityContext();
  const client = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      reason,
      member,
    }: {
      id: string;
      reason: string;
      member: boolean;
    }) =>
      member
        ? revokeMemberSessions(ctx, id, reason)
        : revokeSession(ctx, id, reason),
    onSuccess: () => client.invalidateQueries({ queryKey: key }),
  });
};
export const useSecurityAudit = (filters: AuditFilters, cursor = "") => {
  const { ctx, enabled, key } = useSecurityContext();
  return useQuery({
    queryKey: [...key, "audit", filters, cursor],
    queryFn: () => listAudit(ctx, filters, cursor),
    enabled,
  });
};
export const useDownloadAudit = () => {
  const { ctx } = useSecurityContext();
  return useMutation({
    mutationFn: (filters: AuditFilters) => downloadAudit(ctx, filters),
  });
};
