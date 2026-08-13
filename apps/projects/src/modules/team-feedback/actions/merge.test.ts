/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post } from "@/lib/http";
import { mergeTeamFeedbackAction } from "./merge";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ post: jest.fn() }));
jest.mock("@/utils", () => ({
  getApiError: (error: Error) => ({
    data: null,
    error: { message: error.message },
  }),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);

describe("mergeTeamFeedbackAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it("posts the canonical target to the admin merge endpoint", async () => {
    const session = { user: { id: "admin-id" } } as Awaited<
      ReturnType<typeof auth>
    >;
    const response = {
      data: {
        sourceItemId: "source-id",
        targetItemId: "target-id",
      },
    };
    authMock.mockResolvedValue(session);
    postMock.mockResolvedValue(response);

    await expect(
      mergeTeamFeedbackAction("source/id", "target-id", "workspace"),
    ).resolves.toBe(response);
    expect(postMock).toHaveBeenCalledWith(
      "feedback/items/source%2Fid/merge",
      { targetItemId: "target-id" },
      { session, workspaceSlug: "workspace" },
    );
  });

  it("surfaces API failures through the shared action contract", async () => {
    authMock.mockResolvedValue({} as Awaited<ReturnType<typeof auth>>);
    postMock.mockRejectedValue(new Error("Merge conflict"));

    await expect(
      mergeTeamFeedbackAction("source-id", "target-id", "workspace"),
    ).resolves.toEqual({
      data: null,
      error: { message: "Merge conflict" },
    });
  });
});
