/* global beforeEach, afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, render } from "@testing-library/react";
import type { StoryCommentState } from "../types/comment-collaboration";
import { useCommentReadTracking } from "./comment-read-tracking";

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
  const { containerRef, lastReadAt, hadUnread } = useCommentReadTracking(
    "story",
    state,
    latestLoadedAt,
  );
  return (
    <div
      data-had-unread={hadUnread}
      data-last-read={lastReadAt}
      ref={containerRef}
    />
  );
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
  it("waits for the comments to be visible and never reads beyond the loaded list", () => {
    render(<Tracking />);
    expect(mockMarkRead).not.toHaveBeenCalled();
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith(
      "2026-10-16T11:00:00Z",
      expect.any(Object),
    );
  });
  it("caps a newer loaded comment at the server snapshot", () => {
    render(<Tracking latestLoadedAt="2026-10-16T13:00:00Z" />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    expect(mockMarkRead).toHaveBeenCalledWith(
      baseline.observedAt,
      expect.any(Object),
    );
  });
  it("preserves the visit's unread boundary after the durable marker moves", () => {
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
      baseline.lastReadAt,
    );
    expect(container.firstChild).toHaveAttribute("data-had-unread", "true");
    expect(mockMarkRead).toHaveBeenCalledTimes(1);
  });
  it("retries the same cutoff after a failure when the next snapshot arrives", () => {
    const { rerender } = render(<Tracking />);
    act(() => {
      intersect([{ isIntersecting: true }]);
    });
    const callbacks = mockMarkRead.mock.calls[0][1] as { onError: () => void };
    act(() => {
      callbacks.onError();
    });
    rerender(<Tracking state={{ ...baseline }} />);
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
