/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen } from "@testing-library/react";
import { UnreadThreads } from "./unread-threads";

const mockThreads = {
  data: {
    pages: [
      {
        items: [
          {
            id: "old-root",
            conversationId: "older-team",
            conversationName: "Product",
            conversationKind: "team",
            text: "An older conversation",
            author: { fullName: "Sarah" },
            unreadReplyCount: 100,
            firstUnreadReplyId: "first-unread",
          },
        ],
      },
    ],
  },
  isPending: false,
  error: null as Error | null,
  hasNextPage: true,
  isFetchingNextPage: false,
  fetchNextPage: jest.fn(),
  refetch: jest.fn(),
};
const mockUseUnreadThreads = jest.fn((_enabled: boolean) => mockThreads);
jest.mock("./queries", () => ({
  useUnreadThreads: (enabled: boolean) => mockUseUnreadThreads(enabled),
}));
jest.mock("icons", () => ({ ChevronRightIcon: () => null }));

beforeEach(() => {
  mockUseUnreadThreads.mockClear();
  mockThreads.fetchNextPage.mockClear();
  mockThreads.refetch.mockClear();
  mockThreads.error = null;
});

describe("UnreadThreads", () => {
  it("loads only when expanded and opens an old root with its conversation context", () => {
    const onSelect = jest.fn();
    render(<UnreadThreads count={2} onSelect={onSelect} />);
    expect(mockUseUnreadThreads).toHaveBeenLastCalledWith(false);
    fireEvent.click(screen.getByRole("button", { name: /Threads/ }));
    expect(mockUseUnreadThreads).toHaveBeenLastCalledWith(true);
    const thread = screen.getByRole("button", {
      name: /An older conversation/,
    });
    expect(thread).toHaveTextContent("100");
    fireEvent.click(thread);
    expect(onSelect).toHaveBeenCalledWith(
      expect.objectContaining({ id: "old-root", conversationId: "older-team" }),
      thread,
    );
    fireEvent.click(screen.getByRole("button", { name: "Load more threads" }));
    expect(mockThreads.fetchNextPage).toHaveBeenCalledTimes(1);
  });

  it("keeps failed unread-thread loading recoverable", () => {
    mockThreads.error = new Error("Offline");
    render(<UnreadThreads count={2} onSelect={jest.fn()} />);
    fireEvent.click(screen.getByRole("button", { name: /Threads/ }));
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Could not load unread threads",
    );
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockThreads.refetch).toHaveBeenCalledTimes(1);
  });
});
