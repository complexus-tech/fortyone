/* global beforeEach, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ButtonHTMLAttributes, ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { useStoryById } from "./hooks/story";
import { StoryPage } from ".";

jest.mock("ui", () => ({
  Box: "div",
  Button: ({ children, onClick }: ButtonHTMLAttributes<HTMLButtonElement>) => (
    <button onClick={onClick} type="button">
      {children}
    </button>
  ),
  Text: ({ children }: { children: ReactNode }) => <p>{children}</p>,
}));
jest.mock("next/dynamic", () => ({
  __esModule: true,
  default: () =>
    function MainDetails({
      mainHeader,
      storyId,
    }: {
      mainHeader?: ReactNode;
      storyId: string;
    }) {
      return (
        <main aria-label="Ticket details">
          {mainHeader}
          <input aria-label="Title" defaultValue={storyId} />
          <textarea aria-label="Description" defaultValue={storyId} />
        </main>
      );
    },
}));
jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => path }),
}));
jest.mock("@/shared/story/use-recent-story-history", () => ({
  useRecordStoryVisit: jest.fn(),
}));
jest.mock("@/components/ui/illustrations/empty-state-illustrations", () => ({
  NotFoundIllustration: () => null,
}));
jest.mock("@/components/ui/resource-not-found-state", () => ({
  ResourceNotFoundState: () => <p>Item not found</p>,
}));
jest.mock("./components/main-details-skeleton", () => ({
  MainDetailsSkeleton: () => null,
}));
jest.mock("./components/options", () => ({ Options: () => null }));
jest.mock("./components/story-skeleton", () => ({
  StorySkeleton: () => <p>Loading item</p>,
}));
jest.mock("./hooks/story", () => ({ useStoryById: jest.fn() }));

const mockedUseStoryById = jest.mocked(useStoryById);
const mockRefetch = jest.fn();
const ready = {
  data: { id: "story", deletedAt: null },
  isError: false,
  isFetching: false,
  isPending: false,
  refetch: mockRefetch,
} as unknown as ReturnType<typeof useStoryById>;

beforeEach(() => {
  mockRefetch.mockClear();
  mockedUseStoryById.mockReturnValue(ready);
});

it.each([
  ["full page", {}],
  ["dialog", { isDialog: true }],
  ["notification", { isNotifications: true }],
])(
  "mounts one editor pair for a %s and preserves it across updates",
  (_, props) => {
    const { rerender } = render(
      <StoryPage
        {...props}
        mainHeader={<button type="button">Close item</button>}
        storyId="story"
      />,
    );

    expect(
      screen.getAllByRole("main", { name: "Ticket details" }),
    ).toHaveLength(1);
    expect(screen.getAllByRole("textbox")).toHaveLength(2);
    const title = screen.getByRole("textbox", { name: "Title" });
    fireEvent.change(title, { target: { value: "Unsaved title" } });
    rerender(
      <StoryPage
        {...props}
        mainHeader={<button type="button">Close item</button>}
        storyId="story"
      />,
    );

    expect(screen.getByRole("textbox", { name: "Title" })).toBe(title);
    expect(title).toHaveValue("Unsaved title");
    expect(screen.getAllByRole("button", { name: "Close item" })).toHaveLength(
      1,
    );
  },
);

it("does not mount editors while loading or when the item is unavailable", () => {
  mockedUseStoryById.mockReturnValue({
    ...ready,
    data: undefined,
    isPending: true,
  } as ReturnType<typeof useStoryById>);
  const { rerender } = render(<StoryPage storyId="story" />);
  expect(screen.getByText("Loading item")).toBeInTheDocument();
  expect(screen.queryByRole("textbox")).not.toBeInTheDocument();

  mockedUseStoryById.mockReturnValue({ ...ready, data: null } as ReturnType<
    typeof useStoryById
  >);
  rerender(<StoryPage storyId="story" />);
  expect(screen.getByText("Item not found")).toBeInTheDocument();
  expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
});

it("preserves the error retry path without mounting editors", () => {
  mockedUseStoryById.mockReturnValue({
    ...ready,
    data: undefined,
    isError: true,
  } as ReturnType<typeof useStoryById>);
  render(<StoryPage storyId="story" />);
  fireEvent.click(screen.getByRole("button", { name: "Try again" }));
  expect(mockRefetch).toHaveBeenCalledTimes(1);
  expect(screen.queryByRole("textbox")).not.toBeInTheDocument();
});
