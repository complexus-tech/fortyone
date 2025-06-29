import { headers } from "next/headers";
import { auth } from "@/auth";
import { getTeams } from "@/modules/teams/queries/get-teams";

export async function getUserContext(): Promise<string> {
  const session = await auth();
  if (!session?.user) {
    return "";
  }
  const headersList = await headers();
  const subdomain = headersList.get("host")?.split(".")[0] || "";
  const workspace = session.workspaces.find(
    (w) => w.slug.toLowerCase() === subdomain.toLowerCase(),
  );
  const userRole = workspace?.userRole;

  const now = new Date();
  const currentDate = now.toISOString().split("T")[0]; // YYYY-MM-DD
  const currentTime = now.toLocaleTimeString("en-US", {
    hour12: false,
    timeZone: "UTC",
  });

  const teams = await getTeams(session);
  const teamsList = teams.map((t) => `${t.name} (${t.id})`).join(", ");

  return `
    **Current User Context:**
    - User ID: ${session.user.id}
    - Name: ${session.user.name}
    - Role: ${userRole}
    - Workspace: ${workspace?.name}
    - Current Date: ${currentDate}
    - Current Time: ${currentTime} UTC
    - Teams: ${teamsList}

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
