/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { HTMLAttributes, PropsWithChildren } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useSearch } from "@/modules/search/hooks/use-search";
import type { SearchResponse } from "@/modules/search/types";
import { WorkspaceSearchPage } from "./workspace-search-page";

const mockSetTab = jest.fn();
const mockSetQuery = jest.fn();
const mockUseQueryState = jest.fn();
const mockStoriesBoard = jest.fn();
const mockListObjectives = jest.fn();

jest.mock("nuqs", () => ({
  parseAsStringLiteral: () => ({
    withDefault: () => ({}),
  }),
  useQueryState: (...args: unknown[]) => mockUseQueryState(...args),
}));

jest.mock("ui", () => {
  let onValueChange: ((value: string) => void) | undefined;
  const Box = ({
    children,
    ...props
  }: PropsWithChildren<HTMLAttributes<HTMLDivElement>>) => (
    <div {...props}>{children}</div>
  );
  const Text = ({
    children,
    color: _color,
    fontSize: _fontSize,
    ...props
  }: PropsWithChildren<
    HTMLAttributes<HTMLParagraphElement> & {
      color?: string;
      fontSize?: string;
    }
  >) => <p {...props}>{children}</p>;
  const Tabs = Object.assign(
    ({
      children,
      onValueChange: nextOnValueChange,
      value,
    }: PropsWithChildren<{
      onValueChange: (value: string) => void;
      value: string;
    }>) => {
      onValueChange = nextOnValueChange;
      return <div data-active-tab={value}>{children}</div>;
    },
    {
      List: ({ children }: PropsWithChildren) => <div>{children}</div>,
      Panel: ({ children }: PropsWithChildren) => <div>{children}</div>,
      Tab: ({ children, value }: PropsWithChildren<{ value: string }>) => (
        <button
          onClick={() => {
            onValueChange?.(value);
          }}
          type="button"
        >
          {children}
        </button>
      ),
    },
  );

  return { Box, Tabs, Text };
});

jest.mock("@/components/ui/board-skeleton", () => ({
  BoardSkeleton: () => <div data-testid="search-loading" />,
}));

jest.mock("@/components/ui/illustrations/empty-state-illustrations", () => ({
  SearchEmptyIllustration: () => (
    <div data-testid="search-empty-illustration" />
  ),
}));

jest.mock("@/components/ui/stories-board", () => ({
  StoriesBoard: (props: unknown) => {
    mockStoriesBoard(props);
    return <div data-testid="stories-board" />;
  },
}));

jest.mock("@/modules/objectives/components/list-objectives", () => ({
  ListObjectives: (props: unknown) => {
    mockListObjectives(props);
    return <div data-testid="objectives-list" />;
  },
}));

jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: jest.fn(),
}));

jest.mock("@/modules/search/hooks/use-search", () => ({
  useSearch: jest.fn(),
}));

const searchResponse = {
  objectives: [{ id: "objective-1", name: "Reliable roadmap" }],
  page: 1,
  pageSize: 20,
  stories: [{ id: "story-1", title: "Ship search" }],
  totalObjectives: 1,
  totalPages: 1,
  totalStories: 1,
} as unknown as SearchResponse;

describe("WorkspaceSearchPage", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(useTerminology).mockReturnValue({
      getTermDisplay: (term: string, options?: { capitalize?: boolean }) =>
        options?.capitalize ? `${term}s` : term,
    } as ReturnType<typeof useTerminology>);
    mockUseQueryState.mockImplementation((key: string) =>
      key === "type" ? ["stories", mockSetTab] : ["", mockSetQuery],
    );
    jest.mocked(useSearch).mockReturnValue({
      data: undefined,
      isFetching: false,
    } as ReturnType<typeof useSearch>);
  });

  it("keeps the no-query guidance without composing feature views", () => {
    render(<WorkspaceSearchPage />);

    expect(screen.getByText("Search your workspace")).toBeVisible();
    expect(useSearch).toHaveBeenCalledWith({ query: "", type: "stories" });
    expect(mockStoriesBoard).not.toHaveBeenCalled();
    expect(mockListObjectives).not.toHaveBeenCalled();
  });

  it("passes one search result contract to the Stories and Objectives views", () => {
    mockUseQueryState.mockImplementation((key: string) =>
      key === "type" ? ["all", mockSetTab] : ["  roadmap  ", mockSetQuery],
    );
    jest.mocked(useSearch).mockReturnValue({
      data: searchResponse,
      isFetching: false,
    } as ReturnType<typeof useSearch>);

    render(<WorkspaceSearchPage />);

    expect(useSearch).toHaveBeenCalledWith({
      query: "roadmap",
      type: undefined,
    });
    expect(mockStoriesBoard.mock.calls[0]?.[0]).toEqual(
      expect.objectContaining({
        groupedStories: expect.objectContaining({
          groups: [
            expect.objectContaining({
              stories: searchResponse.stories,
              totalCount: 1,
            }),
          ],
        }),
        isInSearch: true,
        layout: "list",
      }),
    );
    expect(mockListObjectives.mock.calls[0]?.[0]).toEqual({
      isInSearch: true,
      objectives: searchResponse.objectives,
    });
  });

  it("keeps the tab selection in the search URL state", () => {
    render(<WorkspaceSearchPage />);

    fireEvent.click(screen.getByRole("button", { name: "objectiveTerms" }));

    expect(mockSetTab).toHaveBeenCalledWith("objectives");
  });

  it("keeps the existing list loading state while a query is in flight", () => {
    mockUseQueryState.mockImplementation((key: string) =>
      key === "type" ? ["stories", mockSetTab] : ["roadmap", mockSetQuery],
    );
    jest.mocked(useSearch).mockReturnValue({
      data: undefined,
      isFetching: true,
    } as ReturnType<typeof useSearch>);

    render(<WorkspaceSearchPage />);

    expect(screen.getByTestId("search-loading")).toBeVisible();
    expect(mockStoriesBoard).not.toHaveBeenCalled();
    expect(mockListObjectives).not.toHaveBeenCalled();
  });
});
