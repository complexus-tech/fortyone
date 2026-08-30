/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { PublicRequest } from "@/shared/feedback-widget/types";
import { getFeedbackEmptyBody, mergeRequests, replaceRequest } from "./utils";

const request = (id: string, title = id): PublicRequest => ({
  authorAvatar: null,
  authorId: null,
  authorName: "Amina",
  boardId: "board-1",
  commentCount: 0,
  comments: [],
  createdAtLabel: "Just now",
  description: "A request",
  id,
  slug: id,
  status: "pending",
  storyLinks: [],
  title,
  voteCount: 1,
});

describe("feedback widget view utilities", () => {
  it("merges a later roadmap page without duplicating an existing request", () => {
    const first = request("first", "First version");
    const merged = mergeRequests(
      [first, request("second")],
      [request("first", "Updated version"), request("third")],
    );

    expect(merged.map((item) => item.id)).toEqual(["first", "second", "third"]);
    expect(merged[0]?.title).toBe("Updated version");
  });

  it("keeps each empty feedback message tied to its visible filter state", () => {
    expect(getFeedbackEmptyBody("crossing", "active")).toContain(
      "different phrase",
    );
    expect(getFeedbackEmptyBody("", "completed")).toContain("status");
    expect(getFeedbackEmptyBody("", "active")).toContain("New suggestions");
  });

  it("updates every visible copy of a request without mutating unrelated rows", () => {
    const original = request("first", "Original");
    const unchanged = request("second", "Unchanged");
    const updated = request("first", "Updated");
    const current = [original, unchanged];

    const next = replaceRequest(current, updated);

    expect(next).toEqual([updated, unchanged]);
    expect(next).not.toBe(current);
    expect(original.title).toBe("Original");
  });
});
