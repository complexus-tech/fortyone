import { z } from "zod";
import type { WorkspaceCtx } from "@/lib/http";
import { listPresets } from "./api";
import type { Preset, SavedViewConfiguration } from "./types";

export type SavedView = Preset<SavedViewConfiguration> & { kind: "view" };

const MAX_RESOLUTION_PAGES = 100;

/** Resolve current definitions without retaining a configuration in browser preferences. */
export const resolveSavedViews = async (
  teamId: string,
  viewIds: readonly string[],
  ctx: WorkspaceCtx,
  signal?: AbortSignal,
): Promise<SavedView[]> => {
  const remaining = new Set(
    viewIds.filter((id) => z.uuid().safeParse(id).success),
  );
  if (!remaining.size) return [];
  const found: SavedView[] = [];
  const seen = new Set<string>();
  let cursor = "";
  for (let page = 0; page < MAX_RESOLUTION_PAGES; page++) {
    if (signal?.aborted)
      throw new DOMException("View lookup cancelled", "AbortError");
    if (seen.has(cursor))
      throw new Error("Views could not be loaded. Please try again.");
    seen.add(cursor);
    // Each cursor comes from the preceding response; pages cannot run concurrently.
    // eslint-disable-next-line no-await-in-loop -- Cursor pagination requires sequential requests.
    const result = await listPresets(teamId, "view", cursor, ctx, signal);
    for (const preset of result.items) {
      if (
        preset.kind === "view" &&
        preset.teamId === teamId &&
        remaining.delete(preset.id)
      ) {
        found.push(preset as SavedView);
      }
    }
    if (!remaining.size || !result.nextCursor) return found;
    cursor = result.nextCursor;
  }
  throw new Error(
    "This team has too many views to search automatically. Open Views to browse them.",
  );
};
