import type { ReactNode } from "react";
import { fireEvent, render, screen } from "@testing-library/react";
import { Tabs } from "ui";
import { Header } from "./header";

let mockTab = "overview";

jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team-1", objectiveId: "objective-1" }),
}));
jest.mock("nuqs", () => ({
  parseAsString: { withDefault: jest.fn() },
  useQueryState: () => [mockTab],
}));
jest.mock("react-hotkeys-hook", () => ({ useHotkeys: jest.fn() }));
jest.mock("@/components/shared", () => ({
  HeaderContainer: ({ children }: { children: ReactNode }) => (
    <header>{children}</header>
  ),
  MobileMenuButton: () => null,
}));
jest.mock("@/components/ui", () => ({
  LayoutSwitcher: () => <button type="button">Layout</button>,
  StoriesFilterButton: () => <button type="button">Filters</button>,
  StoriesViewOptionsButton: () => <button type="button">View options</button>,
  TeamColor: () => null,
}));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks/use-objective", () => ({
  useObjective: () => ({
    data: { id: "objective-1", name: "Launch MVP", createdBy: "owner-1" },
  }),
}));
jest.mock("@/hooks", () => ({
  useTerminology: () => ({ getTermDisplay: () => "Tickets" }),
}));
jest.mock("./provider", () => ({
  useObjectiveOptions: () => ({
    viewOptions: {},
    filters: {},
    setViewOptions: jest.fn(),
    setFilters: jest.fn(),
    resetFilters: jest.fn(),
  }),
}));
jest.mock("./header-actions", () => ({
  HeaderActions: () => <button type="button">Objective actions</button>,
}));

beforeEach(() => {
  mockTab = "overview";
});

it("keeps section navigation and actions in the breadcrumb header and switches the tab", () => {
  const onValueChange = jest.fn();
  render(
    <Tabs defaultValue="overview" onValueChange={onValueChange}>
      <Header layout="list" setLayout={jest.fn()} />
    </Tabs>,
  );

  const header = screen.getByRole("banner");
  expect(header).toContainElement(
    screen.getByRole("tablist", { name: "Objective sections" }),
  );
  expect(header).toContainElement(
    screen.getByRole("button", { name: "Objective actions" }),
  );
  expect(screen.getByRole("tab", { name: "Overview" })).toHaveAttribute(
    "aria-selected",
    "true",
  );
  fireEvent.mouseDown(screen.getByRole("tab", { name: "Tickets" }), {
    button: 0,
    ctrlKey: false,
  });
  expect(onValueChange).toHaveBeenCalledWith("stories");
  expect(screen.getByRole("tab", { name: "Tickets" })).toHaveAttribute(
    "aria-selected",
    "true",
  );
  expect(
    screen.queryByRole("button", { name: "Filters" }),
  ).not.toBeInTheDocument();
});

it("preserves the layout, filter and view controls on the tickets tab", () => {
  mockTab = "stories";
  render(
    <Tabs defaultValue="stories">
      <Header layout="kanban" setLayout={jest.fn()} />
    </Tabs>,
  );

  for (const name of [
    "Layout",
    "Filters",
    "View options",
    "Objective actions",
  ]) {
    expect(screen.getByRole("banner")).toContainElement(
      screen.getByRole("button", { name }),
    );
  }
});
