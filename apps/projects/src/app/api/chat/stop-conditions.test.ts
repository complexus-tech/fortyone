/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { StepResult } from "ai";
import type { tools } from "@/lib/ai/tools";
import { hasTerminalStoryCreationResult } from "./stop-conditions";

const stepWithToolResult = (toolName: string, output: unknown) =>
  ({
    toolResults: [{ output, toolName }],
  }) as unknown as StepResult<typeof tools>;

describe("hasTerminalStoryCreationResult", () => {
  it.each(["createStory", "bulkCreateStories"])(
    "stops after a successful %s result",
    (toolName) => {
      expect(
        hasTerminalStoryCreationResult({
          steps: [stepWithToolResult(toolName, { success: true })],
        }),
      ).toBe(true);
    },
  );

  it("stops when story creation needs confirmation", () => {
    expect(
      hasTerminalStoryCreationResult({
        steps: [
          stepWithToolResult("bulkCreateStories", {
            needsConfirmation: true,
            success: false,
          }),
        ],
      }),
    ).toBe(true);
  });

  it("stops after a failed story result but not a read-only result", () => {
    expect(
      hasTerminalStoryCreationResult({
        steps: [stepWithToolResult("listTeamStories", { success: true })],
      }),
    ).toBe(false);
    expect(
      hasTerminalStoryCreationResult({
        steps: [
          stepWithToolResult("bulkCreateStories", {
            error: "Story creation failed.",
            success: false,
          }),
        ],
      }),
    ).toBe(true);
  });
});
