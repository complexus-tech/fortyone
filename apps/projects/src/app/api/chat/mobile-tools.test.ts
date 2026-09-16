/* global describe, expect, it, jest -- Jest globals. */
import { z } from "zod";
import type { ToolExecutionOptions } from "ai";
import { canUseMobileMayaTool, mobileMayaTools } from "./mobile-tools";

describe("mobile Maya capability boundary", () => {
  it("exposes the focus briefing tool used by daily-priority prompts", () => {
    const focusBrief = {
      inputSchema: z.object({}),
      execute: jest.fn(),
    };
    expect(canUseMobileMayaTool("focusBrief", {})).toBe(true);
    expect(mobileMayaTools({ focusBrief }).focusBrief).toBe(focusBrief);
  });
  it("keeps exact task operations and excludes unsupported management", () => {
    expect(canUseMobileMayaTool("deleteStory", {})).toBe(true);
    expect(canUseMobileMayaTool("navigation", {})).toBe(false);
    expect(canUseMobileMayaTool("deleteObjectiveTool", {})).toBe(false);
    expect(
      canUseMobileMayaTool("statuses", { action: "list-team-statuses" }),
    ).toBe(true);
    expect(canUseMobileMayaTool("statuses", { action: "delete-status" })).toBe(
      false,
    );
    expect(canUseMobileMayaTool("labels", { action: "delete-label" })).toBe(
      false,
    );
    expect(
      canUseMobileMayaTool("storyLabels", { action: "add-labels-to-story" }),
    ).toBe(true);
  });
  it("denies unsupported mixed-tool mutations before calling their executor", async () => {
    const execute = jest.fn<
      Promise<{ success: boolean }>,
      [unknown, ToolExecutionOptions]
    >(async () => ({ success: true }));
    const registered = {
      inputSchema: z.object({ action: z.string() }),
      execute,
    };
    const selected = mobileMayaTools({
      statuses: registered,
      navigation: registered,
    });
    expect(Object.keys(selected)).toEqual(["statuses"]);
    const options = { toolCallId: "call", messages: [] };
    expect(
      await selected.statuses.execute({ action: "delete-status" }, options),
    ).toMatchObject({ success: false });
    expect(execute).not.toHaveBeenCalled();
    expect(
      await selected.statuses.execute(
        { action: "list-team-statuses" },
        options,
      ),
    ).toMatchObject({ success: true });
    expect(execute).toHaveBeenCalledTimes(1);
  });
});
