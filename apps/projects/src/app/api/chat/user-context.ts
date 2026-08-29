import type { Memory } from "@/modules/ai-chats/types";
import type { Team } from "@/modules/teams/types";
import type { Workspace } from "@/types";
import type { StoryCreationDefaults } from "./story-creation-defaults";

type UserContextIdentity = {
  id: string;
  name?: string | null;
};

const MAX_CONTEXT_MEMORIES = 12;
const MAX_MEMORY_CHARACTERS = 500;
const MAX_CONTEXT_TEAMS = 20;

const truncateMemory = (content: string) =>
  content.length <= MAX_MEMORY_CHARACTERS
    ? content
    : `${content.slice(0, MAX_MEMORY_CHARACTERS).trimEnd()}…`;

export function getUserContext({
  user,
  currentPath,
  currentTheme,
  resolvedTheme,
  joinedTeams,
  username,
  terminology,
  workspace,
  memories,
  storyCreationDefaults,
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
  storyCreationDefaults?: StoryCreationDefaults | null;
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
            .slice(0, MAX_CONTEXT_TEAMS)
            .map((team) => `${team.name} (${team.code}) [${team.id}]`)
            .join(", ") +
          (joinedTeams.length > MAX_CONTEXT_TEAMS
            ? `, +${joinedTeams.length - MAX_CONTEXT_TEAMS} more`
            : "")
        : "None";
  }

  const memoriesSummary =
    memories.length > 0
      ? memories
          .slice(0, MAX_CONTEXT_MEMORIES)
          .map((memory) => `- ${memory.id}: ${truncateMemory(memory.content)}`)
          .join("\n") +
        (memories.length > MAX_CONTEXT_MEMORIES
          ? `\n- +${memories.length - MAX_CONTEXT_MEMORIES} more memories omitted`
          : "")
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

  const storyCreationDefaultsSummary = storyCreationDefaults
    ? `single-story suggestions: time needed=${storyCreationDefaults.singleStory.estimatedDurationMinutes} minutes, calendar scheduling=${storyCreationDefaults.singleStory.autoSchedulingEnabled ? "on" : "off"}; multiple-story default: no shared time estimate, calendar scheduling=off; calendar scheduling availability=${storyCreationDefaults.autoSchedulingAvailable ? "available" : "not available on the current plan"}`
    : "single-story suggestions unavailable; multiple-story default: no shared time estimate, calendar scheduling=off";

  return `
Runtime context:
- User: ${displayName}${usernameLabel} [${user.id}]
- Workspace: ${workspace.name} (${workspace.slug}) [${workspace.id}]
- Role: ${workspace.userRole}; path: ${currentPath}
- Local time: ${currentDate} ${currentTime} (${timezone})
- Theme: ${currentTheme} (resolved ${resolvedTheme})
- Terms: stories=${terminology.stories}; sprints=${terminology.sprints}; objectives=${terminology.objectives}; key results=${terminology.keyResults}
- Account story-creation defaults: ${storyCreationDefaultsSummary}
- Single-story defaults are suggestions, not consent. Never apply or mention them as batch-wide defaults. A date or future-work phrase is not consent to reserve calendar time.

Joined teams:
- ${joinedTeamsSummary}
- ${teamSelectionGuidance}
- Never treat a discoverable public team as one of the user's teams.

Memories:
${memoriesSummary}

Resolution: "me"/"my"/"assign to me" resolve to ${displayName} [${user.id}]. Server dates are UTC; present them in ${timezone} without seconds.
`;
}
