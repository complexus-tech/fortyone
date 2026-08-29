/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { get } from "@/lib/http";
import type { WorkspaceCtx } from "@/lib/http";
import { getFeedbackUpdateCandidates } from "./queries";

jest.mock("@/lib/http", () => ({
  get: jest.fn(),
}));

const mockedGet = jest.mocked(get);
const ctx = {} as WorkspaceCtx;

describe("feedback Update candidate query", () => {
  beforeEach(() => {
    mockedGet.mockReset();
  });

  it("searches canonical portal feedback across every status", async () => {
    const page = {
      candidates: [
        {
          commentCount: 4,
          id: "completed-feedback-id",
          slug: "completed-feedback",
          status: "completed",
          title: "Completed feedback",
          voteCount: 12,
        },
      ],
      hasMore: false,
    };
    mockedGet.mockResolvedValue({ data: page });

    await expect(
      getFeedbackUpdateCandidates("portal-id", "  shipped  ", ctx),
    ).resolves.toBe(page);
    expect(mockedGet).toHaveBeenCalledWith(
      "feedback/portals/portal-id/item-candidates?limit=30&search=shipped",
      ctx,
    );
  });

  it("surfaces authorization and API failures", async () => {
    mockedGet.mockResolvedValue({
      data: null,
      error: { message: "You don't have permission to manage Updates" },
    });

    await expect(
      getFeedbackUpdateCandidates("portal-id", "", ctx),
    ).rejects.toThrow("You don't have permission");
  });
});
