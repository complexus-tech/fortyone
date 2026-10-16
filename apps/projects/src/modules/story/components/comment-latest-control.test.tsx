/* global beforeEach, afterEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { useRef } from "react";
import { act, fireEvent, render, screen } from "@testing-library/react";
import { CommentLatestControl } from "./comment-latest-control";

let intersect: IntersectionObserverCallback;
let changed: MutationCallback;
const observe = jest.fn();
const disconnect = jest.fn();
const savedIntersectionObserver = global.IntersectionObserver;
const savedMutationObserver = global.MutationObserver;

const Feed = ({ expanded = true }: { expanded?: boolean }) => {
  const containerRef = useRef<HTMLDivElement>(null);
  return (
    <div ref={containerRef}>
      <CommentLatestControl
        commentId="latest"
        containerRef={containerRef}
        onJump={jump}
      />
      {expanded ? (
        <div id="comment-latest">
          <div data-comment-anchor>Latest author</div>
          <p>Latest comment</p>
        </div>
      ) : null}
    </div>
  );
};
const jump = jest.fn();
const reportVisibility = (visible: boolean) => {
  const target = document.querySelector("[data-comment-anchor]");
  act(() => {
    intersect(
      [{ target, isIntersecting: visible }] as IntersectionObserverEntry[],
      {} as IntersectionObserver,
    );
  });
};
const reportDOMChange = () => {
  act(() => {
    changed([], {} as MutationObserver);
  });
};

beforeEach(() => {
  jest.clearAllMocks();
  global.IntersectionObserver = jest.fn().mockImplementation((callback) => {
    intersect = callback;
    return { observe, disconnect };
  });
  global.MutationObserver = jest.fn().mockImplementation((callback) => {
    changed = callback;
    return { observe: jest.fn(), disconnect: jest.fn() };
  });
});
afterEach(() => {
  global.IntersectionObserver = savedIntersectionObserver;
  global.MutationObserver = savedMutationObserver;
});

describe("latest comment navigation", () => {
  it("only offers the jump while the latest comment is outside the viewport", () => {
    render(<Feed />);
    expect(observe).toHaveBeenCalledWith(screen.getByText("Latest author"));
    reportVisibility(true);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    reportVisibility(false);
    fireEvent.click(screen.getByRole("button", { name: "Jump to latest" }));
    expect(jump).toHaveBeenCalledTimes(1);
    reportVisibility(true);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
  });

  it("finds a newly expanded or fetched reply and hides the control once it is visible", () => {
    const { rerender } = render(<Feed expanded={false} />);
    expect(
      screen.getByRole("button", { name: "Jump to latest" }),
    ).toBeInTheDocument();
    expect(observe).not.toHaveBeenCalled();
    rerender(<Feed />);
    reportDOMChange();
    reportVisibility(true);
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    rerender(<Feed expanded={false} />);
    reportDOMChange();
    expect(
      screen.getByRole("button", { name: "Jump to latest" }),
    ).toBeInTheDocument();
  });

  it("disconnects its observers when the feed closes", () => {
    const { unmount } = render(<Feed />);
    disconnect.mockClear();
    unmount();
    expect(disconnect).toHaveBeenCalledTimes(1);
  });
});
