/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("NewStoryDialog team selection", () => {
  it("offers only teams the current user has joined", () => {
    const source = readSource("src/components/ui/new-story-dialog.tsx");

    expect(source).toContain("useJoinedTeams()");
    expect(source).not.toContain("useTeams()");
  });
});
