import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { WorkspaceData } from "../types";
import { Sidebar } from "./sidebar";
const session = {
  user: { id: "u", fullName: "Test User", email: "u@example.com" },
  workspaces: [
    { id: "a", name: "Studio", slug: "studio" },
    { id: "b", name: "Engineering", slug: "engineering" },
  ],
  activeWorkspaceId: "a",
  needsSignIn: false,
};

describe("workspace switch feedback", () => {
  it("announces the target workspace and disables switching while navigation is pending", () => {
    render(
      <Sidebar
        session={session}
        data={null}
        page="my-work"
        onPage={vi.fn()}
        onWorkspace={vi.fn()}
        switchingId="b"
        online
        syncing={false}
      />,
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      "Switching to Engineering",
    );
    expect(screen.getByRole("combobox", { name: "Workspace" })).toBeDisabled();
  });
  it("sends the selected workspace identity", () => {
    const select = vi.fn();
    render(
      <Sidebar
        session={session}
        data={null}
        page="my-work"
        onPage={vi.fn()}
        onWorkspace={select}
        switchingId={null}
        online={false}
        syncing={false}
      />,
    );
    fireEvent.change(screen.getByRole("combobox", { name: "Workspace" }), {
      target: { value: "b" },
    });
    expect(select).toHaveBeenCalledWith("b");
    expect(screen.getByText("Working offline")).toBeInTheDocument();
  });
});

const emptyWorkspace: WorkspaceData = {
  workspaceId: "a",
  stories: [],
  documents: [],
  comments: [],
  teams: [],
  statuses: [],
  pending: [],
  lastSyncedAt: null,
};
describe("initial offline download feedback", () => {
  it("keeps preparation visible until a workspace snapshot is downloaded", () => {
    const props = {
      session,
      page: "settings" as const,
      onPage: vi.fn(),
      onWorkspace: vi.fn(),
      switchingId: null,
      online: true,
      syncing: false,
    };
    const { rerender } = render(<Sidebar {...props} data={emptyWorkspace} />);
    expect(
      screen.getByText("Preparing offline workspace…"),
    ).toBeInTheDocument();
    expect(
      screen.queryByText("Available on this device"),
    ).not.toBeInTheDocument();
    rerender(
      <Sidebar
        {...props}
        data={{
          ...emptyWorkspace,
          downloadState: { workspaceId: "a", stage: "settings", downloaded: 0 },
        }}
      />,
    );
    expect(screen.getByText("Downloading workspace…")).toBeInTheDocument();
    expect(screen.getByText("Downloading settings")).toBeInTheDocument();
    rerender(
      <Sidebar
        {...props}
        data={{
          ...emptyWorkspace,
          lastSyncedAt: "2026-10-14T12:00:00Z",
          downloadState: { workspaceId: "a", stage: "ready", downloaded: 8 },
        }}
      />,
    );
    expect(screen.getByText("Available on this device")).toBeInTheDocument();
  });
});
