import type { WorkspaceCtx } from "@/lib/http";
import { getSprint } from "@/modules/sprints/queries/get-sprint-details";
import { normalizeOptionalString } from "../normalize-input";

export const createSprintEndDateResolver = (ctx: WorkspaceCtx) => {
  const endDatesBySprint = new Map<string, Promise<string>>();

  return async (
    sprintId: string | undefined,
    endDate: string | null | undefined,
  ) => {
    if (normalizeOptionalString(endDate) || !sprintId) {
      return endDate;
    }

    let endDatePromise = endDatesBySprint.get(sprintId);
    if (!endDatePromise) {
      endDatePromise = getSprint(sprintId, ctx).then((sprint) => {
        if (!sprint) {
          throw new Error(
            "Could not resolve the sprint before creating a story.",
          );
        }
        return sprint.endDate;
      });
      endDatesBySprint.set(sprintId, endDatePromise);
    }

    return endDatePromise;
  };
};
