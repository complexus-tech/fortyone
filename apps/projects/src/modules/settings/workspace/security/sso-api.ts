"use client";

import { z } from "zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { get, post, put, remove } from "@/lib/http";
import { getWorkspaceSSOStatus } from "@/modules/auth/public/sso-status";

const connectionSchema = z.object({
  id: z.string().uuid(),
  workspaceId: z.string().uuid(),
  issuer: z.string().url(),
  clientId: z.string(),
  enabled: z.boolean(),
  requireSSO: z.boolean(),
  generation: z.number().int().positive(),
  version: z.number().int().positive(),
  createdAt: z.string(),
  updatedAt: z.string(),
});
const settingsSchema = z.object({
  connection: connectionSchema.nullable(),
  verified: z.boolean(),
  callbackUrl: z.string().url(),
  signInUrl: z.string().url(),
});
export type SSOConnection = z.infer<typeof connectionSchema>;
export type SSOSettings = z.infer<typeof settingsSchema>;
export type SSOCreate = {
  issuer: string;
  clientId: string;
  clientSecret: string;
};
export type SSOUpdate = {
  enabled: boolean;
  requireSSO: boolean;
  expectedVersion: number;
  clientSecret?: string;
};
const decode = (response: unknown) => {
  if (
    typeof response !== "object" ||
    response === null ||
    !("data" in response)
  )
    throw new Error("Missing API data");
  return settingsSchema.parse(response.data);
};
const useContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { userRole } = useUserRole();
  return {
    ctx: { session, workspaceSlug },
    enabled: Boolean(session && workspaceSlug && userRole === "admin"),
    key: ["workspace-security", workspaceSlug, "sso"] as const,
  };
};
export const useSSOSettings = () => {
  const { ctx, enabled, key } = useContext();
  return useQuery({
    queryKey: key,
    queryFn: () => get("security/sso", ctx, undefined, decode),
    enabled,
    refetchOnWindowFocus: false,
  });
};

export const useWorkspaceSSOStatus = () => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["workspace-sso", workspaceSlug, "status"],
    enabled: Boolean(session && workspaceSlug),
    queryFn: () => getWorkspaceSSOStatus(workspaceSlug),
  });
};
export const useSSOMutations = () => {
  const { ctx, key } = useContext();
  const client = useQueryClient();
  const invalidate = () =>
    client.invalidateQueries({ queryKey: key.slice(0, 2) });
  return {
    create: useMutation({
      mutationFn: (input: SSOCreate) =>
        post("security/sso", input, ctx, { retry: 0 }),
      onSuccess: invalidate,
    }),
    update: useMutation({
      mutationFn: (input: SSOUpdate) =>
        put("security/sso", input, ctx, { retry: 0 }),
      onSuccess: invalidate,
    }),
    archive: useMutation({
      mutationFn: (reason: string) =>
        remove("security/sso", ctx, { json: { reason }, retry: 0 }),
      onSuccess: invalidate,
    }),
  };
};
