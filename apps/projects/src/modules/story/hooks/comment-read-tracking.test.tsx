/* global beforeEach, afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, render } from "@testing-library/react";
import type { StoryCommentState } from "../types/comment-collaboration";
import {
  useCommentReadTracking,
  useCommentThreadReadTracking,
} from "./comment-read-tracking";

const mockMarkRead = jest.fn();
jest.mock("./comment-collaboration", () => ({
  useMarkCommentsRead: () => ({ mutate: mockMarkRead }),
}));
let intersect: (entries: { isIntersecting: boolean }[]) => void;
const savedObserver = global.IntersectionObserver;
const baseline: StoryCommentState = {
  lastReadAt: "2026-10-15T12:00:00Z",
  observedAt: "2026-10-16T12:00:00Z",
  unreadCount: 2,
  latestCommentId: "latest",
  comments: [],
};

const Tracking = ({
  state = baseline,
  latestLoadedAt = "2026-10-16T11:00:00Z",
}: {
  state?: StoryCommentState;
  latestLoadedAt?: string;
}) => {
  const { containerRef, lastReadAt } = useCommentReadTracking(
    "story",
    state,
    latestLoadedAt,
  );
  return <div data-last-read={lastReadAt} ref={containerRef} />;
};

const ThreadTracking = ({
  expanded = false,
  latest = "2026-10-16T11:00:00Z",
  readAt,
}: {
  expanded?: boolean;
  latest?: string;
  readAt?: string;
}) => {
  const ref = useCommentThreadReadTracking(
    "story",
    "root-thread",
    expanded,
    latest,
    "2026-10-16T13:00:00Z",
    readAt,
  );
  return <div hidden={!expanded} ref={ref} />;
};

beforeEach(() => {
  mockMarkRead.mockReset();
  jest.spyOn(document, "hasFocus").mockReturnValue(true);
  global.IntersectionObserver = jest
    .fn()
    .mockImplementation((callback: typeof intersect) => {
      intersect = callback;
      return { observe: jest.fn(), disconnect: jest.fn() };
    });
});
afterEach(() => {
  jest.restoreAllMocks();
  global.IntersectionObserver = savedObserver;
});

describe("comment read tracking", () => {
  it("reads only an opened visible thread, then preserves its receipt through collapse and revisit", () => {
    const { rerender, unmount } = render(<ThreadTracking />);
    expect(global.IntersectionObserver).not.toHaveBeenCalled();
    expect(mockMarkRead).not.toHaveBeenCalled();
    rerender(<ThreadTracking expanded />);
    act(() => {
      intersect([{ isIntersecting: false }]);
    });
    expect(mockMarkRead).not.toHaveBeenCalled();
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith({
      threadId: "root-thread",
      observedAt: "2026-10-16T11:00:00Z",
    });
    rerender(<ThreadTracking readAt="2026-10-16T11:00:00Z" />);
    rerender(<ThreadTracking expanded readAt="2026-10-16T11:00:00Z" />);
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
    unmount();
    render(<ThreadTracking expanded readAt="2026-10-16T11:00:00Z" />);
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
  });

  it("leaves a newly arriving reply unread while its thread is collapsed", () => {
    const { rerender } = render(
      <ThreadTracking
        readAt="2026-10-16T11:00:00Z"
        latest="2026-10-16T12:00:00Z"
      />,
    );
    expect(mockMarkRead).not.toHaveBeenCalled();
    rerender(
      <ThreadTracking
        expanded
        readAt="2026-10-16T11:00:00Z"
        latest="2026-10-16T12:00:00Z"
      />,
    );
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith({
      threadId: "root-thread",
      observedAt: "2026-10-16T12:00:00Z",
    });
  });
  it("waits for the comments to be visible and never reads beyond the loaded list", () => {
    render(<Tracking />);
    expect(mockMarkRead).not.toHaveBeenCalled();
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith({
      observedAt: "2026-10-16T11:00:00Z",
    });
  });
  it("caps a newer loaded comment at the server snapshot", () => {
    render(<Tracking latestLoadedAt="2026-10-16T13:00:00Z" />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith({
      observedAt: baseline.observedAt,
    });
  });
  it("uses the current durable marker as it moves", () => {
    const { container, rerender } = render(<Tracking />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    rerender(
      <Tracking
        state={{
          ...baseline,
          lastReadAt: "2026-10-16T11:00:00Z",
          unreadCount: 0,
        }}
      />,
    );
    expect(container.firstChild).toHaveAttribute(
      "data-last-read",
      "2026-10-16T11:00:00Z",
    );
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
  });
  it("uses the saved read marker on revisit instead of retaining the prior visit's unread state", () => {
    const firstVisit = render(<Tracking />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
    firstVisit.unmount();

    const { container } = render(
      <Tracking
        state={{
          ...baseline,
          lastReadAt: "2026-10-16T11:00:00Z",
          unreadCount: 0,
        }}
      />,
    );
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(container.firstChild).toHaveAttribute(
      "data-last-read",
      "2026-10-16T11:00:00Z",
    );
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
  });
  it("retries the same cutoff after a failure when the next snapshot arrives", () => {
    const { rerender } = render(<Tracking />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
    rerender(
      <Tracking state={{ ...baseline, observedAt: "2026-10-16T13:00:00Z" }} />,
    );
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledTimes(2);
  });
  it("does not mark an unfocused window as read", () => {
    jest.spyOn(document, "hasFocus").mockReturnValue(false);
    render(<Tracking />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).not.toHaveBeenCalled();
  });
});
