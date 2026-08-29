export const invitationKeys = {
  all: ["invitations"] as const,
  pending: (workspaceSlug: string) =>
    [...invitationKeys.all, "pending", workspaceSlug] as const,
  mine: ["my-invitations"] as const,
};
