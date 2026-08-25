/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { StepResult } from "ai";
import type { tools } from "@/lib/ai/tools";
import { hasTerminalMutationResult } from "./stop-conditions";

const stepWithToolResult = (
  toolName: string,
  output: unknown,
  input: unknown = {},
) =>
  ({
    toolResults: [{ input, output, toolName }],
  }) as unknown as StepResult<typeof tools>;

describe("hasTerminalMutationResult", () => {
  it.each([
    "createStory",
    "bulkCreateStories",
    "updateTeam",
    "deleteObjectiveTool",
  ])("stops after a successful %s result", (toolName) => {
    expect(
      hasTerminalMutationResult({
        steps: [stepWithToolResult(toolName, { success: true })],
      }),
    ).toBe(true);
  });

  it("stops when story creation needs confirmation", () => {
    expect(
      hasTerminalMutationResult({
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
      hasTerminalMutationResult({
        steps: [stepWithToolResult("listTeamStories", { success: true })],
      }),
    ).toBe(false);
    expect(
      hasTerminalMutationResult({
        steps: [
          stepWithToolResult("bulkCreateStories", {
            error: "Story creation failed.",
            success: false,
          }),
        ],
      }),
    ).toBe(true);
  });

  it("stops only for mutating actions on multi-action tools", () => {
    expect(
      hasTerminalMutationResult({
        steps: [
          stepWithToolResult(
            "labels",
            { success: true },
            {
              action: "create-label",
            },
          ),
        ],
      }),
    ).toBe(true);
    expect(
      hasTerminalMutationResult({
        steps: [
          stepWithToolResult(
            "labels",
            { success: true },
            {
              action: "list-labels",
            },
          ),
        ],
      }),
    ).toBe(false);
  });

  it("allows a follow-up when a mutation returns an actionable install URL", () => {
    expect(
      hasTerminalMutationResult({
        steps: [
          stepWithToolResult("createGitHubInstallSessionTool", {
            installUrl: "https://github.com/apps/fortyone/installations/new",
            success: true,
          }),
        ],
      }),
    ).toBe(false);
  });
});
