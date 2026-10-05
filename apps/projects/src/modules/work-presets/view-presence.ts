import type { WorkspaceCtx } from "@/lib/http";
import { listPresets } from "./api";

const MAX_PRESENCE_PAGES = 100;

/** Existence checks use the same actor-authorized, active-only list as Views. */
export const hasSavedViews = async (
  teamId: string,
  ctx: WorkspaceCtx,
  signal: AbortSignal,
): Promise<boolean> => {
  const seen = new Set<string>();
  let cursor = "";
  for (let page = 0; page < MAX_PRESENCE_PAGES; page++) {
    if (signal.aborted)
      throw new DOMException("View lookup cancelled", "AbortError");
    if (seen.has(cursor))
      throw new Error("Views could not be checked. Please try again.");
    seen.add(cursor);
    // eslint-disable-next-line no-await-in-loop -- Each signed cursor comes from the preceding response.
    const result = await listPresets(teamId, "view", cursor, ctx, signal, 1);
    if (
      result.items.some(
        (item) => item.kind === "view" && item.teamId === teamId,
      )
    )
      return true;
    if (!result.nextCursor) return false;
    cursor = result.nextCursor;
  }
  throw new Error("Views could not be checked. Please try again.");
};
