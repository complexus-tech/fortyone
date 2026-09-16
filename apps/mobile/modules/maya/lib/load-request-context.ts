import type { QueryClient } from "@tanstack/react-query";
import type { ApiResponse, User } from "@/types";
import type { Workspace, WorkspaceSettings } from "@/types/workspace";
import { userKeys, workspaceKeys } from "@/constants/keys";
import { get } from "@/lib/http";
import { getWorkspaceSettings } from "@/lib/queries/get-workspace-settings";
import { getSubscription } from "@/lib/queries/get-subscription";
import { getProfile } from "@/modules/users/queries/get-profile";
import { buildMayaLegacyContext, type MayaMemory } from "./request-context";
import type { MayaSessionScope } from "./session-scope";
import { assertMayaRequestNotAborted } from "./abort";

export const loadMayaRequestContext = async ({
  scope,
  queryClient,
  signal,
  assertCurrent,
}: {
  scope: MayaSessionScope;
  queryClient: QueryClient;
  signal: AbortSignal;
  assertCurrent: () => void;
}) => {
  assertCurrent();
  assertMayaRequestNotAborted(signal);
  const cachedWorkspaces = queryClient.getQueryData<Workspace[]>(
    workspaceKeys.lists(),
  );
  const cachedWorkspace = cachedWorkspaces?.find(
    (workspace) => workspace.slug === scope.workspace,
  );
  const cachedProfile = queryClient.getQueryData<User>(userKeys.profile());
  const cachedSettings = queryClient.getQueryData<WorkspaceSettings>(
    workspaceKeys.settings(),
  );
  const [workspace, settings, profile, memories, subscription] =
    await Promise.all([
      cachedWorkspace ??
        get<ApiResponse<Workspace>>("", { signal }).then(
          (response) => response.data!,
        ),
      cachedSettings ?? getWorkspaceSettings(signal),
      cachedProfile ?? getProfile(signal),
      get<ApiResponse<MayaMemory[]>>("users/memory", { signal }).then(
        (response) => response.data!,
      ),
      getSubscription(signal).catch((error: unknown) => {
        if (
          error &&
          typeof error === "object" &&
          "status" in error &&
          error.status === 404
        )
          return null;
        throw error;
      }),
    ]);
  assertCurrent();
  assertMayaRequestNotAborted(signal);
  return buildMayaLegacyContext({
    scope,
    workspace,
    settings,
    profile,
    memories,
    subscription,
  });
};
