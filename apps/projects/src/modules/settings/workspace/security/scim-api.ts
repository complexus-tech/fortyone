"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { z } from "zod";
import { useUserRole } from "@/hooks/role";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { get, post, remove } from "@/lib/http";

const credentialSchema = z.object({
  id: z.string().uuid(),
  name: z.string(),
  prefix: z.string(),
  createdAt: z.string().datetime(),
  expiresAt: z.string().datetime(),
  revokedAt: z.string().datetime().nullable(),
});
const statusSchema = z.object({
  credentials: z.array(credentialSchema),
  managedUsers: z.number().int().nonnegative(),
  pendingSeatSync: z.boolean(),
  seatSyncError: z.string(),
});
const mintedSchema = z.object({
  credential: credentialSchema,
  token: z.string(),
});
const decode =
  <T>(schema: z.ZodType<T>) =>
  (response: unknown): T => {
    if (typeof response !== "object" || !response || !("data" in response))
      throw new Error("Missing provisioning data");
    return schema.parse(response.data);
  };
const useContext = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const { userRole } = useUserRole();
  return {
    ctx: { session, workspaceSlug },
    enabled: Boolean(session && workspaceSlug && userRole === "admin"),
    key: ["workspace-security", workspaceSlug, "scim"] as const,
  };
};
export const useSCIMStatus = () => {
  const { ctx, enabled, key } = useContext();
  return useQuery({
    queryKey: key,
    queryFn: () => get("security/scim", ctx, undefined, decode(statusSchema)),
    enabled,
  });
};
export const useSCIMMutations = () => {
  const { ctx, key } = useContext();
  const client = useQueryClient();
  const invalidate = async () => {
    await client.invalidateQueries({ queryKey: key });
    await client.invalidateQueries({
      queryKey: ["workspace-security", ctx.workspaceSlug, "audit"],
    });
  };
  const mint = useMutation({
    mutationFn: (input: { name: string; lifetimeDays: number }) =>
      post(
        "security/scim/credentials",
        input,
        ctx,
        { retry: 0 },
        decode(mintedSchema),
      ),
    onSuccess: invalidate,
  });
  const revoke = useMutation({
    mutationFn: (id: string) =>
      remove(`security/scim/credentials/${id}`, ctx, { retry: 0 }),
    onSuccess: invalidate,
  });
  const retry = useMutation({
    mutationFn: () => post("security/scim/retry-seats", {}, ctx, { retry: 0 }),
    onSuccess: invalidate,
  });
  return { mint, revoke, retry };
};
