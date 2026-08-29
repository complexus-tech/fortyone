/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import { getStory, getStoryRef } from "./get-story";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
}));

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    data: unknown;
    status: number;

    constructor(message: string, status: number, data: unknown) {
      super(message);
      this.data = data;
      this.status = status;
    }
  },
}));

const mockedGet = jest.mocked(get);
const ctx = {} as WorkspaceCtx;

describe("story detail queries", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("uses the story reference endpoint for human-readable codes", async () => {
    const story = {
      collaboratorIds: ["collaborator-id"],
      id: "story-id",
      title: "Story",
    };
    mockedGet.mockResolvedValue({ data: story });

    await expect(getStoryRef("PRD-453", ctx)).resolves.toEqual(story);
    expect(mockedGet).toHaveBeenCalledWith("story-by-ref/PRD-453", ctx);
  });

  it.each([
    ["UUID lookup", () => getStory("story-id", ctx)],
    ["reference lookup", () => getStoryRef("PRD-453", ctx)],
  ])("normalizes null collaborator IDs for a %s", async (_, load) => {
    mockedGet.mockResolvedValue({
      data: {
        collaboratorIds: null,
        id: "story-id",
        title: "Story",
      },
    });

    await expect(load()).resolves.toMatchObject({ collaboratorIds: [] });
  });

  it.each([
    ["UUID lookup", () => getStory("story-id", ctx)],
    ["reference lookup", () => getStoryRef("PRD-453", ctx)],
  ])("returns null only when a %s is genuinely not found", async (_, load) => {
    mockedGet.mockRejectedValue(new ApiError("Not found", 404, null));

    await expect(load()).resolves.toBeNull();
  });

  it.each([
    ["UUID lookup", () => getStory("story-id", ctx)],
    ["reference lookup", () => getStoryRef("PRD-453", ctx)],
  ])("preserves a %s API failure", async (_, load) => {
    const error = new ApiError("Invalid request", 400, null);
    mockedGet.mockRejectedValue(error);

    await expect(load()).rejects.toBe(error);
  });
});
