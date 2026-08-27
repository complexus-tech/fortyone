/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { auth } from "@/auth";
import { remove } from "@/lib/http";
import { deleteStoryAction } from "./delete-story";
import { StoryDeletionOutcomeUncertainError } from "./story-deletion-error";

jest.mock("api-client", () => ({
  ApiError: class ApiError extends Error {
    constructor(
      message: string,
      public status: number,
      public data: unknown,
    ) {
      super(message);
    }
  },
}));

jest.mock("@/auth", () => ({
  auth: jest.fn(),
}));

jest.mock("@/lib/http", () => ({
  remove: jest.fn(),
}));

const authMock = jest.mocked(auth);
const removeMock = jest.mocked(remove);
const session = {
  user: {
    id: "user-1",
    name: "Joseph Mukorivo",
    email: "joseph@example.com",
    image: null,
    username: "joseph",
    fullName: "Joseph Mukorivo",
    isInternal: false,
    lastUsedWorkspaceId: "workspace-1",
  },
};

describe("deleteStoryAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("preserves a typed 4xx response as a definite deletion failure", async () => {
    const response = {
      data: null,
      error: { message: "You cannot delete this story." },
    };
    removeMock.mockRejectedValue(new ApiError("Forbidden", 403, response));

    await expect(deleteStoryAction("story-1", "complexus")).resolves.toEqual(
      response,
    );
    expect(removeMock).toHaveBeenCalledTimes(1);
  });

  it("recovers a lost response with one exact retry", async () => {
    const response = { data: null };
    removeMock
      .mockRejectedValueOnce(new TypeError("response lost after commit"))
      .mockResolvedValueOnce(response);

    await expect(deleteStoryAction("story-1", "complexus")).resolves.toEqual(
      response,
    );
    expect(removeMock).toHaveBeenCalledTimes(2);
    expect(removeMock).toHaveBeenNthCalledWith(
      1,
      "stories/story-1",
      expect.objectContaining({ workspaceSlug: "complexus" }),
    );
    expect(removeMock).toHaveBeenNthCalledWith(
      2,
      "stories/story-1",
      expect.objectContaining({ workspaceSlug: "complexus" }),
    );
  });

  it.each([
    [
      "a 5xx response that may follow a commit",
      new ApiError("Response serialization failed", 500, {
        error: { message: "Internal server error" },
      }),
    ],
    ["a lost response", new TypeError("fetch failed after request upload")],
    ["a response parse failure", new SyntaxError("Unexpected end of JSON")],
    ["a request timeout", new Error("Request timed out")],
  ])("propagates %s after one failed retry", async (_, cause) => {
    removeMock.mockRejectedValue(cause);

    const action = deleteStoryAction("story-1", "complexus");
    const uncertainty = await action.catch((error: unknown) => error);

    expect(uncertainty).toBeInstanceOf(StoryDeletionOutcomeUncertainError);
    expect((uncertainty as Error).cause).toBeInstanceOf(AggregateError);
    expect(((uncertainty as Error).cause as AggregateError).errors).toEqual([
      cause,
      cause,
    ]);
    expect(removeMock).toHaveBeenCalledTimes(2);
  });

  it("remains uncertain when the retry returns a typed 4xx", async () => {
    const firstError = new TypeError("response lost after commit");
    const retryError = new ApiError("Already deleted", 404, {
      error: { message: "Story not found" },
    });
    removeMock
      .mockRejectedValueOnce(firstError)
      .mockRejectedValueOnce(retryError);

    const uncertainty = await deleteStoryAction("story-1", "complexus").catch(
      (error: unknown) => error,
    );

    expect(uncertainty).toBeInstanceOf(StoryDeletionOutcomeUncertainError);
    expect(((uncertainty as Error).cause as AggregateError).errors).toEqual([
      firstError,
      retryError,
    ]);
    expect(removeMock).toHaveBeenCalledTimes(2);
  });

  it("does not classify authentication failure as a started mutation", async () => {
    const authError = new Error("Session lookup failed");
    authMock.mockRejectedValue(authError);

    await expect(deleteStoryAction("story-1", "complexus")).rejects.toBe(
      authError,
    );
    expect(removeMock).not.toHaveBeenCalled();
  });
});
