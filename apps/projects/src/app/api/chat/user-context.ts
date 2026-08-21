import type { Memory } from "@/modules/ai-chats/types";
import type { Team } from "@/modules/teams/types";
import type { Workspace } from "@/types";

type UserContextIdentity = {
  id: string;
  name?: string | null;
};

export function getUserContext({
  user,
  currentPath,
  currentTheme,
  resolvedTheme,
  subscription,
  joinedTeams,
  username,
  terminology,
  workspace,
  memories,
  totalMessages,
}: {
  user?: UserContextIdentity;
  currentPath: string;
  currentTheme: string;
  resolvedTheme: string;
  subscription?: {
    tier: string;
    billingInterval: string;
    billingEndsAt: string;
    status: string;
  };
  joinedTeams: Team[] | null;
  memories: Memory[];
  username?: string;
  terminology: {
    stories: string;
    sprints: string;
    objectives: string;
    keyResults: string;
  };
  workspace: Workspace;
  totalMessages: {
    current: number;
    limit: number;
  };
}): string {
  if (!user) {
    return "";
  }

  const displayName = user.name ?? "User";
  const usernameLabel = username ? ` (@${username})` : "";
  const now = new Date();
  const currentDate = now.toISOString().split("T")[0];
  const currentTime = now.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  let joinedTeamsSummary = "Unavailable";
  if (joinedTeams) {
    joinedTeamsSummary =
      joinedTeams.length > 0
        ? joinedTeams
            .map((team) => `${team.name} (${team.code}) [${team.id}]`)
            .join(", ")
        : "None";
  }

  const memoriesSummary =
    memories.length > 0
      ? memories.map((memory) => `- ${memory.id}: ${memory.content}`).join("\n")
      : "- None";

  let teamSelectionGuidance =
    'Joined-team membership could not be loaded. Call listTeams before answering a request about "my team"; do not infer membership from accessible or public teams.';
  if (joinedTeams?.length === 0) {
    teamSelectionGuidance =
      'The user has not joined a team. Say that plainly when they ask about "my team"; do not offer public teams as substitutes.';
  } else if (joinedTeams?.length === 1) {
    teamSelectionGuidance = `If the user says "my team" without naming one, use ${joinedTeams[0]?.name} [${joinedTeams[0]?.id}].`;
  } else if (joinedTeams && joinedTeams.length > 1) {
    teamSelectionGuidance =
      'If the user says "my team" and belongs to multiple teams, infer from conversation or the current path only when that team appears in this joined list; otherwise ask which joined team they mean.';
  }

  return `
Runtime context:
- User: ${displayName}${usernameLabel} [${user.id}]
- Workspace: ${workspace.name} (${workspace.slug}) [${workspace.id}]
- Workspace role: ${workspace.userRole}
- Current path: ${currentPath}
- Today: ${currentDate}
- Current time: ${currentTime}
- Timezone: ${timezone}
- Theme preference: ${currentTheme}
- Resolved theme: ${resolvedTheme}

Terminology:
- Stories => ${terminology.stories}
- Sprints => ${terminology.sprints}
- Objectives => ${terminology.objectives}
- Key Results => ${terminology.keyResults}

Joined teams:
- ${joinedTeamsSummary}
- ${teamSelectionGuidance}
- Public teams the user has not joined are not included here. Never treat a discoverable public team as one of the user's teams.

Subscription:
- Tier: ${subscription?.tier ?? "unknown"}
- Status: ${subscription?.status ?? "unknown"}
- Billing interval: ${subscription?.billingInterval ?? "unknown"}
- Billing ends at: ${subscription?.billingEndsAt ?? "unknown"}

Message usage:
- Current: ${totalMessages.current}
- Limit: ${totalMessages.limit}

Memories:
${memoriesSummary}

"Me" resolution:
- When the user says "me", "my", or "assign to me", resolve to ${displayName} [${user.id}].

Date handling:
- Server dates are UTC.
- Present dates/times in ${timezone}.
- Do not show seconds.
`;
}
