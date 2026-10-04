import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { ViewRow } from "./view-row";
import type { SavedView } from "./resolve-views";

const mockRename = jest.fn();
const mockArchive = jest.fn();
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ withWorkspace: (path: string) => `/acme${path}` }),
}));
jest.mock("./hooks", () => ({
  usePresetMutations: () => ({
    rename: { mutateAsync: mockRename },
    archive: { mutateAsync: mockArchive },
  }),
}));
jest.mock("./view-favorite-button", () => ({ ViewFavoriteButton: () => null }));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));
const VIEW: SavedView = {
  id: "view",
  teamId: "team",
  ownerId: "owner",
  kind: "view",
  visibility: "team",
  name: "Pipeline",
  canEdit: true,
  createdAt: "2026-10-01",
  updatedAt: "2026-10-01",
  configuration: {
    version: 1,
    layout: "list",
    filters: DEFAULT_STORIES_FILTER,
    viewOptions: {
      groupBy: "status",
      orderBy: "created",
      orderDirection: "desc",
      displayColumns: [],
      showSubStories: false,
      showEmptyGroups: true,
    },
  },
};

describe("view management", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockRename.mockResolvedValue(VIEW);
  });
  it.each(["Cancel", "Escape", "Save"])(
    "restores the Manage button after %s",
    async (dismissal) => {
      render(<ViewRow detail="Product · You" view={VIEW} />);
      const trigger = screen.getByRole("button", { name: "Manage Pipeline" });
      trigger.focus();
      fireEvent.keyDown(trigger, { key: "Enter", code: "Enter" });
      const rename = await screen.findByRole("menuitem", { name: "Rename" });
      rename.focus();
      fireEvent.keyDown(rename, { key: "Enter", code: "Enter" });
      const dialog = await screen.findByRole("dialog", { name: "Rename view" });
      await waitFor(() => {
        expect(within(dialog).getByLabelText("Name")).toHaveFocus();
      });
      if (dismissal === "Escape")
        fireEvent.keyDown(dialog, { key: "Escape", code: "Escape" });
      else
        fireEvent.click(
          within(dialog).getByRole("button", { name: dismissal }),
        );
      await waitFor(() => {
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
        expect(trigger).toHaveFocus();
      });
      if (dismissal === "Save")
        expect(mockRename).toHaveBeenCalledWith({
          id: "view",
          name: "Pipeline",
        });
    },
  );
  it("allows read-only views to open without exposing editing actions", () => {
    render(
      <ViewRow
        detail="Product · Teammate"
        view={{ ...VIEW, canEdit: false }}
      />,
    );
    expect(
      screen.getByRole("link", { name: "Pipeline Product · Teammate" }),
    ).toHaveAttribute("href", "/acme/teams/team/stories?view=view");
    expect(
      screen.queryByRole("button", { name: "Manage Pipeline" }),
    ).not.toBeInTheDocument();
  });
});
