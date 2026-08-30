/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getFeedbackSubmitLabel,
  toSimilarFeedbackRequest,
} from "./feedback-controls";

describe("public feedback control utilities", () => {
  it("uses the duplicate flow before transient submit state", () => {
    expect(
      getFeedbackSubmitLabel({
        hasDuplicate: true,
        isSubmitting: true,
        requiresIdentity: true,
      }),
    ).toBe("View existing feedback");
    expect(
      getFeedbackSubmitLabel({
        hasDuplicate: false,
        isSubmitting: false,
        requiresIdentity: true,
      }),
    ).toBe("Continue");
  });

  it("creates a safe vote target from a similar feedback result", () => {
    expect(
      toSimilarFeedbackRequest({
        authorAvatar: null,
        authorId: null,
        authorName: "",
        commentCount: 3,
        confidence: 0.98,
        id: "request-1",
        isDuplicate: true,
        slug: "crossing-request",
        title: "Add a crossing",
        voteCount: 12,
      }),
    ).toEqual(
      expect.objectContaining({
        authorName: "Unknown contributor",
        boardId: "",
        comments: [],
        status: "pending",
        storyLinks: [],
        voteCount: 12,
      }),
    );
  });
});
