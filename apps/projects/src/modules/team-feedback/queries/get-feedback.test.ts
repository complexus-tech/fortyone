/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import {
  getTeamFeedbackMergeCandidates,
  getTeamFeedbackPrivateAuthor,
} from "./get-feedback";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
}));

const mockedGet = jest.mocked(get);
const ctx = {} as WorkspaceCtx;

describe("team feedback private author query", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("uses the admin-only private-author endpoint", async () => {
    const author = {
      avatarUrl: null,
      contributorId: "contributor-id",
      displayName: "Ada Ndlovu",
      email: "ada@example.com",
      kind: "verified_guest" as const,
      publicMasked: true,
      userId: null,
    };
    mockedGet.mockResolvedValue({ data: author });

    await expect(
      getTeamFeedbackPrivateAuthor("feedback-id", ctx),
    ).resolves.toBe(author);
    expect(mockedGet).toHaveBeenCalledWith(
      "feedback/items/feedback-id/private-author",
      ctx,
    );
  });

  it("does not hide authorization failures", async () => {
    mockedGet.mockResolvedValue({
      data: null,
      error: { message: "You don't have permission to access this resource" },
    });

    await expect(
      getTeamFeedbackPrivateAuthor("feedback-id", ctx),
    ).rejects.toThrow("You don't have permission");
  });
});

describe("team feedback merge candidate query", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("uses the admin-only server-side candidate search", async () => {
    const page = {
      candidates: [
        {
          commentCount: 2,
          id: "target-id",
          slug: "target-feedback",
          status: "completed" as const,
          title: "Target feedback",
          voteCount: 7,
        },
      ],
      hasMore: false,
    };
    mockedGet.mockResolvedValue({ data: page });

    await expect(
      getTeamFeedbackMergeCandidates("source-id", "  target idea  ", ctx),
    ).resolves.toBe(page);
    expect(mockedGet).toHaveBeenCalledWith(
      "feedback/items/source-id/merge-candidates?limit=30&search=target+idea",
      ctx,
    );
  });

  it("surfaces candidate authorization failures", async () => {
    mockedGet.mockResolvedValue({
      data: null,
      error: { message: "You don't have permission to merge feedback" },
    });

    await expect(
      getTeamFeedbackMergeCandidates("source-id", "", ctx),
    ).rejects.toThrow("You don't have permission");
  });
});
