/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { ApiError } from "api-client";
import { auth } from "@/auth";
import { post } from "@/lib/http";
import { createStoryAction } from "./create-story";
import { StoryCreationOutcomeUncertainError } from "./story-creation-error";

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
  post: jest.fn(),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);
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
const storyInput = {
  idempotencyKey: "maya:chat-1:tool-1",
  statusId: "00000000-0000-4000-8000-000000000002",
  teamId: "00000000-0000-4000-8000-000000000001",
  title: "Prepare launch",
};

describe("createStoryAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue(session);
  });

  it("preserves a typed 4xx response as a definite creation failure", async () => {
    const response = {
      data: null,
      error: { message: "The selected status does not belong to this team." },
    };
    postMock.mockRejectedValue(new ApiError("Invalid story", 422, response));

    await expect(createStoryAction(storyInput, "complexus")).resolves.toEqual(
      response,
    );
    expect(postMock).toHaveBeenCalledTimes(1);
  });

  it("recovers a lost response with one idempotent retry", async () => {
    const response = {
      data: { id: "story-1", title: storyInput.title },
    };
    postMock
      .mockRejectedValueOnce(new TypeError("response lost after commit"))
      .mockResolvedValueOnce(response);

    await expect(createStoryAction(storyInput, "complexus")).resolves.toEqual(
      response,
    );
    expect(postMock).toHaveBeenCalledTimes(2);
  });

  it.each([
    [
      "a 5xx response that may follow a commit",
      new ApiError("Response enrichment failed", 500, {
        error: { message: "Internal server error" },
      }),
    ],
    ["a lost response", new TypeError("fetch failed after request upload")],
    ["a response parse failure", new SyntaxError("Unexpected end of JSON")],
    ["a request timeout", new Error("Request timed out")],
  ])("propagates %s as an uncertain creation outcome", async (_, cause) => {
    postMock.mockRejectedValue(cause);

    const action = createStoryAction(storyInput, "complexus");
    await expect(action).rejects.toBeInstanceOf(
      StoryCreationOutcomeUncertainError,
    );
    const uncertainty = await action.catch((error: unknown) => error);
    expect(uncertainty).toEqual(
      expect.objectContaining({ name: "StoryCreationOutcomeUncertainError" }),
    );
    expect((uncertainty as Error).cause).toBeInstanceOf(AggregateError);
    expect(((uncertainty as Error).cause as AggregateError).errors).toEqual([
      cause,
      cause,
    ]);
    expect(postMock).toHaveBeenCalledTimes(2);
  });

  it("remains uncertain when the retry returns a typed 4xx", async () => {
    const firstError = new TypeError("response lost after commit");
    const retryError = new ApiError("Session expired", 401, {
      error: { message: "Unauthorized" },
    });
    postMock
      .mockRejectedValueOnce(firstError)
      .mockRejectedValueOnce(retryError);

    const action = createStoryAction(storyInput, "complexus");
    await expect(action).rejects.toBeInstanceOf(
      StoryCreationOutcomeUncertainError,
    );
    const uncertainty = await action.catch((error: unknown) => error);
    expect(((uncertainty as Error).cause as AggregateError).errors).toEqual([
      firstError,
      retryError,
    ]);
    expect(postMock).toHaveBeenCalledTimes(2);
  });

  it("does not retry an uncertain request without an idempotency key", async () => {
    const cause = new TypeError("response lost after commit");
    postMock.mockRejectedValue(cause);

    await expect(
      createStoryAction(
        { ...storyInput, idempotencyKey: undefined },
        "complexus",
      ),
    ).rejects.toBeInstanceOf(StoryCreationOutcomeUncertainError);
    expect(postMock).toHaveBeenCalledTimes(1);
  });

  it("does not classify authentication failure as a started mutation", async () => {
    const authError = new Error("Session lookup failed");
    authMock.mockRejectedValue(authError);

    await expect(createStoryAction(storyInput, "complexus")).rejects.toBe(
      authError,
    );
    expect(postMock).not.toHaveBeenCalled();
  });
});
