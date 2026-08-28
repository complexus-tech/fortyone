import type { FortyOneClient } from "./client.js";
import { apiErrorFromResponse } from "./errors.js";
import type { Story, StoryPage } from "./models.js";

export type { Story, StoryPage } from "./models.js";

export interface StoryPaginationOptions {
  teamId?: string;
  limit?: number;
  signal?: AbortSignal;
}

export async function* paginateStoryPages(
  client: FortyOneClient,
  workspaceId: string,
  options: StoryPaginationOptions = {},
): AsyncGenerator<StoryPage> {
  let cursor: string | undefined;
  const seenCursors = new Set<string>();
  for (;;) {
    const query = {
      ...(cursor ? { cursor } : {}),
      ...(options.limit === undefined ? {} : { limit: options.limit }),
      ...(options.teamId === undefined ? {} : { teamId: options.teamId }),
    };
    const result = await client.GET(
      "/api/v1/workspaces/{workspaceId}/stories",
      {
        params: { path: { workspaceId }, query },
        ...(options.signal ? { signal: options.signal } : {}),
      },
    );
    if (result.error !== undefined || result.data === undefined) {
      throw apiErrorFromResponse(result.response, result.error);
    }
    yield result.data;
    if (!result.data.meta.hasMore) {
      return;
    }
    const nextCursor = result.data.meta.nextCursor;
    if (!nextCursor || seenCursors.has(nextCursor)) {
      throw new Error(
        "FortyOne API returned an invalid or repeated pagination cursor",
      );
    }
    seenCursors.add(nextCursor);
    cursor = nextCursor;
  }
}

export async function* paginateStories(
  client: FortyOneClient,
  workspaceId: string,
  options: StoryPaginationOptions = {},
): AsyncGenerator<Story> {
  for await (const page of paginateStoryPages(client, workspaceId, options)) {
    yield* page.data;
  }
}
