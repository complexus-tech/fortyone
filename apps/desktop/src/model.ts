import type { Operation, Story, WorkspaceData } from "./types";

export function findStories(
  data: WorkspaceData,
  userId: string,
  query: string,
  mine: boolean,
) {
  const search = query.trim().toLocaleLowerCase();
  return data.stories.filter((story) => {
    if (mine && story.assigneeId !== userId) return false;
    const team = data.teams.find((candidate) => candidate.id === story.teamId);
    return (
      !search ||
      `${story.title} ${story.description ?? ""} ${story.reference ?? `${team?.code ?? ""}-${story.sequenceId ?? ""}`}`
        .toLocaleLowerCase()
        .includes(search)
    );
  });
}

export function storyReference(story: Story, data: WorkspaceData) {
  const team = data.teams.find((candidate) => candidate.id === story.teamId);
  if (story.reference) return story.reference;
  return story.sequenceId
    ? `${team?.code ?? "STORY"}-${story.sequenceId}`
    : "New story";
}

export function formatTime(value: string | null) {
  if (!value) return "Not downloaded yet";
  const date = new Date(value);
  return Number.isNaN(date.getTime())
    ? "Unknown"
    : date.toLocaleString(undefined, {
        dateStyle: "medium",
        timeStyle: "short",
      });
}

export function resourceBaseVersion(
  kind: "story" | "document",
  resourceId: string,
  version: string | number | null | undefined,
  pending: Operation[],
): string | null {
  if (version !== undefined && version !== null && version !== "")
    return String(version);
  return pending.some(
    (operation) =>
      operation.kind === `${kind}.create` &&
      operation.resourceId === resourceId,
  )
    ? `pending-create:${resourceId}`
    : null;
}
