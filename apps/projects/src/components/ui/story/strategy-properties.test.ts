/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { readFileSync } from "node:fs";
import { join } from "node:path";

const source = readFileSync(
  join(process.cwd(), "src/components/ui/story/strategy-properties.tsx"),
  "utf8",
);

describe("story strategy properties", () => {
  it("uses the objective short summary in the tooltip", () => {
    expect(source).toContain("objectiveDetails?.shortSummary");
    expect(source).not.toContain("resolvedObjective.description");
  });

  it("shows key-result progress in the tooltip", () => {
    expect(source).toContain("getKeyResultProgress(selectedKeyResult)");
    expect(source).toContain("Progress");
    expect(source).toContain("{keyResultProgress}%");
  });

  it("keeps tooltip wrappers outside interactive menu triggers", () => {
    const objectiveSection = source.slice(source.indexOf("{showObjective"));
    const keyResultSection = source.slice(source.indexOf("{showKeyResult"));

    expect(objectiveSection.indexOf("<Tooltip")).toBeLessThan(
      objectiveSection.indexOf("<ObjectiveKeyResultMenu"),
    );
    expect(keyResultSection.indexOf("<Tooltip")).toBeLessThan(
      keyResultSection.indexOf("<KeyResultMenu"),
    );
  });
});
