/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("TeamsMenu", () => {
  it("renders one compact placeholder set while initial teams load", () => {
    const source = readSource("src/components/ui/teams-menu.tsx");
    const initialLoaderUsages = source.match(
      /rows=\{INITIAL_TEAM_MENU_SKELETON_ROWS\}/g,
    );

    expect(source).not.toContain("<MenuLoadingSkeleton rows={4} />");
    expect(initialLoaderUsages).toHaveLength(1);
  });
});
