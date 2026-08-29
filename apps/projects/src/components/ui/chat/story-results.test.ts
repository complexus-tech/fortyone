/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { ToolMessagePart } from "./tool-output-policy";
import {
  getStoryResults,
  getVisibleStoryResults,
  isStoryResultsOutput,
  STORY_RESULTS_PREVIEW_LIMIT,
} from "./story-results-data";
import {
  isRenderableToolPart,
  isSupportingToolType,
} from "./tool-output-policy";

const asToolPart = (part: { output: unknown; state: string; type: string }) =>
  part as ToolMessagePart;

describe("story results generative UI", () => {
  it("normalizes grouped story results without exposing group helper data", () => {
    const output = {
      success: true,
      stories: [
        {
          key: "started",
          totalCount: 2,
          stories: [
            {
              id: "story-1",
              priority: "High",
              statusId: "status-started",
              title: "Add generative story results",
            },
            {
              id: "story-2",
              priority: "Low",
              statusId: "status-started",
              title: "Document the result renderer",
            },
          ],
        },
      ],
    };

    expect(getStoryResults(output)).toEqual([
      {
        id: "story-1",
        priority: "High",
        statusId: "status-started",
        title: "Add generative story results",
      },
      {
        id: "story-2",
        priority: "Low",
        statusId: "status-started",
        title: "Document the result renderer",
      },
    ]);
  });

  it("normalizes flat search results and tolerates invalid saved entries", () => {
    const output = {
      success: true,
      stories: [
        {
          id: "story-1",
          priority: "Unexpected",
          statusId: "status-backlog",
          title: "Search result",
        },
        {
          id: "story-1",
          priority: "High",
          statusId: "status-backlog",
          title: "Duplicate result",
        },
        { id: "incomplete" },
      ],
    };

    expect(getStoryResults(output)).toEqual([
      {
        id: "story-1",
        priority: "No Priority",
        statusId: "status-backlog",
        title: "Search result",
      },
    ]);
  });

  it("recognizes an empty successful result as renderable", () => {
    const output = { success: true, stories: [] };
    const part = asToolPart({
      output,
      state: "output-available",
      type: "tool-listTeamStories",
    });

    expect(isStoryResultsOutput(output)).toBe(true);
    expect(isRenderableToolPart(part)).toBe(true);
  });

  it("keeps status lookups supporting-only when a prompt uses multiple tools", () => {
    const statusesPart = asToolPart({
      output: {
        success: true,
        data: [{ id: "status-started", name: "Started" }],
      },
      state: "output-available",
      type: "tool-statuses",
    });
    const storiesPart = asToolPart({
      output: {
        success: true,
        stories: [
          {
            id: "story-1",
            priority: "High",
            statusId: "status-started",
            title: "Visible story",
          },
        ],
      },
      state: "output-available",
      type: "tool-listTeamStories",
    });

    expect(isSupportingToolType(statusesPart.type)).toBe(true);
    expect(isRenderableToolPart(statusesPart)).toBe(false);
    expect(isRenderableToolPart(storiesPart)).toBe(true);
  });

  it("shows five stories initially and can reveal the complete result", () => {
    const stories = Array.from(
      { length: STORY_RESULTS_PREVIEW_LIMIT + 3 },
      (_, index) => ({
        id: `story-${index}`,
        priority: "Medium" as const,
        statusId: "status-started",
        title: `Story ${index}`,
      }),
    );

    expect(getVisibleStoryResults(stories, false)).toHaveLength(5);
    expect(getVisibleStoryResults(stories, true)).toHaveLength(8);
  });
});
