import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import { SavedViews } from "./saved-views";
import { requestCreatedViewFocus, selectCurrentView } from "./view-selection";
import type { SavedView } from "./resolve-views";

let mockViewId: string | null = "00000000-0000-4000-8000-000000000001";
let mockUserId = "owner";
let mockWorkspace = "acme";
let mockRole = "member";
const mockSetViewId = jest.fn();
const mockSetCreateView = jest.fn();
const mockRename = jest.fn();
const mockArchive = jest.fn();
const mockSelected = jest.fn();
const mockCreate = jest.fn();
const mockRefetch = jest.fn();
jest.mock("next/navigation", () => ({
  useRouter: () => ({ push: jest.fn() }),
}));
jest.mock("nuqs", () => ({
  parseAsString: {},
  useQueryState: (key: string) =>
    key === "view" ? [mockViewId, mockSetViewId] : [null, mockSetCreateView],
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUserId } } }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({
    workspaceSlug: mockWorkspace,
    withWorkspace: (path: string) => `/${mockWorkspace}${path}`,
  }),
}));
jest.mock("./use-selected-view", () => ({
  useSelectedView: (...args: unknown[]) => mockSelected(...args),
}));
jest.mock("./view-favorite-button", () => ({
  ViewFavoriteButton: ({ name }: { name: string }) => (
    <button type="button">Favorite {name}</button>
  ),
}));
jest.mock("./hooks", () => ({
  usePresetMutations: () => ({
    create: { mutateAsync: mockCreate },
    rename: { mutateAsync: mockRename },
    archive: { mutateAsync: mockArchive },
  }),
  useWorkPresets: () => ({
    data: { pages: [{ items: [VIEW] }] },
    isPending: false,
    isError: false,
  }),
}));
jest.mock("sonner", () => ({
  toast: { error: jest.fn(), success: jest.fn() },
}));

const VIEW: SavedView = {
  id: "00000000-0000-4000-8000-000000000001",
  teamId: "team",
  ownerId: "owner",
  kind: "view",
  visibility: "team",
  name: "My pipeline",
  canEdit: false,
  createdAt: "2026-10-01",
  updatedAt: "2026-10-01",
  configuration: {
    version: 1,
    layout: "kanban",
    filters: { ...DEFAULT_STORIES_FILTER, assignedToMe: true },
    viewOptions: {
      groupBy: "priority",
      orderBy: "deadline",
      orderDirection: "asc",
      displayColumns: ["Status"],
      showSubStories: true,
      showEmptyGroups: false,
    },
  },
};
const loaded = (extra = {}) => ({
  data: VIEW,
  isError: false,
  isFetching: false,
  isFetchedAfterMount: true,
  refetch: mockRefetch,
  ...extra,
});

describe("selected saved views", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockViewId = VIEW.id;
    mockUserId = "owner";
    mockWorkspace = "acme";
    mockRole = "member";
    mockSelected.mockReturnValue(loaded());
    mockCreate.mockResolvedValue(VIEW);
    mockRename.mockResolvedValue(VIEW);
  });
  afterEach(() => jest.restoreAllMocks());

  it("does not add a Views toolbar control or apply cached data to an ordinary task list", () => {
    mockViewId = null;
    const onApply = jest.fn();
    const onLoadStateChange = jest.fn();
    const { container, rerender } = render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(container).toBeEmptyDOMElement();
    expect(onApply).not.toHaveBeenCalled();
    expect(onLoadStateChange).toHaveBeenLastCalledWith({
      workspaceSlug: "acme",
      userId: "owner",
      teamId: "team",
      viewId: null,
      status: "idle",
    });

    mockSelected.mockReturnValue(
      loaded({ data: undefined, isFetchedAfterMount: false, isFetching: true }),
    );
    rerender(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    expect(container).toBeEmptyDOMElement();
    expect(onApply).not.toHaveBeenCalled();
  });

  it.each(["workspace", "account", "view"])(
    "does not restore deferred created-view focus into a changed %s",
    (change) => {
      const deferredFrames: FrameRequestCallback[] = [];
      jest
        .spyOn(window, "requestAnimationFrame")
        .mockImplementation((callback) => {
          deferredFrames.push(callback);
          return deferredFrames.length;
        });
      requestCreatedViewFocus({
        workspaceSlug: "acme",
        userId: "owner",
        teamId: "team",
        id: VIEW.id,
      });
      const props = {
        configuration: VIEW.configuration,
        onApply: jest.fn(),
        teamId: "team",
      };
      const { rerender } = render(<SavedViews {...props} />);
      expect(deferredFrames).toHaveLength(1);
      if (change === "workspace") mockWorkspace = "other-workspace";
      if (change === "account") mockUserId = "other-user";
      if (change === "view") {
        mockViewId = "00000000-0000-4000-8000-000000000002";
        mockSelected.mockReturnValue(
          loaded({
            data: { ...VIEW, id: mockViewId, name: "Another pipeline" },
          }),
        );
      }
      rerender(<SavedViews {...props} />);
      act(() => {
        deferredFrames.forEach((callback) => {
          callback(0);
        });
      });
      expect(document.body).toHaveFocus();
      expect(
        screen.getByRole("button", {
          name:
            change === "view"
              ? "Manage Another pipeline"
              : "Manage My pipeline",
        }),
      ).not.toHaveFocus();
    },
  );

  it("waits for a fresh direct-link lookup before applying all saved options", () => {
    const onApply = jest.fn();
    const onLoadStateChange = jest.fn();
    mockSelected.mockReturnValue(
      loaded({ isFetchedAfterMount: false, isFetching: true }),
    );
    const { rerender } = render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(onApply).not.toHaveBeenCalled();
    expect(onLoadStateChange).toHaveBeenLastCalledWith({
      workspaceSlug: "acme",
      userId: "owner",
      teamId: "team",
      viewId: VIEW.id,
      status: "loading",
    });
    mockSelected.mockReturnValue(loaded());
    rerender(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(onApply).toHaveBeenCalledTimes(1);
    expect(onApply).toHaveBeenCalledWith(VIEW.configuration);
    expect(
      screen.getByRole("button", { name: "Manage My pipeline" }),
    ).toBeInTheDocument();
  });

  it("preserves in-session adjustments while background requests finish", () => {
    const onApply = jest.fn();
    const onLoadStateChange = jest.fn();
    const { rerender } = render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    mockSelected.mockReturnValue(loaded({ isFetching: true }));
    rerender(
      <SavedViews
        configuration={{ ...VIEW.configuration, layout: "list" }}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(onLoadStateChange).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: "ready" }),
    );
    mockSelected.mockReturnValue(
      loaded({
        data: {
          ...VIEW,
          name: "Renamed pipeline",
          configuration: { ...VIEW.configuration, layout: "list" },
        },
      }),
    );
    rerender(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        onLoadStateChange={onLoadStateChange}
        teamId="team"
      />,
    );
    expect(onApply).toHaveBeenCalledTimes(1);
    expect(
      screen.getByRole("button", { name: "Manage Renamed pipeline" }),
    ).toBeInTheDocument();
  });

  it.each([
    { data: null, isError: false, status: "unavailable" },
    { data: VIEW, isError: true, status: "error" },
  ])(
    "does not apply stale or unavailable data ($status)",
    ({ data, isError, status }) => {
      mockSelected.mockReturnValue(loaded({ data, isError }));
      const onApply = jest.fn();
      const onLoadStateChange = jest.fn();
      render(
        <SavedViews
          configuration={VIEW.configuration}
          onApply={onApply}
          onLoadStateChange={onLoadStateChange}
          teamId="team"
        />,
      );
      expect(onApply).not.toHaveBeenCalled();
      expect(onLoadStateChange).toHaveBeenLastCalledWith(
        expect.objectContaining({ status }),
      );
      fireEvent.click(screen.getByRole("button", { name: "Try again" }));
      expect(mockRefetch).toHaveBeenCalledTimes(1);
    },
  );

  it("retains edited filters and layout when the header replaces its application callback", () => {
    const initialApply = jest.fn();
    const editedApply = jest.fn();
    const { rerender } = render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={initialApply}
        teamId="team"
      />,
    );
    const edited = {
      ...VIEW.configuration,
      layout: "list" as const,
      filters: { ...VIEW.configuration.filters, statusIds: ["in-progress"] },
    };
    mockSelected.mockReturnValue(loaded({ data: { ...VIEW } }));
    rerender(
      <SavedViews configuration={edited} onApply={editedApply} teamId="team" />,
    );
    expect(initialApply).toHaveBeenCalledTimes(1);
    expect(editedApply).not.toHaveBeenCalled();
  });

  it("keeps the view switcher and keyboard focus mounted while a different selection loads", () => {
    const props = {
      configuration: VIEW.configuration,
      onApply: jest.fn(),
      teamId: "team",
    };
    const { rerender } = render(<SavedViews {...props} />);
    const switcher = screen.getByRole("button", {
      name: "Switch view: My pipeline",
    });
    switcher.focus();
    mockViewId = "00000000-0000-4000-8000-000000000002";
    mockSelected.mockReturnValue(
      loaded({ data: undefined, isFetchedAfterMount: false, isFetching: true }),
    );
    rerender(<SavedViews {...props} />);
    expect(screen.getByRole("button", { name: "Switch view" })).toBe(switcher);
    expect(switcher).toHaveFocus();
    mockSelected.mockReturnValue(
      loaded({ data: { ...VIEW, id: mockViewId, name: "Release queue" } }),
    );
    rerender(<SavedViews {...props} />);
    expect(
      screen.getByRole("button", { name: "Switch view: Release queue" }),
    ).toBe(switcher);
    expect(switcher).toHaveFocus();
  });

  it("resolves a deliberate current-view selection again, without applying the menu snapshot", async () => {
    const onApply = jest.fn();
    mockSelected.mockImplementation(
      (_team: string, _view: string, selection: number) =>
        selection
          ? loaded({
              data: undefined,
              isFetchedAfterMount: false,
              isFetching: true,
            })
          : loaded(),
    );
    const { rerender } = render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        teamId="team"
      />,
    );
    const manage = screen.getByRole("button", { name: "Manage My pipeline" });
    act(() => {
      manage.focus();
    });
    fireEvent.keyDown(manage, { key: "Enter" });
    const switchView = await screen.findByRole("menuitem", {
      name: "Switch view",
    });
    act(() => {
      switchView.focus();
    });
    fireEvent.keyDown(switchView, { key: "ArrowRight" });
    const currentView = await screen.findByRole("menuitem", {
      name: "My pipeline",
    });
    act(() => {
      currentView.focus();
    });
    fireEvent.keyDown(currentView, { key: "Enter" });
    expect(mockSelected).toHaveBeenLastCalledWith("team", VIEW.id, 1);
    expect(onApply).toHaveBeenCalledTimes(1);
    mockSelected.mockReturnValue(loaded());
    rerender(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        teamId="team"
      />,
    );
    expect(onApply).toHaveBeenCalledTimes(2);
  });

  it("only handles active-favorite requests for the current account and view", () => {
    const onApply = jest.fn();
    mockSelected.mockImplementation(
      (_team: string, _view: string, selection: number) =>
        selection
          ? loaded({ isFetchedAfterMount: false, isFetching: true })
          : loaded(),
    );
    render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={onApply}
        teamId="team"
      />,
    );
    act(() => {
      selectCurrentView({
        id: VIEW.id,
        teamId: "team",
        workspaceSlug: "acme",
        userId: "other",
      });
    });
    expect(mockSelected).toHaveBeenLastCalledWith("team", VIEW.id, 0);
    act(() => {
      selectCurrentView({
        id: VIEW.id,
        teamId: "team",
        workspaceSlug: "acme",
        userId: "owner",
      });
    });
    expect(mockSelected).toHaveBeenLastCalledWith("team", VIEW.id, 1);
    expect(onApply).toHaveBeenCalledTimes(1);
  });

  it.each(["member", "guest"])(
    "offers navigation without alternate creation actions for %s",
    async (role) => {
      mockRole = role;
      render(
        <SavedViews
          configuration={VIEW.configuration}
          onApply={jest.fn()}
          teamId="team"
        />,
      );
      const manage = screen.getByRole("button", { name: "Manage My pipeline" });
      act(() => {
        manage.focus();
      });
      fireEvent.keyDown(manage, { key: "Enter" });
      await screen.findByRole("menuitem", { name: "Browse all views" });
      expect(
        screen.queryByRole("menuitem", { name: "Save current view" }),
      ).not.toBeInTheDocument();
      expect(
        screen.queryByRole("menuitem", { name: "Create view" }),
      ).not.toBeInTheDocument();
      expect(
        screen.getByRole("menuitem", { name: "Browse all views" }),
      ).toHaveAttribute("href", "/acme/views");
    },
  );
  it("focuses the saved-header Rename input after menu teardown and returns to Manage", async () => {
    mockSelected.mockReturnValue(loaded({ data: { ...VIEW, canEdit: true } }));
    render(
      <SavedViews
        configuration={VIEW.configuration}
        onApply={jest.fn()}
        teamId="team"
      />,
    );
    const manage = screen.getByRole("button", { name: "Manage My pipeline" });
    act(() => {
      manage.focus();
    });
    fireEvent.keyDown(manage, { key: "Enter" });
    const rename = await screen.findByRole("menuitem", { name: "Rename view" });
    act(() => {
      rename.focus();
    });
    fireEvent.keyDown(rename, { key: "Enter" });
    const dialog = await screen.findByRole("dialog", { name: "Rename view" });
    await waitFor(() => {
      expect(within(dialog).getByLabelText("Name")).toHaveFocus();
    });
    fireEvent.click(within(dialog).getByRole("button", { name: "Cancel" }));
    await waitFor(() => {
      expect(manage).toHaveFocus();
    });
  });

  it("restores the replacement Views trigger after archiving the selected view", async () => {
    mockSelected.mockReturnValue(loaded({ data: { ...VIEW, canEdit: true } }));
    const props = {
      configuration: VIEW.configuration,
      onApply: jest.fn(),
      teamId: "team",
    };
    const { rerender } = render(<SavedViews {...props} />);
    mockArchive.mockImplementation(async () => {
      mockSelected.mockReturnValue(loaded({ data: null }));
      rerender(<SavedViews {...props} />);
    });
    const manage = screen.getByRole("button", { name: "Manage My pipeline" });
    act(() => {
      manage.focus();
    });
    fireEvent.keyDown(manage, { key: "Enter" });
    const archive = await screen.findByRole("menuitem", {
      name: "Archive view",
    });
    act(() => {
      archive.focus();
    });
    fireEvent.keyDown(archive, { key: "Enter" });
    await waitFor(() => {
      expect(screen.getByRole("button", { name: "Switch view" })).toHaveFocus();
    });
  });
});
