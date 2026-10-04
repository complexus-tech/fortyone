import { fireEvent, render, screen } from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { SavedViewConfiguration } from "@/modules/work-presets/public/types";
import { TeamOptionsProvider, useTeamOptions } from "./provider";

jest.mock("@/hooks", () => ({
  useLocalStorage: jest.requireActual("@/hooks/local-storage").useLocalStorage,
}));
const mockSetFilters = jest.fn();
jest.mock("@/components/ui/stories-filter-state", () => ({
  useStoriesFilters: () => ({
    filters: {},
    setFilters: mockSetFilters,
    resetFilters: jest.fn(),
  }),
}));

const saved: SavedViewConfiguration = {
  version: 1,
  layout: "kanban",
  description: "My workspace delivery",
  icon: "star",
  scope: { kind: "my-work", tab: "created" },
  filters: { ...DEFAULT_STORIES_FILTER, assignedToMe: true },
  viewOptions: {
    groupBy: "priority",
    orderBy: "deadline",
    orderDirection: "asc",
    showEmptyGroups: false,
    showSubStories: true,
    displayColumns: ["Status"],
  },
};
const Probe = () => {
  const { applyView, viewOptions, viewMetadata } = useTeamOptions();
  return (
    <>
      <button
        onClick={() => {
          applyView(saved);
        }}
        type="button"
      >
        Load view
      </button>
      <output>{viewOptions.groupBy}</output>
      <output aria-label="View metadata">{JSON.stringify(viewMetadata)}</output>
    </>
  );
};
describe("saved team views", () => {
  beforeEach(() => {
    localStorage.clear();
    mockSetFilters.mockClear();
  });
  it("restores the destination layout without changing the prior layout's preference", () => {
    const original = { ...saved.viewOptions, groupBy: "status" };
    localStorage.setItem(
      "teams:stories:view-options:list",
      JSON.stringify(original),
    );
    const { rerender } = render(
      <TeamOptionsProvider layout="list">
        <Probe />
      </TeamOptionsProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Load view" }));
    expect(mockSetFilters).toHaveBeenCalledWith(saved.filters);
    expect(
      JSON.parse(localStorage.getItem("teams:stories:view-options:list")!),
    ).toEqual(original);
    rerender(
      <TeamOptionsProvider layout="kanban">
        <Probe />
      </TeamOptionsProvider>,
    );
    expect(screen.getByText("priority")).toBeInTheDocument();
  });
  it("preserves saved scope/icon/description across layout changes and drops them for another or unselected view", () => {
    const { rerender } = render(
      <TeamOptionsProvider layout="list" viewId="selected">
        <Probe />
      </TeamOptionsProvider>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Load view" }));
    const metadata = {
      scope: saved.scope,
      icon: saved.icon,
      description: saved.description,
    };
    expect(
      screen.getByRole("status", { name: "View metadata" }),
    ).toHaveTextContent(JSON.stringify(metadata));
    rerender(
      <TeamOptionsProvider layout="kanban" viewId="selected">
        <Probe />
      </TeamOptionsProvider>,
    );
    expect(
      screen.getByRole("status", { name: "View metadata" }),
    ).toHaveTextContent(JSON.stringify(metadata));
    rerender(
      <TeamOptionsProvider layout="kanban" viewId="different">
        <Probe />
      </TeamOptionsProvider>,
    );
    expect(
      screen.getByRole("status", { name: "View metadata" }),
    ).toHaveTextContent("{}");
    rerender(
      <TeamOptionsProvider layout="kanban" viewId={null}>
        <Probe />
      </TeamOptionsProvider>,
    );
    expect(
      screen.getByRole("status", { name: "View metadata" }),
    ).toHaveTextContent("{}");
  });
});
