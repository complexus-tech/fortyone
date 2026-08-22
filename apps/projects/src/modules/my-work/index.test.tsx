/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { fireEvent, render, screen } from "@testing-library/react";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import type { MyWorkLayout } from "./types";
import { ListMyStories } from ".";

jest.mock("next/navigation", () => ({
  useSearchParams: () => new URLSearchParams(),
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
    useQueryState: (_key: string, parser: { defaultValue: string }) =>
      React.useState(parser.defaultValue),
  };
});

jest.mock("@/hooks", () => {
  const { useLocalStorage } = jest.requireActual("@/hooks/local-storage");

  return {
    useLocalStorage,
    useMediaQuery: () => false,
  };
});

jest.mock("@/components/ui/stories-filter-state", () => ({
  useStoriesFilters: () => ({
    filters: {},
    resetFilters: jest.fn(),
    setFilters: jest.fn(),
  }),
}));

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
    const { setViewOptions, viewOptions } = useMyWork();

    return (
      <header>
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
});
