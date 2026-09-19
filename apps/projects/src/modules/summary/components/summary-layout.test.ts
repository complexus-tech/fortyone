/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const readSource = (path: string) =>
  readFileSync(join(process.cwd(), path), "utf8");

describe("summary responsive layout", () => {
  it("allows both dashboard columns and their cards to shrink", () => {
    const summarySource = readSource("src/modules/summary/index.tsx");
    const storiesSource = readSource(
      "src/modules/summary/components/my-stories.tsx",
    );
    const activitiesSource = readSource(
      "src/modules/summary/components/activities.tsx",
    );

    expect(summarySource).toContain(
      "grid min-w-0 grid-cols-1 gap-4 @5xl:grid-cols-2",
    );
    expect(summarySource.match(/className="grid min-w-0"/g)).toHaveLength(2);
    expect(storiesSource).toContain("min-w-0 gap-4 overflow-hidden");
    expect(storiesSource).toContain("min-w-0 flex-1");
    expect(activitiesSource).toContain("min-w-0 overflow-hidden");
  });
});
