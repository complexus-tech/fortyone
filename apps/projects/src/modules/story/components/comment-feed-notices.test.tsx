/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen, within } from "@testing-library/react";
import { CommentFeedNotices } from "./comment-feed-notices";

describe("comment feed notices", () => {
  it("shows target lookup feedback only while an unloaded target is requested", () => {
    const query = { isError: false, refetch: jest.fn() };
    const props = { commentsQuery: query, detailsQuery: query };
    const { container, rerender } = render(<CommentFeedNotices {...props} />);
    expect(container).toBeEmptyDOMElement();
    rerender(
      <CommentFeedNotices
        {...props}
        targetQuery={{ isPending: true, isError: false, refetch: jest.fn() }}
      />,
    );
    expect(screen.getByText("Finding comment…")).toBeVisible();
    rerender(
      <CommentFeedNotices
        {...props}
        targetQuery={{
          isPending: false,
          isError: true,
          error: { status: 404 },
          refetch: jest.fn(),
        }}
      />,
    );
    expect(screen.queryByText("Finding comment…")).not.toBeInTheDocument();
    expect(
      screen.getByText(
        "This comment is no longer available or you do not have access.",
      ),
    ).toBeVisible();
    rerender(<CommentFeedNotices {...props} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("retries the specific failed query without triggering the other retry", () => {
    const commentsQuery = { isError: true, refetch: jest.fn() };
    const detailsQuery = { isError: true, refetch: jest.fn() };
    render(
      <CommentFeedNotices
        commentsQuery={commentsQuery}
        detailsQuery={detailsQuery}
      />,
    );
    const details = screen.getByText(
      "Some comment details could not load.",
    ).parentElement!;
    fireEvent.click(within(details).getByRole("button", { name: "Retry" }));
    expect(detailsQuery.refetch).toHaveBeenCalledTimes(1);
    expect(commentsQuery.refetch).not.toHaveBeenCalled();
    const comments = screen.getByText(
      "Could not load comments.",
    ).parentElement!;
    fireEvent.click(within(comments).getByRole("button", { name: "Retry" }));
    expect(commentsQuery.refetch).toHaveBeenCalledTimes(1);
    expect(detailsQuery.refetch).toHaveBeenCalledTimes(1);
  });

  it("offers a targeted retry for a temporary linked-comment failure", () => {
    const query = { isError: false, refetch: jest.fn() };
    const targetQuery = {
      isError: true,
      isPending: false,
      error: new TypeError("Failed to fetch"),
      refetch: jest.fn(),
      isFetching: false,
    };
    const { rerender } = render(
      <CommentFeedNotices
        commentsQuery={query}
        detailsQuery={query}
        targetQuery={targetQuery}
      />,
    );
    expect(screen.getByText("Could not refresh this comment.")).toBeVisible();
    expect(
      screen.queryByText(
        "This comment is no longer available or you do not have access.",
      ),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(targetQuery.refetch).toHaveBeenCalledTimes(1);
    expect(query.refetch).not.toHaveBeenCalled();

    rerender(
      <CommentFeedNotices
        commentsQuery={query}
        detailsQuery={query}
        targetQuery={{ ...targetQuery, isFetching: true }}
      />,
    );
    expect(screen.getByRole("button", { name: "Retrying…" })).toBeDisabled();
  });
});
