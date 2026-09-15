import { render, screen } from "@testing-library/react";
import type { TeamFeedbackItem, TeamFeedbackPrivateAuthor } from "../types";
import { FeedbackProperties } from "./feedback-properties";

jest.mock("ui", () => ({ ...jest.requireActual("ui"), TimeAgo: () => null }));
jest.mock("../status", () => ({ FeedbackStatus: () => null }));
jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({ getTermDisplay: () => "task" }),
}));
jest.mock("@/hooks/media", () => ({ useMediaQuery: () => false }));

it("keeps Former user after admin identity loads with a masked public author ID", () => {
  const feedback = {
    authorId: null,
    authorName: "Former user",
    authorAvatar: null,
    storyLinks: [],
    board: { name: "Ideas", color: "#000000" },
    status: "pending",
    createdAt: "2026-09-15T12:00:00Z",
    upvoteCount: 2,
    downvoteCount: 0,
    commentCount: 1,
  } as unknown as TeamFeedbackItem;
  const privateAuthor: TeamFeedbackPrivateAuthor = {
    contributorId: "anonymous-bucket",
    userId: null,
    kind: "anonymous",
    displayName: "Former user",
    email: null,
    avatarUrl: null,
    publicMasked: false,
  };
  render(
    <FeedbackProperties feedback={feedback} privateAuthor={privateAuthor} />,
  );
  expect(screen.getByText("Former user")).toBeInTheDocument();
  expect(screen.queryByText("Anonymous")).not.toBeInTheDocument();
  expect(screen.queryByRole("link")).not.toBeInTheDocument();
  expect(screen.queryByText("Contact")).not.toBeInTheDocument();
});
