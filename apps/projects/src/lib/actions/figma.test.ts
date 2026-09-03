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

  it("returns the server-persisted generic link when Figma metadata is unavailable", async () => {
    const url = "https://www.figma.com/design/file-key/file-name";
    postMock.mockResolvedValueOnce({
      data: {
        id: "figma-link-1",
        storyId: "story-1",
        storyLinkId: "link-1",
        artifact: {
          canonicalUrl: url,
          fileName: "Figma design",
          nodeName: null,
        },
        unavailableAt: "2026-09-03T08:00:00Z",
        createdAt: "2026-09-03T08:00:00Z",
        updatedAt: "2026-09-03T08:00:00Z",
      },
    });

    await expect(
      linkFigmaStoryAction("workspace", "story-1", url, "Checkout flow"),
    ).resolves.toEqual({
      data: {
        kind: "generic",
        link: {
          id: "link-1",
          storyId: "story-1",
          title: "Checkout flow",
          url,
          createdAt: "2026-09-03T08:00:00Z",
          updatedAt: "2026-09-03T08:00:00Z",
        },
      },
    });
    expect(postMock).toHaveBeenCalledTimes(1);
    expect(postMock).toHaveBeenCalledWith(
      "stories/story-1/figma-links",
      { url },
      expect.objectContaining({ workspaceSlug: "workspace" }),
    );
  });

  it("uses the degraded Figma link title when no title is provided", async () => {
    const url = "https://www.figma.com/design/file-key/file-name";
    postMock.mockResolvedValueOnce({
      data: {
        id: "figma-link-1",
        storyId: "story-1",
        storyLinkId: "link-1",
        artifact: {
          canonicalUrl: url,
          fileName: "Figma design",
          nodeName: null,
        },
        unavailableAt: "2026-09-03T08:00:00Z",
        createdAt: "2026-09-03T08:00:00Z",
        updatedAt: "2026-09-03T08:00:00Z",
      },
    });

    const result = await linkFigmaStoryAction("workspace", "story-1", url);

    expect(result.data?.kind).toBe("generic");
    if (result.data?.kind === "generic") {
      expect(result.data.link.title).toBe("Figma design");
    }
    expect(postMock).toHaveBeenCalledTimes(1);
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
