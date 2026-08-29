import type { WorkspaceCtx } from "@/lib/http";
import { getTeamStatuses } from "@/lib/queries/states/get-team-states";
import type { StoryStatus } from "./select-story-status";
import { selectStoryStatusId } from "./select-story-status";

export const createStoryStatusResolver = (ctx: WorkspaceCtx) => {
  const statusesByTeam = new Map<string, Promise<StoryStatus[]>>();

  return async (teamId: string, requestedStatusId?: string | null) => {
    let statusesPromise = statusesByTeam.get(teamId);
    if (!statusesPromise) {
      statusesPromise = getTeamStatuses(teamId, ctx);
      statusesByTeam.set(teamId, statusesPromise);
    }

    const statuses = await statusesPromise;
    return selectStoryStatusId(statuses, requestedStatusId);
  };
};
