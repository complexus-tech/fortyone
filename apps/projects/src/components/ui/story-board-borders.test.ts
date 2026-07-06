/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("story board border sizing", () => {
  it("uses half-pixel borders for profile tabs and story filter toolbar seams", () => {
    const sources = [
      readSource("src/modules/profile/components/all-stories.tsx"),
      readSource("src/components/ui/stories-filter-bar.tsx"),
    ].join("\n");

    expect(sources).not.toContain(" border-b ");
    expect(sources).toContain("border-b-[0.5px]");
  });

  it("uses half-pixel borders in board loading skeleton seams", () => {
    const boardSkeleton = readSource("src/components/ui/board-skeleton.tsx");

    expect(boardSkeleton).not.toContain(" border-t ");
    expect(boardSkeleton).not.toContain(" border-b ");
    expect(boardSkeleton).toContain("border-y-[0.5px]");
  });
});
