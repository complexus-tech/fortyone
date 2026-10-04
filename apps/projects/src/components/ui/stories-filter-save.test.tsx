import { fireEvent, render, screen } from "@testing-library/react";
import type { ReactNode } from "react";
import { SaveViewProvider } from "@/shared/views/save-slot";
import type { SavedViewConfiguration } from "@/shared/story/view-configuration";
import { DEFAULT_STORIES_FILTER } from "./stories-filter-types";
import { StoriesFilterBar } from "./stories-filter-bar";

jest.mock("next/navigation", () => ({ useParams: () => ({}) }));
jest.mock("@/lib/hooks/statuses", () => ({
  useStatuses: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/members", () => ({ useMembers: () => ({ data: [] }) }));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: [] }),
}));
jest.mock("@/lib/hooks/labels", () => ({ useLabels: () => ({ data: [] }) }));
jest.mock("@/modules/teams/hooks/teams", () => ({
  useTeams: () => ({ data: [] }),
}));
jest.mock("@/modules/teams/hooks/use-team-settings", () => ({
  useTeamSettings: () => ({}),
}));
jest.mock("@/modules/objectives/hooks/use-objectives", () => ({
  useTeamObjectives: () => ({ data: [] }),
}));
jest.mock("@/modules/objectives/hooks", () => ({
  useKeyResults: () => ({ data: [] }),
}));
jest.mock("@/modules/sprints/hooks/team-sprints", () => ({
  useTeamSprints: () => ({ data: [] }),
}));
jest.mock("@/hooks/use-terminology-display", () => ({
  useTerminology: () => ({ getTermDisplay: () => "task" }),
}));
jest.mock("./stories-filter-menu", () => ({
  StoriesFilterMenu: ({ children }: { children: ReactNode }) => children,
}));

const viewOptions: SavedViewConfiguration["viewOptions"] = {
  groupBy: "status",
  orderBy: "created",
  orderDirection: "desc",
  showEmptyGroups: true,
  showSubStories: false,
  displayColumns: ["Status"],
};
const SaveButton = () => <button type="button">Save as</button>;
const reset = jest.fn();
const harness = (
  filters = DEFAULT_STORIES_FILTER,
  scope?: SavedViewConfiguration["scope"],
) => (
  <SaveViewProvider SaveView={SaveButton}>
    <StoriesFilterBar
      filters={filters}
      resetFilters={reset}
      saveView={{
        configuration: {
          version: 1,
          layout: "list",
          filters,
          viewOptions,
          ...(scope ? { scope } : {}),
        },
      }}
      setFilters={jest.fn()}
    />
  </SaveViewProvider>
);

beforeEach(() => reset.mockClear());
it.each([
  { contentContains: "response" },
  { assignedToMe: true },
  { createdByMe: true },
  { completedAfter: "2026-10-01" },
  { isNotCompleted: true },
])("shows Save as before Clear all for applied filter %j", (patch) => {
  render(harness({ ...DEFAULT_STORIES_FILTER, ...patch }));
  const save = screen.getByRole("button", { name: "Save as" });
  const clear = screen.getByRole("button", { name: "Clear all" });
  expect(save.compareDocumentPosition(clear)).toBe(
    Node.DOCUMENT_POSITION_FOLLOWING,
  );
  fireEvent.click(clear);
  expect(reset).toHaveBeenCalledTimes(1);
});
it("does not show the save toolbar for an unfiltered scope", () => {
  render(harness());
  expect(
    screen.queryByRole("button", { name: "Save as" }),
  ).not.toBeInTheDocument();
});
it("shows Save as for category/date/overdue scope even with no explicit filter chips", () => {
  render(
    harness(DEFAULT_STORIES_FILTER, {
      kind: "my-work",
      tab: "all",
      overdue: true,
      category: "started",
      createdAfter: "2026-10-01",
    }),
  );
  expect(screen.getByRole("button", { name: "Save as" })).toBeInTheDocument();
  expect(screen.getByRole("button", { name: "Clear all" })).toBeInTheDocument();
});
