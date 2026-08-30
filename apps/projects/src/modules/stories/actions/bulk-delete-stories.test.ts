/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { auth } from "@/auth";
import { remove } from "@/lib/http";
import { StoryDeletionOutcomeUncertainError } from "@/shared/story/deletion";
import { bulkDeleteAction } from "./bulk-delete-stories";

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
const storyIds = ["story-1", "story-2"];

describe("bulkDeleteAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("preserves a typed 4xx response without retrying", async () => {
    const response = {
      data: null,
      error: { message: "One or more stories cannot be deleted." },
    };
    removeMock.mockRejectedValue(new ApiError("Forbidden", 403, response));

    await expect(bulkDeleteAction({ storyIds }, "complexus")).resolves.toEqual(
      response,
    );
    expect(removeMock).toHaveBeenCalledTimes(1);
  });

  it("recovers a soft-deletion response loss with one exact retry", async () => {
    const response = {
      data: { deletedCount: storyIds.length, storyIds },
    };
    removeMock
      .mockRejectedValueOnce(new TypeError("response lost after commit"))
      .mockResolvedValueOnce(response);

    await expect(bulkDeleteAction({ storyIds }, "complexus")).resolves.toEqual(
      response,
    );
    expect(removeMock).toHaveBeenCalledTimes(2);
    expect(removeMock).toHaveBeenNthCalledWith(
      1,
      "stories",
      expect.objectContaining({ workspaceSlug: "complexus" }),
      { json: { storyIds, hardDelete: undefined } },
    );
    expect(removeMock).toHaveBeenNthCalledWith(
      2,
      "stories",
      expect.objectContaining({ workspaceSlug: "complexus" }),
      { json: { storyIds, hardDelete: undefined } },
    );
  });

  it("throws a deletion uncertainty marker after the soft retry also fails", async () => {
    const firstError = new TypeError("response lost after commit");
    const retryError = new ApiError("Server unavailable", 503, {
      error: { message: "Service unavailable" },
    });
    removeMock
      .mockRejectedValueOnce(firstError)
      .mockRejectedValueOnce(retryError);

    const uncertainty = await bulkDeleteAction({ storyIds }, "complexus").catch(
      (error: unknown) => error,
    );

    expect(uncertainty).toBeInstanceOf(StoryDeletionOutcomeUncertainError);
    expect(((uncertainty as Error).cause as AggregateError).errors).toEqual([
      firstError,
      retryError,
    ]);
    expect(removeMock).toHaveBeenCalledTimes(2);
  });

  it("does not retry an uncertain hard delete", async () => {
    const cause = new TypeError("hard-delete response lost after commit");
    removeMock.mockRejectedValue(cause);

    const uncertainty = await bulkDeleteAction(
      { storyIds, hardDelete: true },
      "complexus",
    ).catch((error: unknown) => error);

    expect(uncertainty).toBeInstanceOf(StoryDeletionOutcomeUncertainError);
    expect((uncertainty as Error).cause).toBe(cause);
    expect(removeMock).toHaveBeenCalledTimes(1);
  });
});
