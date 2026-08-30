import type { DetailedStory } from "../../types";

export type CollaboratorSummary = Pick<
  DetailedStory["collaborators"][number],
  "id" | "username" | "fullName" | "avatarUrl"
>;

export const selectCollaborators = (
  collaboratorIds: readonly string[] | null | undefined,
  collaborators: DetailedStory["collaborators"] | null | undefined,
  members: CollaboratorSummary[],
): CollaboratorSummary[] => {
  const collaboratorLookup = new Map<string, CollaboratorSummary>(
    members.map(({ id, username, fullName, avatarUrl }) => [
      id,
      { id, username, fullName, avatarUrl },
    ]),
  );

  for (const collaborator of collaborators ?? []) {
    collaboratorLookup.set(collaborator.id, collaborator);
  }

  return (collaboratorIds ?? []).flatMap((id) => {
    const collaborator = collaboratorLookup.get(id);
    return collaborator ? [collaborator] : [];
  });
};
