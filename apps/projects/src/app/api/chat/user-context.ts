import { headers } from "next/headers";
import { auth } from "@/auth";

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

  return `
    **Current User Context:**
    - User ID: ${session.user.id}
    - Name: ${session.user.name}
    - Role: ${userRole}

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

    Examples:
    - "assign story to me" → "Assigned story to ${session.user.name}."
    - "show my stories" → "Here are your current stories, ${session.user.name}:" 
    - "create objective for me" → "Created objective with you as the lead, ${session.user.name}."
    - "what's my workload" → "${session.user.name}, you have 5 stories assigned to you."
    - General confirmations → "Done, ${session.user.name}!" or "Got it, ${session.user.name}!"`;
}
