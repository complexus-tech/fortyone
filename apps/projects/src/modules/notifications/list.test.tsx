/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import { useNotificationsInfinite } from "./hooks/notifications";
import { ListNotifications } from "./list";

const mockFetchNextPage = jest.fn();
const mockRefetch = jest.fn();

jest.mock("./hooks/notifications", () => ({
  useNotificationsInfinite: jest.fn(),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "story" }),
}));
jest.mock("nuqs", () => ({
  parseAsString: { withDefault: () => ({}) },
  useQueryState: () => ["", jest.fn()],
}));
jest.mock("react-intersection-observer-hook", () => ({
  useIntersectionObserver: () => [
    jest.fn(),
    { entry: { isIntersecting: true } },
  ],
}));
jest.mock("./header", () => ({ NotificationsHeader: () => null }));
jest.mock("./card", () => ({
  NotificationCard: ({ title }: { title: string }) => <div>{title}</div>,
}));

const setQuery = (overrides = {}) => {
  jest.mocked(useNotificationsInfinite).mockReturnValue({
    data: undefined,
    fetchNextPage: mockFetchNextPage,
    hasNextPage: false,
    isError: false,
    isFetchNextPageError: false,
    isFetching: false,
    isFetchingNextPage: false,
    isPending: false,
    refetch: mockRefetch,
    ...overrides,
  } as unknown as ReturnType<typeof useNotificationsInfinite>);
};

beforeEach(() => {
  jest.clearAllMocks();
  setQuery();
});

describe("notification loading errors", () => {
  it("shows a recoverable failure instead of a false empty inbox", () => {
    setQuery({ isError: true });
    render(<ListNotifications />);

    expect(screen.getByRole("alert")).toHaveTextContent(
      "Couldn't load notifications.",
    );
    expect(screen.queryByText("No notifications")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockRefetch).toHaveBeenCalledTimes(1);
    expect(mockFetchNextPage).not.toHaveBeenCalled();
  });

  it("keeps existing rows and retries only the failed next page", () => {
    setQuery({
      data: {
        pages: [
          {
            notifications: [
              { id: "notification-1", title: "Existing notification" },
            ],
          },
        ],
      },
      hasNextPage: true,
      isError: true,
      isFetchNextPageError: true,
    });
    render(<ListNotifications />);

    expect(screen.getByText("Existing notification")).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Couldn't load more notifications.",
    );
    expect(mockFetchNextPage).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Try again" }));
    expect(mockFetchNextPage).toHaveBeenCalledTimes(1);
    expect(mockRefetch).not.toHaveBeenCalled();
  });

  it("disables retry while the request is running", () => {
    setQuery({ isError: true, isFetching: true });
    render(<ListNotifications />);

    expect(screen.getByRole("button", { name: "Retrying..." })).toBeDisabled();
  });
});
