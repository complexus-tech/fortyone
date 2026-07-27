/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { getVisibleToolPartIndexes } from "./chat-message-utils";
import { getEntityResultsModel } from "./entity-results-data";
import type { ToolMessagePart } from "./tool-output-policy";
import { isRenderableToolPart } from "./tool-output-policy";

const asToolPart = (part: { output: unknown; state: string; type: string }) =>
  part as ToolMessagePart;

describe("entity results generative UI", () => {
  it("maps objectives to their interactive objective pages", () => {
    const output = {
      success: true,
      objectives: [
        {
          health: "At Risk",
          id: "objective-1",
          name: "Improve activation",
          teamId: "team-1",
        },
      ],
    };

    expect(getEntityResultsModel("tool-listObjectivesTool", output)).toEqual({
      emptyMessage: "No objectives found.",
      items: [
        {
          href: "/teams/team-1/objectives/objective-1",
          icon: { kind: "icon", name: "objective" },
          id: "objective-1",
          title: "Improve activation",
          trailing: { kind: "status", label: "At Risk", tone: "warning" },
        },
      ],
      title: "Objectives",
    });
  });

  it("flattens multi-team integration requests and preserves priority", () => {
    const output = {
      success: true,
      teams: [
        {
          requests: [
            {
              id: "request-1",
              priority: "Urgent",
              status: "pending",
              title: "Handle failed deployment",
            },
          ],
          teamId: "team-1",
        },
      ],
    };

    expect(
      getEntityResultsModel("tool-listIntegrationRequestsTool", output),
    ).toMatchObject({
      items: [
        {
          href: "/teams/team-1/requests/request-1",
          icon: { kind: "priority", priority: "Urgent" },
          title: "Handle failed deployment",
          trailing: { kind: "status", label: "pending", tone: "info" },
        },
      ],
      title: "Integration requests",
    });
  });

  it("renders empty successful list results but ignores mutation outputs", () => {
    const listPart = asToolPart({
      output: { success: true, notifications: [] },
      state: "output-available",
      type: "tool-notifications",
    });
    const mutationPart = asToolPart({
      output: { message: "Notification marked as read", success: true },
      state: "output-available",
      type: "tool-notifications",
    });

    expect(isRenderableToolPart(listPart)).toBe(true);
    expect(isRenderableToolPart(mutationPart)).toBe(false);
  });

  it("keeps unsafe link protocols non-interactive", () => {
    const model = getEntityResultsModel("tool-links", {
      links: [
        {
          id: "link-1",
          title: "Unsafe link",
          url: "data:text/html,unsafe",
        },
      ],
      success: true,
    });

    expect(model?.items[0]).toMatchObject({
      href: undefined,
      title: "Unsafe link",
    });
  });

  it("maps all supported list families to the shared result model", () => {
    const cases: [string, Record<string, unknown>, string][] = [
      ["tool-listKeyResultsTool", { keyResults: [] }, "Key results"],
      ["tool-listSprints", { sprints: [] }, "Sprints"],
      ["tool-listTeams", { teams: [] }, "Teams"],
      ["tool-listTeamMembers", { members: [] }, "Members"],
      ["tool-listCustomerFeedbackTool", { teams: [] }, "Customer feedback"],
      ["tool-comments", { comments: [] }, "Comments"],
      ["tool-labels", { labels: [] }, "Labels"],
      ["tool-links", { links: [] }, "Links"],
    ];

    cases.forEach(([toolType, output, title]) => {
      expect(
        getEntityResultsModel(toolType, { ...output, success: true })?.title,
      ).toBe(title);
    });
  });

  it("hides team resolution output when it only feeds a requested list", () => {
    const message = {
      id: "message-1",
      parts: [
        {
          output: {
            success: true,
            teams: [{ id: "team-1", name: "Product" }],
          },
          state: "output-available",
          type: "tool-listTeams",
        },
        {
          output: {
            success: true,
            teams: [{ feedback: [], teamId: "team-1" }],
          },
          state: "output-available",
          type: "tool-listCustomerFeedbackTool",
        },
      ],
      role: "assistant",
    } as unknown as MayaUIMessage;

    expect(Array.from(getVisibleToolPartIndexes(message))).toEqual([1]);
  });

  it("keeps independent list results when neither feeds the other", () => {
    const message = {
      id: "message-2",
      parts: [
        {
          output: { labels: [], success: true },
          state: "output-available",
          type: "tool-labels",
        },
        {
          output: { notifications: [], success: true },
          state: "output-available",
          type: "tool-notifications",
        },
      ],
      role: "assistant",
    } as unknown as MayaUIMessage;

    expect(Array.from(getVisibleToolPartIndexes(message))).toEqual([0, 1]);
  });

  it("renders only the final result when a visible list tool runs repeatedly", () => {
    const message = {
      id: "message-3",
      parts: [
        {
          output: { kind: "story-list", stories: [], success: true },
          state: "output-available",
          type: "tool-searchStories",
        },
        {
          output: { kind: "story-list", stories: [], success: true },
          state: "output-available",
          type: "tool-searchStories",
        },
        {
          output: {
            kind: "story-list",
            stories: [{ id: "story-1", title: "Review community" }],
            success: true,
          },
          state: "output-available",
          type: "tool-searchStories",
        },
      ],
      role: "assistant",
    } as unknown as MayaUIMessage;

    expect(Array.from(getVisibleToolPartIndexes(message))).toEqual([2]);
  });
});
