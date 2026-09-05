/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */
import { fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { Button } from "ui";
import { DEFAULT_STORIES_FILTER } from "./stories-filter-types";
import { StoriesFilterButton } from "./stories-filter-button";
import { StoriesFilterMenu } from "./stories-filter-bar";
import type {
  StoriesFilterEditorProps,
  StoriesFilterField,
} from "./stories-filter-bar/types";

Object.defineProperty(globalThis, "ResizeObserver", {
  configurable: true,
  value: class {
    observe = jest.fn();
    unobserve = jest.fn();
    disconnect = jest.fn();
  },
});
Element.prototype.scrollIntoView = jest.fn();

jest.mock("next/navigation", () => ({ useParams: () => ({}) }));
jest.mock("react-hotkeys-hook", () => ({ useHotkeys: jest.fn() }));
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [] }),
}));
jest.mock("@/modules/teams/public/client", () => ({
  useTeamSettings: () => ({}),
}));
jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({
    getTermDisplay: (term: string) =>
      term === "objectiveTerm" ? "Goal" : "Result",
  }),
}));
jest.mock("./story-status-icon", () => ({ StoryStatusIcon: () => <span /> }));
jest.mock("./stories-filter-bar/filter-chip", () => ({
  TitleFilterDialog: () => <div role="dialog">Content filter</div>,
}));
jest.mock("./stories-filter-bar", () => {
  const { StoriesFilterMenu } = jest.requireActual("./stories-filter-menu");
  const { StatusEditor } = jest.requireActual(
    "./stories-filter-bar/people-editors",
  );
  return {
    StoriesFilterMenu: (props: StoriesFilterEditorProps) => (
      <StoriesFilterMenu
        {...props}
        renderEditor={() => (
          <StatusEditor {...props} statuses={[{ id: "todo", name: "To Do" }]} />
        )}
      />
    ),
  };
});

const Harness = ({
  entry,
  hiddenFields = [],
}: {
  entry: "header" | "toolbar";
  hiddenFields?: StoriesFilterField[];
}) => {
  const [filters, setFilters] = useState(DEFAULT_STORIES_FILTER);
  if (entry === "header") {
    return (
      <StoriesFilterButton
        filters={filters}
        hiddenFields={hiddenFields}
        resetFilters={() => {
          setFilters(DEFAULT_STORIES_FILTER);
        }}
        setFilters={setFilters}
      />
    );
  }
  return (
    <StoriesFilterMenu
      filters={filters}
      hiddenFields={hiddenFields}
      setFilters={setFilters}
    >
      <Button>Add filter</Button>
    </StoriesFilterMenu>
  );
};

const openMenu = (name: string) => {
  fireEvent.pointerDown(screen.getByRole("button", { name }), {
    button: 0,
    ctrlKey: false,
  });
};

describe("shared filter menu", () => {
  it.each(["header", "toolbar"] as const)(
    "opens the complete menu directly from the %s",
    (entry) => {
      render(<Harness entry={entry} />);
      openMenu(entry === "header" ? "Filters" : "Add filter");
      expect(screen.queryByText("Apply Filters")).not.toBeInTheDocument();
      expect(
        screen.getAllByRole("menuitem").map((item) => item.textContent),
      ).toEqual([
        "Status",
        "Assignee",
        "Creator",
        "Content",
        "Priority",
        "Team",
        "Sprint",
        "Label",
        "Complexity",
        "Goal",
        "Result",
        "Start date",
        "End date",
      ]);
      fireEvent.change(screen.getByPlaceholderText("Add filter..."), {
        target: { value: "date" },
      });
      expect(
        screen.getAllByRole("menuitem").map((item) => item.textContent),
      ).toEqual(["Start date", "End date"]);
    },
  );

  it("preserves page-specific hidden fields", () => {
    render(
      <Harness entry="header" hiddenFields={["teamIds", "objectiveId"]} />,
    );
    openMenu("Filters");
    expect(
      screen.queryByRole("menuitem", { name: "Team" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("menuitem", { name: "Goal" }),
    ).not.toBeInTheDocument();
    expect(
      screen.getByRole("menuitem", { name: "End date" }),
    ).toBeInTheDocument();
  });

  it("opens the existing content editor from the header menu", () => {
    render(<Harness entry="header" />);
    openMenu("Filters");
    fireEvent.click(screen.getByRole("menuitem", { name: "Content" }));
    expect(screen.getByRole("dialog")).toHaveTextContent("Content filter");
  });
  it("applies a status through the existing submenu editor", () => {
    render(<Harness entry="header" />);
    openMenu("Filters");
    fireEvent.click(screen.getByRole("menuitem", { name: "Status" }));
    fireEvent.pointerDown(screen.getByRole("option", { name: "To Do 0" }), {
      button: 0,
    });
    fireEvent.click(screen.getByRole("option", { name: "To Do 0" }));
    expect(
      screen.getByRole("button", { name: "1 filter applied" }),
    ).toBeInTheDocument();
  });
});
