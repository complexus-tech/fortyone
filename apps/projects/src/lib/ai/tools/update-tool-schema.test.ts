/** @jest-environment node */ // eslint-disable-line tsdoc/syntax -- Jest requires this exact file-level pragma.
/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { asSchema } from "ai";
import type { ToolSet } from "ai";
import { updateKeyResultTool } from "./key-results/update-key-result";
import { updateObjectiveTool } from "./objectives/update-objective";
import { updateSprintSettings } from "./sprints/update-sprint-settings";
import { updateStory } from "./stories/update-story";
import { updateTeam } from "./teams/update-team";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/queries/workspaces/get-workspace", () => ({
  getWorkspace: jest.fn(),
}));
jest.mock("@/modules/story/actions/update-story", () => ({
  updateStoryAction: jest.fn(),
}));
jest.mock("@/modules/teams/actions/update-team", () => ({
  updateTeamAction: jest.fn(),
}));
jest.mock("@/modules/objectives/actions/update-objective", () => ({
  updateObjective: jest.fn(),
}));
jest.mock("@/modules/objectives/queries/get-objective", () => ({
  getObjective: jest.fn(),
}));
jest.mock("@/modules/objectives/actions/update-key-result", () => ({
  updateKeyResult: jest.fn(),
}));
jest.mock("@/modules/teams/actions/update-sprint-settings", () => ({
  updateSprintSettingsAction: jest.fn(),
}));

const validate = async (
  registeredTool: ToolSet[string],
  input: Record<string, unknown>,
) => asSchema(registeredTool.inputSchema).validate?.(input);

describe("Maya update tool schemas", () => {
  it.each([
    ["story", updateStory, { storyId: "story-1" }],
    ["team", updateTeam, { teamId: "team-1" }],
    ["objective", updateObjectiveTool, { objectiveId: "objective-1" }],
    ["key result", updateKeyResultTool, { keyResultId: "key-result-1" }],
    ["sprint settings", updateSprintSettings, { teamId: "team-1" }],
  ])("rejects an ID-only %s update", async (_name, registeredTool, input) => {
    const result = await validate(registeredTool, input);

    expect(result?.success).toBe(false);
  });

  it("accepts false as an explicit boolean update", async () => {
    const [teamResult, sprintResult] = await Promise.all([
      validate(updateTeam, { isPrivate: false, teamId: "team-1" }),
      validate(updateSprintSettings, {
        autoCreateSprints: false,
        teamId: "team-1",
      }),
    ]);

    expect(teamResult?.success).toBe(true);
    expect(sprintResult?.success).toBe(true);
  });

  it("accepts null as an intentional clear for a nullable story field", async () => {
    const result = await validate(updateStory, {
      estimateValue: null,
      storyId: "story-1",
    });

    expect(result?.success).toBe(true);
  });
});
