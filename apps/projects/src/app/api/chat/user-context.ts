import { auth } from "@/auth";
import type { Team } from "@/modules/teams/types";
import type { Workspace } from "@/types";
import { getMemories } from "@/modules/ai-chats/queries/get-memory";

export async function getUserContext({
  currentPath,
  currentTheme,
  resolvedTheme,
  subscription,
  teams,
  username,
  terminology,
  workspace,
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
  username: string;
  terminology: {
    stories: string;
    sprints: string;
    objectives: string;
    keyResults: string;
  };
  workspace: Workspace;
}): Promise<string> {
  const session = await auth();
  if (!session?.user) {
    return "";
  }

  const now = new Date();
  const currentDate = now.toISOString().split("T")[0]; // YYYY-MM-DD
  const currentTime = now.toLocaleTimeString("en-US", {
    hour12: false,
  });
  const memories = await getMemories(session);

  const teamsList = teams
    .map((t) => `name: ${t.name} - id: ${t.id} - code: ${t.code}`)
    .join(", ");

  return `
    **Current User Context:**
    - User ID: ${session.user.id}
    - Name: ${session.user.name}
    - Username: ${username}
    - Role: ${workspace.userRole}
    - Workspace: ${workspace.name}
    - Current Date: ${currentDate}
    - Current Time: ${currentTime}
    - Teams: ${teamsList}
    - Current Path: ${currentPath} the current page the user is on
    - Current Theme: ${currentTheme} the current theme the user is using(light, dark, system)
    - Resolved Theme: ${resolvedTheme} the resolved theme the user is using(light, dark)
    - Subscription Tier: ${subscription?.tier}
    - Subscription Billing Interval: ${subscription?.billingInterval}
    - Subscription Billing Ends At: ${subscription?.billingEndsAt}
    - Subscription Status: ${subscription?.status}
    ${
      memories.length > 0
        ? `
    **Long-term User Memories:**
    ${memories.map((m) => `
      - id: ${m.id}
      - content: ${m.content}
      - created at: ${m.createdAt}
      - updated at: ${m.updatedAt}
      `).join("\n")}

    `
        : ""
    }


    **Current Terminology:**
    - Stories: the user's selected terminology for stories is ${terminology.stories}
    - Sprints: the user's selected terminology for sprints is ${terminology.sprints}
    - Objectives: the user's selected terminology for objectives is ${terminology.objectives}
    - Key Results: the user's selected terminology for key results is ${terminology.keyResults}
    - Always use the user's selected terminology for stories, sprints, objectives, and key results.
    

    **Timezone:**
    - Timezone: ${Intl.DateTimeFormat().resolvedOptions().timeZone}
    - When displaying dates and times, use the user's timezone, the data returned from the server is in UTC

    **Time**
    - Do not display seconds in the time

    **"Me" Resolution:**
    When the user says "me", "my", "assign to me", "show my work", etc., use:
    - User ID: ${session.user.id}  
    - Name: ${session.user.name}

    **Personalized Responses:**
    Use the user's name (${session.user.name}) in responses where it makes sense to be more personal and engaging:
    - When greeting or acknowledging the user
    - When showing their personal work or assignments  
    - When confirming actions they requested
    - When providing status updates about their work

    **Smart Team Selection:**
    The user belongs to these teams: ${teamsList}
    ${teams.length === 1 ? "- Since user has only one team, auto-select it for story/objective creation" : '- When user says "my team" or doesn\'t specify team, ask which team or suggest based on context'}

    Examples:
    - "assign story to me" → "Assigned story to ${session.user.name}."
    - "show my stories" → "Here are your current stories, ${session.user.name}:" 
    - "create objective for me" → "Created objective with you as the lead, ${session.user.name}."
    - "what's my workload" → "${session.user.name}, you have 5 stories assigned to you."
    - "create story for my team" → ${teams.length === 1 ? `auto-select ${teams[0]?.name}` : "ask which team"}`;
}
