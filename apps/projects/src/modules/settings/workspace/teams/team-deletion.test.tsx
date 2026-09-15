/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import type { Team } from "@/modules/teams/types";
import { deleteTeamAction } from "@/modules/teams/actions/delete-team";
import { WorkspaceTeam } from "./components/team";
import { DeleteTeam } from "./management/components/delete";

const mockReplace = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ replace: mockReplace }),
  usePathname: () => "/test-workspace/settings/workspace/teams",
}));

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useTerminology: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("ui", () => ({
  ...jest.requireActual("ui"),
  TimeAgo: () => null,
}));

jest.mock("@/components/ui", () => ({
  ...jest.requireActual("@/components/ui/confirm-dialog"),
  ...jest.requireActual("@/components/ui/row-wrapper"),
  ...jest.requireActual("@/components/ui/team-color"),
}));

jest.mock("sonner", () => ({
  toast: {
    loading: jest.fn(),
    dismiss: jest.fn(),
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("@/modules/teams/actions/delete-team", () => ({
  deleteTeamAction: jest.fn(),
}));

const team: Team = {
  id: "team-to-delete",
  name: "Engineering",
  color: "blue",
  code: "ENG",
  createdAt: "2026-01-01T00:00:00Z",
  updatedAt: "2026-01-01T00:00:00Z",
  memberCount: 1,
  isPrivate: false,
  workspaceId: "workspace-id",
  sprintsEnabled: true,
};

describe.each(["list", "settings"] as const)(
  "team deletion from %s",
  (entry) => {
    let queryClient: QueryClient;

    beforeEach(() => {
      jest.clearAllMocks();
      jest.mocked(deleteTeamAction).mockReset();
      queryClient = new QueryClient({
        defaultOptions: {
          queries: { retry: false },
          mutations: { retry: false },
        },
      });
      jest.mocked(useWorkspacePath).mockReturnValue({
        workspaceSlug: "test-workspace",
        withWorkspace: (path: string) => `/test-workspace${path}`,
      } as ReturnType<typeof useWorkspacePath>);
      jest.mocked(useAnalytics).mockReturnValue({
        analytics: { track: jest.fn() },
      } as unknown as ReturnType<typeof useAnalytics>);
      jest.mocked(useTerminology).mockReturnValue({
        getTermDisplay: () => "stories",
      } as unknown as ReturnType<typeof useTerminology>);
    });

    afterEach(() => {
      queryClient.clear();
    });

    it("keeps deletion open while pending and after failure, then closes only on successful retry", async () => {
      let resolveDelete!: (
        result: Awaited<ReturnType<typeof deleteTeamAction>>,
      ) => void;
      jest.mocked(deleteTeamAction).mockReturnValueOnce(
        new Promise((resolve) => {
          resolveDelete = resolve;
        }),
      );
      render(
        <QueryClientProvider client={queryClient}>
          {entry === "list" ? (
            <WorkspaceTeam {...team} />
          ) : (
            <DeleteTeam team={team} />
          )}
        </QueryClientProvider>,
      );
      if (entry === "list") {
        fireEvent.keyDown(
          screen.getByRole("button", { name: "More options" }),
          {
            key: "ArrowDown",
          },
        );
        fireEvent.click(
          await screen.findByRole("menuitem", { name: "Delete team..." }),
        );
      } else {
        fireEvent.click(screen.getByRole("button", { name: "Delete Team" }));
      }
      fireEvent.change(await screen.findByRole("textbox"), {
        target: {
          value: entry === "list" ? "  DELETE Team  " : "  I Understand  ",
        },
      });
      fireEvent.click(screen.getByRole("button", { name: "Delete team" }));
      const pendingButton = await screen.findByRole("button", {
        name: "Deleting team...",
      });
      expect(pendingButton).toBeDisabled();
      fireEvent.click(pendingButton);
      fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
      fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(deleteTeamAction).toHaveBeenCalledTimes(1);

      await act(async () => {
        resolveDelete({ error: { message: "Could not delete the team." } });
      });
      expect(await screen.findByRole("alert")).toHaveTextContent(
        "Could not delete the team.",
      );
      expect(screen.getByRole("dialog")).toBeInTheDocument();
      expect(mockReplace).not.toHaveBeenCalled();

      jest.mocked(deleteTeamAction).mockResolvedValueOnce({});
      fireEvent.click(screen.getByRole("button", { name: "Delete team" }));
      await waitFor(() => {
        expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
      });
      expect(deleteTeamAction).toHaveBeenCalledTimes(2);
      if (entry === "settings") {
        expect(mockReplace).toHaveBeenCalledWith(
          "/test-workspace/settings/workspace/teams",
        );
      } else {
        expect(mockReplace).not.toHaveBeenCalled();
      }
    });
  },
);
