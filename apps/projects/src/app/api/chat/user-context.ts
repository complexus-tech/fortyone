import { auth } from "@/auth";
import type { Memory } from "@/modules/ai-chats/types";
import type { Team } from "@/modules/teams/types";
import type { Workspace } from "@/types";

export async function getUserContext({
  currentPath,
  currentTheme,
  resolvedTheme,
  subscription,
  teams,
  username,
  terminology,
  workspace,
  memories,
  totalMessages,
}: {
  currentPath: string;
  currentTheme: string;
  resolvedTheme: string;
  subscription?: {
    tier: string;
    billingInterval: string;
    billingEndsAt: string;
    status: string;
  };
  teams: Team[];
  memories: Memory[];
  username: string;
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
}): Promise<string> {
  const session = await auth();
  if (!session?.user) {
    return "";
  }

  const now = new Date();
  const currentDate = now.toISOString().split("T")[0];
  const currentTime = now.toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
  });
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;

  const teamsSummary =
    teams.length > 0
      ? teams.map((team) => `${team.name} (${team.code}) [${team.id}]`).join(", ")
      : "None";

  const memoriesSummary =
    memories.length > 0
      ? memories.map((memory) => `- ${memory.id}: ${memory.content}`).join("\n")
      : "- None";

  const teamSelectionGuidance =
    teams.length === 1
      ? `If team selection is needed and the user does not specify one, default to ${teams[0]?.name} [${teams[0]?.id}].`
      : "If team selection is needed and the user does not specify one, infer from context or ask a clarifying question.";

  return `
Runtime context:
- User: ${session.user.name} (@${username}) [${session.user.id}]
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

Teams:
- ${teamsSummary}
- ${teamSelectionGuidance}

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
- When the user says "me", "my", or "assign to me", resolve to ${session.user.name} [${session.user.id}].

Date handling:
- Server dates are UTC.
- Present dates/times in ${timezone}.
- Do not show seconds.
`;
}
