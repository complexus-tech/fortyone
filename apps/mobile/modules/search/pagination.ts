import type { SearchResponse } from "./types";

export function getNextSearchPage(page: SearchResponse) {
  return page.page < page.totalPages &&
    (page.stories.length > 0 || page.objectives.length > 0)
    ? page.page + 1
    : undefined;
}

export function mergeSearchPages(
  pages: SearchResponse[] | undefined,
): SearchResponse | undefined {
  if (!pages?.length) return undefined;
  return {
    ...pages[0],
    stories: [
      ...new Map(
        pages.flatMap((page) => page.stories).map((story) => [story.id, story]),
      ).values(),
    ],
    objectives: [
      ...new Map(
        pages
          .flatMap((page) => page.objectives)
          .map((objective) => [objective.id, objective]),
      ).values(),
    ],
  };
}
