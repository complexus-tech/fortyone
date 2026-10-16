/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen } from "@testing-library/react";
import type { CommentMayaRun, CommentMayaStatus } from "../types/comment-maya";
import { CommentMayaApproval } from "./comment-maya-approval";

const mockMutate = jest.fn();
jest.mock("../hooks/comment-maya", () => ({
  useCommentMayaAction: () => ({ mutate: mockMutate, isPending: false }),
}));
const RUN: CommentMayaRun = {
  id: "run",
  sourceId: "request",
  requesterId: "me",
  status: "awaiting_approval",
  summary: "Update the story?",
  proposalReview:
    "Status\nBefore: In progress\nAfter: Done\n\nDescription\nBefore: Original copy\nAfter: <script>Proposed copy</script>",
  updatedAt: "2026-10-16T12:00:00Z",
  version: 3,
};
beforeEach(() => mockMutate.mockClear());

describe("Maya reply approval", () => {
  it("shows complete escaped changes on demand and binds approval to the current version", () => {
    const { container, rerender } = render(
      <CommentMayaApproval currentUserId="me" run={RUN} storyId="story" />,
    );
    expect(container.querySelector("script")).toBeNull();
    expect(container.querySelector("details")).not.toHaveAttribute("open");
    expect(screen.queryByText(RUN.summary)).not.toBeInTheDocument();
    fireEvent.click(screen.getByText("Review proposed changes"));
    expect(
      screen.getByRole("region", { name: "Proposed changes" }),
    ).toHaveTextContent(
      "Before: Original copy After: <script>Proposed copy</script>",
    );
    fireEvent.click(screen.getByRole("button", { name: "Approve and apply" }));
    expect(mockMutate).toHaveBeenCalledWith({ run: RUN, action: "approve" });
    rerender(
      <CommentMayaApproval
        currentUserId="someone-else"
        run={RUN}
        storyId="story"
      />,
    );
    expect(
      screen.queryByRole("button", { name: "Approve and apply" }),
    ).not.toBeInTheDocument();
  });

  it.each([undefined, "", "   "])(
    "does not approve without complete review details (%s)",
    (proposalReview) => {
      render(
        <CommentMayaApproval
          currentUserId="me"
          run={{ ...RUN, proposalReview }}
          storyId="story"
        />,
      );
      const approve = screen.getByRole("button", { name: "Approve and apply" });
      expect(approve).toBeDisabled();
      fireEvent.click(approve);
      expect(mockMutate).not.toHaveBeenCalled();
      expect(
        screen.queryByText("Review proposed changes"),
      ).not.toBeInTheDocument();
    },
  );

  it.each<CommentMayaStatus>([
    "queued",
    "running",
    "needs_input",
    "succeeded",
    "failed",
    "cancelled",
  ])("renders no technical status or controls for %s", (status) => {
    const { container } = render(
      <CommentMayaApproval
        currentUserId="me"
        run={{ ...RUN, status }}
        storyId="story"
      />,
    );
    expect(container).toBeEmptyDOMElement();
  });
});
