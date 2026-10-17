/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { act, render } from "@testing-library/react";
import { markCommentSeen } from "../queries/comment-collaboration";
import { useCommentSeen } from "./use-comment-seen";

jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "first" }),
}));
jest.mock("../queries/comment-collaboration", () => ({
  markCommentSeen: jest.fn(),
}));

const savedObserver = global.IntersectionObserver;
let intersect: IntersectionObserverCallback;
const Tracking = ({ enabled = true, id = "maya-reply" }) => {
  const ref = useCommentSeen("story", id, enabled);
  return <div ref={ref} />;
};
const show = (ratio = 1) => {
  act(() => {
    intersect(
      [
        {
          isIntersecting: ratio > 0,
          intersectionRatio: ratio,
        } as IntersectionObserverEntry,
      ],
      {} as IntersectionObserver,
    );
  });
};
const advance = async (milliseconds: number) => {
  await act(async () => {
    jest.advanceTimersByTime(milliseconds);
  });
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.mocked(markCommentSeen).mockReset().mockResolvedValue(undefined);
  jest.spyOn(document, "hasFocus").mockReturnValue(true);
  jest.spyOn(document, "visibilityState", "get").mockReturnValue("visible");
  global.IntersectionObserver = jest.fn().mockImplementation((callback) => {
    intersect = callback;
    return { observe: jest.fn(), disconnect: jest.fn() };
  });
});
afterEach(() => {
  jest.useRealTimers();
  jest.restoreAllMocks();
  global.IntersectionObserver = savedObserver;
});

describe("exact comment visibility receipts", () => {
  it("acknowledges only this reply after it remains visible for half a second", async () => {
    render(<Tracking />);
    show(0.1);
    await advance(1_000);
    expect(markCommentSeen).not.toHaveBeenCalled();
    show();
    await advance(499);
    expect(markCommentSeen).not.toHaveBeenCalled();
    await advance(1);
    expect(markCommentSeen).toHaveBeenCalledWith("story", "maya-reply", {
      workspaceSlug: "first",
    });
    show();
    await advance(5_000);
    expect(markCommentSeen).toHaveBeenCalledTimes(1);
  });

  it("does not acknowledge hidden or unfocused tabs", async () => {
    const focus = jest.spyOn(document, "hasFocus").mockReturnValue(false);
    const visibility = jest.spyOn(document, "visibilityState", "get");
    render(<Tracking />);
    show();
    await advance(1_000);
    focus.mockReturnValue(true);
    visibility.mockReturnValue("hidden");
    act(() => window.dispatchEvent(new Event("focus")));
    await advance(1_000);
    expect(markCommentSeen).not.toHaveBeenCalled();
    visibility.mockReturnValue("visible");
    act(() => document.dispatchEvent(new Event("visibilitychange")));
    await advance(500);
    expect(markCommentSeen).toHaveBeenCalledTimes(1);
  });

  it("cancels the dwell when the reply leaves the viewport or unmounts", async () => {
    const { unmount } = render(<Tracking />);
    show();
    await advance(200);
    show(0);
    await advance(500);
    expect(markCommentSeen).not.toHaveBeenCalled();
    show();
    unmount();
    await advance(500);
    expect(markCommentSeen).not.toHaveBeenCalled();
  });

  it("bounds failed retries and never retries after the window loses focus", async () => {
    jest.mocked(markCommentSeen).mockRejectedValue(new Error("Offline"));
    const focus = jest.spyOn(document, "hasFocus");
    render(<Tracking />);
    show();
    await advance(500);
    focus.mockReturnValue(false);
    act(() => window.dispatchEvent(new Event("blur")));
    await advance(2_000);
    expect(markCommentSeen).toHaveBeenCalledTimes(1);
    focus.mockReturnValue(true);
    act(() => window.dispatchEvent(new Event("focus")));
    await advance(1_000);
    await advance(1_000);
    await advance(5_000);
    expect(markCommentSeen).toHaveBeenCalledTimes(3);
  });

  it("does not track optimistic or human comments when disabled", () => {
    render(<Tracking enabled={false} />);
    expect(global.IntersectionObserver).not.toHaveBeenCalled();
  });

  it("starts a fresh receipt when navigating to a different comment", async () => {
    const { rerender } = render(<Tracking />);
    show();
    await advance(500);
    rerender(<Tracking id="another-reply" />);
    show();
    await advance(500);
    expect(markCommentSeen).toHaveBeenLastCalledWith("story", "another-reply", {
      workspaceSlug: "first",
    });
  });
});
