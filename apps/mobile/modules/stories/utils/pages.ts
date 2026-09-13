import type { Story } from "../types";

// A refetch can move a story between pages. Keep its first position without
// rendering duplicate rows while the remaining pages catch up.
export const mergeStoryPages = (
  pages: readonly (readonly Story[])[],
): Story[] => {
  const seen = new Set<string>();
  return pages.flatMap((page) =>
    page.filter((story) => {
      if (seen.has(story.id)) return false;
      seen.add(story.id);
      return true;
    }),
  );
};
