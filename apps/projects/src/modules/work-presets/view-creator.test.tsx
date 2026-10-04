import type { Dispatch, SetStateAction } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { DEFAULT_STORIES_FILTER } from "@/components/ui/stories-filter-types";
import type { SavedViewConfiguration } from "./types";
import { ViewCreator } from "./view-creator";

type Route = {
  createView: string | null;
  view: string | null;
  visibility: string | null;
};
let mockInitialRoute: Route;
let mockNavigate: Dispatch<SetStateAction<Route>>;
let mockUser = "owner";
let mockWorkspace = "acme";
let mockRole: string | undefined = "member";
const mockSetRoute = jest.fn();
const mockCreate = jest.fn();
jest.mock("nuqs", () => ({
  parseAsString: {},
  useQueryStates: () => {
    const { useState } = jest.requireActual("react");
    const [route, setRoute] = useState(mockInitialRoute);
    mockNavigate = setRoute;
    return [
      route,
      async (patch: Partial<Route>) => {
        mockSetRoute(patch);
        setRoute((current: Route) => ({ ...current, ...patch }));
      },
    ];
  },
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { user: { id: mockUser } } }),
}));
jest.mock("@/hooks/role", () => ({
  useUserRole: () => ({ userRole: mockRole }),
}));
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: mockWorkspace }),
}));
jest.mock("./hooks", () => ({
  usePresetMutations: () => ({ create: { mutateAsync: mockCreate } }),
}));

const CONFIGURATION: SavedViewConfiguration = {
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
};
const Harness = ({
  teamId = "team",
  configuration = CONFIGURATION,
}: {
  teamId?: string;
  configuration?: SavedViewConfiguration;
}) => (
  <>
    <div data-view-focus-scope={`${mockWorkspace}:${mockUser}:${teamId}`}>
      <button data-view-menu-trigger type="button">
        Views
      </button>
    </div>
    <ViewCreator configuration={configuration} teamId={teamId} />
  </>
);

const scrollDescriptor = Object.getOwnPropertyDescriptor(
  HTMLElement.prototype,
  "scrollIntoView",
);
beforeAll(() =>
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: jest.fn(),
  }),
);
afterAll(() => {
  if (scrollDescriptor)
    Object.defineProperty(
      HTMLElement.prototype,
      "scrollIntoView",
      scrollDescriptor,
    );
  else Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
});

describe("Create view popup", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockInitialRoute = { createView: "true", view: null, visibility: null };
    mockUser = "owner";
    mockWorkspace = "acme";
    mockRole = "member";
    mockCreate.mockResolvedValue({ id: "saved-view" });
  });

  it("uses the current preview configuration and restores header focus immediately after creation", async () => {
    const { rerender } = render(<Harness />);
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "View name" })).toHaveFocus();
    });
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "  My pipeline  " },
    });
    const current = { ...CONFIGURATION, layout: "list" as const };
    rerender(<Harness configuration={current} />);
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({
        teamId: "team",
        kind: "view",
        name: "My pipeline",
        visibility: "personal",
        configuration: current,
      });
    });
    expect(mockSetRoute).toHaveBeenCalledWith({
      createView: null,
      visibility: null,
      view: "saved-view",
    });
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(screen.getByRole("button", { name: "Views" })).toHaveFocus();
    });
  });

  it.each(["Cancel", "Escape", "Close"])(
    "closes on %s and restores the persistent header trigger",
    async (dismissal) => {
      render(<Harness />);
      const dialog = screen.getByRole("dialog", { name: "Create view" });
      if (dismissal === "Escape") fireEvent.keyDown(dialog, { key: "Escape" });
      else fireEvent.click(screen.getByRole("button", { name: dismissal }));
      await waitFor(() => {
        expect(screen.getByRole("button", { name: "Views" })).toHaveFocus();
      });
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      expect(mockCreate).not.toHaveBeenCalled();
    },
  );

  it("first closes the icon picker on Escape without cancelling the draft", async () => {
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "My draft" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: "View icon: Automatic" }),
    );
    const search = screen.getByRole("combobox", { name: "Search view icons" });
    await waitFor(() => {
      expect(search).toHaveFocus();
    });
    fireEvent.keyDown(search, { key: "Escape" });
    await waitFor(() => {
      expect(
        screen.queryByRole("listbox", { name: "View icons" }),
      ).not.toBeInTheDocument();
    });
    expect(
      screen.getByRole("dialog", { name: "Create view" }),
    ).toBeInTheDocument();
    expect(screen.getByRole("textbox", { name: "View name" })).toHaveValue(
      "My draft",
    );
    expect(mockSetRoute).not.toHaveBeenCalled();
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
  });

  it("prevents dismissal, edits and duplicate creation while saving", async () => {
    let resolve!: (value: { id: string }) => void;
    mockCreate.mockImplementation(
      () =>
        new Promise((done) => {
          resolve = done;
        }),
    );
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "Pending view" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    expect(screen.getByRole("textbox", { name: "View name" })).toBeDisabled();
    expect(
      screen.getByRole("button", { name: "View icon: Automatic" }),
    ).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(
      screen.queryByRole("button", { name: "Close" }),
    ).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Creating..." }));
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
    fireEvent.pointerDown(document.body, { button: 0, pointerType: "mouse" });
    expect(mockSetRoute).not.toHaveBeenCalled();
    expect(mockCreate).toHaveBeenCalledTimes(1);
    await act(async () => {
      resolve({ id: "saved-view" });
    });
  });

  it("preserves the draft, error and name focus after failure for a deliberate retry", async () => {
    mockCreate.mockRejectedValueOnce(new Error("Connection lost"));
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "Retry view" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Connection lost",
    );
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "View name" })).toHaveFocus();
    });
    expect(screen.getByRole("textbox", { name: "View name" })).toHaveValue(
      "Retry view",
    );
    expect(mockSetRoute).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledTimes(2);
    });
  });

  it("discards cancelled draft state and initializes Team visibility on a later opening", async () => {
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "Old name" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await waitFor(() => {
      expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    });
    act(() => {
      mockNavigate({ createView: "true", view: null, visibility: "team" });
    });
    expect(screen.getByRole("textbox", { name: "View name" })).toHaveValue("");
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "Team view" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({ visibility: "team" }),
      );
    });
  });

  it.each(["navigate", "team", "account", "workspace"])(
    "does not route a completed save into a changed %s context",
    async (change) => {
      let resolve!: (value: { id: string }) => void;
      mockCreate.mockImplementation(
        () =>
          new Promise((done) => {
            resolve = done;
          }),
      );
      const { rerender } = render(<Harness />);
      fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
        target: { value: "Old context" },
      });
      fireEvent.click(screen.getByRole("button", { name: "Create view" }));
      if (change === "navigate")
        act(() => {
          mockNavigate({ createView: null, view: "other", visibility: null });
        });
      else {
        if (change === "account") mockUser = "other";
        if (change === "workspace") mockWorkspace = "other";
        rerender(<Harness teamId={change === "team" ? "other" : "team"} />);
        expect(screen.getByRole("textbox", { name: "View name" })).toHaveValue(
          "",
        );
      }
      await act(async () => {
        resolve({ id: "old-view" });
      });
      expect(mockSetRoute).not.toHaveBeenCalled();
    },
  );

  it.each(["guest", undefined])(
    "requires a known non-guest role before offering Create (%s)",
    (role) => {
      mockRole = role;
      render(<Harness />);
      expect(screen.getByRole("textbox", { name: "View name" })).toBeDisabled();
      expect(
        screen.getByRole("button", { name: "Create view" }),
      ).toBeDisabled();
    },
  );
  it("captures all current configuration unchanged while adding the chosen icon", async () => {
    const configuration: SavedViewConfiguration = {
      ...CONFIGURATION,
      description: "Existing description",
      viewOptions: {
        ...CONFIGURATION.viewOptions,
        displayColumns: ["Status", "Created", "Updated", "Epic"],
        displayColumnsVersion: 3,
        selectedCustomFieldIds: ["field"],
        hiddenKanbanGroups: { priority: ["Low"] },
      },
    };
    render(<Harness configuration={configuration} />);
    const icon = screen.getByRole("button", { name: "View icon: Automatic" });
    const name = screen.getByRole("textbox", { name: "View name" });
    expect(icon.compareDocumentPosition(name)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );
    expect(
      screen.queryByRole("textbox", { name: "View description" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Display" }),
    ).not.toBeInTheDocument();
    expect(
      screen.queryByRole("combobox", { name: "Visibility" }),
    ).not.toBeInTheDocument();
    fireEvent.change(name, { target: { value: "Incident triage" } });
    fireEvent.click(icon);
    fireEvent.click(screen.getByRole("option", { name: "Calendar" }));
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({
          configuration: { ...configuration, icon: "calendar" },
        }),
      );
    });
  });

  it("does not add optional metadata when no icon was chosen", async () => {
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "My view" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith(
        expect.objectContaining({ configuration: CONFIGURATION }),
      );
    });
    expect(mockCreate.mock.calls[0][0].configuration).not.toHaveProperty(
      "icon",
    );
    expect(mockCreate.mock.calls[0][0].configuration).not.toHaveProperty(
      "description",
    );
  });

  it("rejects an overlong name and preserves focus for correction", async () => {
    render(<Harness />);
    fireEvent.change(screen.getByRole("textbox", { name: "View name" }), {
      target: { value: "🗓".repeat(101) },
    });
    fireEvent.click(screen.getByRole("button", { name: "Create view" }));
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "1–100 characters",
    );
    expect(mockCreate).not.toHaveBeenCalled();
    await waitFor(() => {
      expect(screen.getByRole("textbox", { name: "View name" })).toHaveFocus();
    });
  });
});
