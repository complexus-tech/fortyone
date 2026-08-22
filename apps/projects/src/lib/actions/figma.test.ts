/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { auth } from "@/auth";
import { post } from "@/lib/http";
import { linkFigmaStoryAction } from "./figma";

jest.mock("@/auth", () => ({ auth: jest.fn() }));
jest.mock("@/lib/http", () => ({ post: jest.fn(), remove: jest.fn() }));
jest.mock("@/utils", () => ({
  getApiError: (error: Error) => ({
    data: null,
    error: { message: error.message },
  }),
}));

const authMock = jest.mocked(auth);
const postMock = jest.mocked(post);

describe("linkFigmaStoryAction", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    authMock.mockResolvedValue({} as Awaited<ReturnType<typeof auth>>);
  });

  it("returns the typed Figma link when metadata resolves", async () => {
    const figmaLink = { id: "figma-link-1" };
    postMock.mockResolvedValueOnce({ data: figmaLink });

    await expect(
      linkFigmaStoryAction(
        "workspace",
        "story-1",
        "https://www.figma.com/design/file-key/file-name",
      ),
    ).resolves.toEqual({
      data: { kind: "figma", link: figmaLink },
    });
    expect(postMock).toHaveBeenCalledTimes(1);
  });

  it("saves a generic story link when Figma is rate limited", async () => {
    const url = "https://www.figma.com/design/file-key/file-name";
    const genericLink = { id: "link-1", storyId: "story-1", url };
    postMock
      .mockResolvedValueOnce({
        data: null,
        error: {
          code: "rate_limited",
          message: "Figma rate limit reached; try again later",
        },
      })
      .mockResolvedValueOnce({ data: genericLink });

    await expect(
      linkFigmaStoryAction("workspace", "story-1", url, "Checkout flow"),
    ).resolves.toEqual({
      data: { kind: "generic", link: genericLink },
    });
    expect(postMock).toHaveBeenNthCalledWith(
      2,
      "links",
      { storyId: "story-1", title: "Checkout flow", url },
      expect.objectContaining({ workspaceSlug: "workspace" }),
    );
  });

  it("supports the currently deployed bad-request rate-limit payload", async () => {
    const url = "https://www.figma.com/design/file-key/file-name";
    postMock
      .mockResolvedValueOnce({
        data: null,
        error: {
          code: "bad_request",
          message: "Figma rate limit reached; try again in 108h49m39s",
        },
      })
      .mockResolvedValueOnce({
        data: { id: "link-1", storyId: "story-1", url },
      });

    const result = await linkFigmaStoryAction("workspace", "story-1", url);

    expect(result.data?.kind).toBe("generic");
    expect(postMock).toHaveBeenNthCalledWith(
      2,
      "links",
      { storyId: "story-1", title: "Figma design", url },
      expect.any(Object),
    );
  });

  it("does not downgrade non-rate-limit Figma failures", async () => {
    const response = {
      data: null,
      error: { code: "permission_denied", message: "Access denied" },
    };
    postMock.mockResolvedValueOnce(response);

    await expect(
      linkFigmaStoryAction(
        "workspace",
        "story-1",
        "https://www.figma.com/design/file-key/file-name",
      ),
    ).resolves.toEqual(response);
    expect(postMock).toHaveBeenCalledTimes(1);
  });
});
