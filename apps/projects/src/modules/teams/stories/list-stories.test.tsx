import type { HTMLAttributes } from "react";
import type { Root } from "react-dom/client";
import { act, fireEvent, within } from "@testing-library/react";
import { hydrateRoot } from "react-dom/client";
import { renderToString } from "react-dom/server";
import type { StoriesViewOptions } from "@/components/ui/stories-view-options-button";
import type { StoriesLayout } from "@/components/ui";
import type * as TeamOptionsProviderModule from "./provider";
import { ListStories } from "./list-stories";

jest.mock("@/hooks", () => ({
  useLocalStorage: jest.requireActual("@/hooks/local-storage").useLocalStorage,
}));
jest.mock("next/navigation", () => ({
  useParams: () => ({ teamId: "team" }),
  useSearchParams: () => new URLSearchParams(),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: "user" } } }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "acme" }),
}));
jest.mock("ui", () => ({
  Box: ({ children, ...props }: HTMLAttributes<HTMLDivElement>) => (
    <div {...props}>{children}</div>
  ),
}));
jest.mock("@/components/ui/stories-filter-state", () => ({
  useStoriesFilters: () => ({
    filters: {},
    setFilters: jest.fn(),
    resetFilters: jest.fn(),
  }),
}));
jest.mock("@/components/ui/stories-filter-bar", () => ({
  StoriesFilterBar: () => null,
}));
jest.mock("./header", () => ({
  Header: ({
    layout,
    setLayout,
  }: {
    layout: StoriesLayout;
    setLayout: (value: StoriesLayout) => void;
  }) => (
    <button
      onClick={() => {
        setLayout(layout === "list" ? "kanban" : "list");
      }}
      type="button"
    >
      Switch layout
    </button>
  ),
}));
jest.mock("./all-stories", () => ({
  AllStories: ({ layout }: { layout: StoriesLayout }) => {
    const { useTeamOptions } =
      jest.requireActual<typeof TeamOptionsProviderModule>("./provider");
    const { viewOptions } = useTeamOptions();
    return (
      <output aria-label="Story preferences">
        {JSON.stringify({ layout, viewOptions })}
      </output>
    );
  },
}));

const BOARD_OPTIONS: StoriesViewOptions = {
  groupBy: "priority",
  orderBy: "deadline",
  orderDirection: "asc",
  showEmptyGroups: false,
  showSubStories: true,
  displayColumns: ["Status", "Priority"],
  selectedCustomFieldIds: ["deal-amount"],
};
const LIST_OPTIONS: StoriesViewOptions = {
  ...BOARD_OPTIONS,
  groupBy: "assignee",
  selectedCustomFieldIds: ["campaign"],
};

describe("team stories preference hydration", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("hydrates the server's list before restoring Kanban and its custom-field preferences", async () => {
    localStorage.setItem("teams:stories:layout", JSON.stringify("kanban"));
    localStorage.setItem(
      "teams:stories:view-options:kanban",
      JSON.stringify(BOARD_OPTIONS),
    );
    localStorage.setItem(
      "teams:stories:view-options:list",
      JSON.stringify(LIST_OPTIONS),
    );

    // Server rendering cannot read the browser's persisted preferences.
    const serverStorage = jest
      .spyOn(Storage.prototype, "getItem")
      .mockReturnValue(null);
    let serverHtml: string;
    try {
      serverHtml = renderToString(<ListStories />);
      expect(serverStorage).not.toHaveBeenCalled();
    } finally {
      serverStorage.mockRestore();
    }

    const container = document.createElement("div");
    container.innerHTML = serverHtml;
    document.body.appendChild(container);
    const readPreferences = () =>
      JSON.parse(
        within(container).getByRole("status", {
          name: "Story preferences",
        }).textContent,
      ) as { layout: StoriesLayout; viewOptions: StoriesViewOptions };
    expect(readPreferences()).toMatchObject({
      layout: "list",
      viewOptions: { groupBy: "status" },
    });
    expect(
      readPreferences().viewOptions.selectedCustomFieldIds,
    ).toBeUndefined();

    const onRecoverableError = jest.fn();
    const consoleError = jest
      .spyOn(console, "error")
      .mockImplementation(() => {});
    let root: Root | undefined;
    try {
      await act(async () => {
        root = hydrateRoot(container, <ListStories />, { onRecoverableError });
      });
      expect(onRecoverableError).not.toHaveBeenCalled();
      expect(consoleError).not.toHaveBeenCalled();
      expect(readPreferences()).toEqual({
        layout: "kanban",
        viewOptions: BOARD_OPTIONS,
      });

      // After hydration, layout changes synchronously restore their own scope.
      fireEvent.click(
        within(container).getByRole("button", { name: "Switch layout" }),
      );
      expect(readPreferences()).toEqual({
        layout: "list",
        viewOptions: LIST_OPTIONS,
      });
      fireEvent.click(
        within(container).getByRole("button", { name: "Switch layout" }),
      );
      expect(readPreferences()).toEqual({
        layout: "kanban",
        viewOptions: BOARD_OPTIONS,
      });
      expect(
        JSON.parse(localStorage.getItem("teams:stories:view-options:list")!),
      ).toEqual(LIST_OPTIONS);
      expect(
        JSON.parse(localStorage.getItem("teams:stories:view-options:kanban")!),
      ).toEqual(BOARD_OPTIONS);
    } finally {
      await act(async () => {
        root?.unmount();
      });
      container.remove();
      consoleError.mockRestore();
    }
  });
});
