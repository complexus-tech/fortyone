import { inviteMembers } from "./invite";

const ONBOARDING_INVITATION_ROLE = "admin";

export const inviteOnboardingMembers = (
  emails: string[],
  teamIds: string[],
  workspaceSlug: string,
) => {
  return inviteMembers(
    emails.map((email) => ({
      email,
      role: ONBOARDING_INVITATION_ROLE,
      teamIds,
    })),
    workspaceSlug,
  );
};
