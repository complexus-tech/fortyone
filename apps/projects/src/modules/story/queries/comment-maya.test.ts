/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { get, post } from "@/lib/http";
import { getStoryMayaRuns, updateCommentMayaRun } from "./comment-maya";

jest.mock("@/lib/http", () => ({ get: jest.fn(), post: jest.fn() }));
const ctx = { workspaceSlug: "acme" };
beforeEach(() => jest.clearAllMocks());

describe("story Maya transport", () => {
  it("reads capability and durable runs from the story-scoped route", async () => {
    const state = { enabled: true, runs: [] };
    jest.mocked(get).mockResolvedValue({ data: state });
    await expect(getStoryMayaRuns("story", ctx)).resolves.toEqual(state);
    expect(get).toHaveBeenCalledWith("stories/story/maya-runs", ctx);
  });
  it.each(["cancel", "retry", "approve"] as const)(
    "sends the observed version for %s",
    async (action) => {
      jest.mocked(post).mockResolvedValue({ data: {} });
      await updateCommentMayaRun("story", "run", action, 4, ctx);
      expect(post).toHaveBeenCalledWith(
        `stories/story/maya-runs/run/${action}`,
        { version: 4 },
        ctx,
      );
    },
  );
  it("surfaces a stale approval or committed cancellation conflict", async () => {
    jest.mocked(post).mockResolvedValue({
      error: { message: "This request has changed. Refresh and try again." },
    });
    await expect(
      updateCommentMayaRun("story", "run", "approve", 1, ctx),
    ).rejects.toThrow("This request has changed");
  });
});
