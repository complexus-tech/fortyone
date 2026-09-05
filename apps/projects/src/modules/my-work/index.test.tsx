/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { formatISO } from "date-fns";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import { getGroupedStoryFilterParams } from "@/components/ui/stories-filter-query";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { StoriesFilter } from "@/components/ui/stories-filter-types";
import type { MyWorkLayout } from "./types";
import { ListMyStories } from ".";

let mockAttention: string | null = null;
const mockStatuses = [
  { id: "active", category: "started" },
  { id: "done-team-a", category: "completed" },
  { id: "done-team-b", category: "completed" },
  { id: "cancelled", category: "cancelled" },
];
let mockStatusesData: typeof mockStatuses | undefined = mockStatuses;
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({
    data: mockStatusesData,
    isError: false,
    refetch: jest.fn(),
  }),
}));

jest.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
  usePathname: () => "/workspace/my-work",
}));

jest.mock("sonner", () => ({
  toast: {
    success: jest.fn(),
  },
}));

jest.mock("nuqs", () => {
  const React = jest.requireActual("react");

  return {
    parseAsStringLiteral: () => ({
      withDefault: (defaultValue: string) => ({ defaultValue }),
    }),
    useQueryState: (key: string, parser: { defaultValue?: string }) =>
      React.useState(
        key === "attention" ? mockAttention : parser.defaultValue ?? null,
      ),
  };
});

jest.mock("@/hooks", () => {
  const { useLocalStorage } = jest.requireActual("@/hooks/local-storage");

  return {
    useLocalStorage,
    useMediaQuery: () => false,
  };
});

jest.mock("@/modules/stories/hooks/use-my-stories-grouped", () => ({
  useMyStoriesGrouped: () => ({
    data: undefined,
    isPending: false,
  }),
}));

jest.mock("./components/header", () => {
  const { useMyWork } = jest.requireActual("./components/provider");

  const Header = ({
    layout,
    setLayout,
  }: {
    layout: MyWorkLayout;
    setLayout: (value: MyWorkLayout) => void;
  }) => {
    const {
      setViewOptions,
      viewOptions,
      filters,
      resetFilters,
      attentionStatus,
    } = useMyWork();

    return (
      <header>
        <output aria-label="Filters">{JSON.stringify(filters)}</output>
        <output aria-label="Attention status">
          {attentionStatus ?? "ready"}
        </output>
        <button onClick={resetFilters} type="button">
          Clear filters
        </button>
        <output aria-label="Current view options">
          {layout}:{viewOptions.groupBy}
        </output>
        <button
          onClick={() => {
            setViewOptions({ ...viewOptions, groupBy: "none" });
          }}
          type="button"
        >
          Set none
        </button>
        <button
          onClick={() => {
            setViewOptions({ ...viewOptions, groupBy: "assignee" });
          }}
          type="button"
        >
          Set assignee
        </button>
        <button
          onClick={() => {
            setLayout("list");
          }}
          type="button"
        >
          List
        </button>
        <button
          onClick={() => {
            setLayout("kanban");
          }}
          type="button"
        >
          Kanban
        </button>
      </header>
    );
  };

  return { Header };
});

jest.mock("./components/list-my-work", () => ({
  ListMyWork: ({ layout }: { layout: MyWorkLayout }) => (
    <main data-layout={layout} />
  ),
}));

const createViewOptions = (
  groupBy: StoriesViewOptions["groupBy"],
): StoriesViewOptions => ({
  displayColumns: ["ID"],
  groupBy,
  orderBy: "created",
  orderDirection: "desc",
  showEmptyGroups: true,
  showSubStories: true,
});

describe("ListMyStories", () => {
  beforeEach(() => {
    localStorage.clear();
    mockAttention = null;
    mockStatusesData = mockStatuses;
  });

  it("keeps list and Kanban view options isolated across repeated switches", () => {
    localStorage.setItem("my-stories:stories:layout", JSON.stringify("list"));
    localStorage.setItem(
      "my-work:view-options:list",
      JSON.stringify(createViewOptions("assignee")),
    );
    localStorage.setItem(
      "my-work:view-options:kanban",
      JSON.stringify(createViewOptions("priority")),
    );

    render(<ListMyStories />);

    expect(screen.getByLabelText("Current view options")).toHaveTextContent(
      "list:assignee",
    );

    fireEvent.click(screen.getByRole("button", { name: "Set none" }));
    fireEvent.click(screen.getByRole("button", { name: "Kanban" }));

    expect(screen.getByLabelText("Current view options")).toHaveTextContent(
      "kanban:priority",
    );

    fireEvent.click(screen.getByRole("button", { name: "Set assignee" }));
    fireEvent.click(screen.getByRole("button", { name: "List" }));

    expect(screen.getByLabelText("Current view options")).toHaveTextContent(
      "list:none",
    );

    fireEvent.click(screen.getByRole("button", { name: "Kanban" }));

    expect(screen.getByLabelText("Current view options")).toHaveTextContent(
      "kanban:assignee",
    );
  });
  it.each(["overdue", "today"])(
    "applies and persists editable %s filters after statuses load",
    async (view) => {
      mockAttention = view;
      mockStatusesData = undefined;
      localStorage.setItem(
        "stories:filters:/workspace/my-work",
        JSON.stringify({
          ...DEFAULT_STORIES_FILTER,
          priorities: ["urgent"],
        }),
      );
      const { rerender, unmount } = render(<ListMyStories />);
      expect(screen.getByLabelText("Attention status")).toHaveTextContent(
        "pending",
      );
      mockStatusesData = mockStatuses;
      rerender(<ListMyStories />);
      await waitFor(() => {
        expect(screen.getByLabelText("Attention status")).toHaveTextContent(
          "ready",
        );
      });
      const expected: StoriesFilter = {
        ...DEFAULT_STORIES_FILTER,
        statusIds: ["done-team-a", "done-team-b", "cancelled"],
        endDate: formatISO(new Date(), { representation: "date" }),
        operators: {
          endDate: view === "overdue" ? "isOnOrBefore" : "is",
          statusIds: "isNotAnyOf",
        },
      };
      expect(JSON.parse(screen.getByLabelText("Filters").textContent)).toEqual(
        expected,
      );
      expect(
        JSON.parse(localStorage.getItem("stories:filters:/workspace/my-work")!),
      ).toEqual(expected);
      expect(getGroupedStoryFilterParams(expected)).toMatchObject({
        excludedStatusIds: expected.statusIds,
        deadlineBefore: expected.endDate,
        deadlineAfter: view === "today" ? expected.endDate : undefined,
        priorities: undefined,
      });
      unmount();
      mockAttention = null;
      render(<ListMyStories />);
      expect(JSON.parse(screen.getByLabelText("Filters").textContent)).toEqual(
        expected,
      );
      fireEvent.click(screen.getByRole("button", { name: "Clear filters" }));
      expect(JSON.parse(screen.getByLabelText("Filters").textContent)).toEqual(
        DEFAULT_STORIES_FILTER,
      );
      expect(
        localStorage.getItem("stories:filters:/workspace/my-work"),
      ).toBeNull();
    },
  );
});
