import type { Subscription, User } from "@/types";
import type { Workspace, WorkspaceSettings } from "@/types/workspace";
import type { MayaSessionScope } from "./session-scope";

export type MayaMemory = {
  id: string;
  userId: string;
  workspaceId: string;
  content: string;
};

/** Preserve the deployed web contract until all clients use server hydration. */
export const buildMayaLegacyContext = ({
  scope,
  workspace,
  settings,
  profile,
  memories,
  subscription,
}: {
  scope: MayaSessionScope;
  workspace: Workspace;
  settings: WorkspaceSettings;
  profile: Pick<User, "id" | "username">;
  memories: MayaMemory[];
  subscription: Subscription | null;
}) => {
  if (
    !workspace?.id ||
    workspace.slug !== scope.workspace ||
    !workspace.isActive ||
    workspace.deletedAt ||
    profile?.id !== scope.userId ||
    (subscription && subscription.workspaceId !== workspace.id) ||
    !Array.isArray(memories) ||
    memories.some(
      (memory) =>
        memory.userId !== scope.userId ||
        memory.workspaceId !== workspace.id ||
        typeof memory.id !== "string" ||
        typeof memory.content !== "string",
    )
  ) {
    throw new Error("Your Maya workspace context changed. Open Maya again.");
  }
  const plural = (term: string) => {
    if (typeof term !== "string" || !term.trim())
      throw new Error("Your workspace terminology could not be loaded.");
    return term.endsWith("y") ? `${term.slice(0, -1)}ies` : `${term}s`;
  };
  return {
    workspace,
    username: profile.username,
    memories,
    terminology: {
      stories: plural(settings?.storyTerm),
      sprints: plural(settings?.sprintTerm),
      objectives: plural(settings?.objectiveTerm),
      keyResults: plural(settings?.keyResultTerm),
    },
    ...(subscription
      ? {
          subscription: {
            tier: subscription.tier,
            status: subscription.status,
            billingInterval: subscription.billingInterval,
            billingEndsAt: subscription.billingEndsAt,
          },
        }
      : {}),
  };
};
