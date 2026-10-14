/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { act, renderHook, waitFor } from "@testing-library/react";
import { sprintKeys } from "@/constants/keys";
import { createSprintAction } from "../actions/create-sprint";
import { updateSprintAction } from "../actions/update-sprint";
import { deleteSprintAction } from "../actions/delete-sprint";
import { getRunningSprints } from "../queries/get-running-sprints";
import type { Sprint } from "../types";
import { useCreateSprintMutation } from "./create-sprint-mutation";
import { useUpdateSprintMutation } from "./update-sprint-mutation";
import { useDeleteSprintMutation } from "./delete-sprint-mutation";
import { useRunningSprints } from "./running-sprints";

jest.mock("@/hooks", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "workspace-a" }),
  useAnalytics: () => ({ analytics: { track: jest.fn() } }),
}));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: { token: "token" } }),
}));
jest.mock("sonner", () => ({
  toast: { success: jest.fn(), error: jest.fn() },
}));
jest.mock("../actions/create-sprint", () => ({
  createSprintAction: jest.fn(),
}));
jest.mock("../actions/update-sprint", () => ({
  updateSprintAction: jest.fn(),
}));
jest.mock("../actions/delete-sprint", () => ({
  deleteSprintAction: jest.fn(),
}));
jest.mock("../queries/get-running-sprints", () => ({
  getRunningSprints: jest.fn(),
}));

const sprint: Sprint = {
  id: "sprint-a",
  name: "Current sprint",
  goal: "Deliver the next milestone",
  objectiveId: "objective-a",
  teamId: "team-a",
  workspaceId: "workspace-a",
  startDate: "2026-10-14",
  endDate: "2026-10-28",
  createdAt: "2026-10-14T00:00:00Z",
  updatedAt: "2026-10-14T00:00:00Z",
  scheduleManagedByAutomation: false,
  stats: {
    total: 0,
    cancelled: 0,
    completed: 0,
    started: 0,
    unstarted: 0,
    backlog: 0,
  },
};

type MutationKind = "create" | "update" | "delete";

const useSprintMutations = () => {
  const create = useCreateSprintMutation();
  const update = useUpdateSprintMutation();
  const remove = useDeleteSprintMutation();
  const running = useRunningSprints();

  const mutate = (kind: MutationKind) => {
    if (kind === "create") return create.mutateAsync(sprint);
    if (kind === "update") {
      return update.mutateAsync({
        sprintId: sprint.id,
        updates: { endDate: "2026-10-29" },
      });
    }
    return remove.mutateAsync(sprint.id);
  };

  return { mutate, running };
};

const setup = () => {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  queryClient.setQueryData(sprintKeys.running("workspace-a"), [sprint]);
  queryClient.setQueryData(sprintKeys.running("workspace-b"), []);
  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );
  return { queryClient, ...renderHook(useSprintMutations, { wrapper }) };
};

describe("running sprint freshness after mutations", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    jest.mocked(createSprintAction).mockResolvedValue({ data: sprint });
    jest.mocked(updateSprintAction).mockResolvedValue({ data: null });
    jest.mocked(deleteSprintAction).mockResolvedValue({ data: null });
    jest.mocked(getRunningSprints).mockResolvedValue([]);
  });

  it.each<MutationKind>(["create", "update", "delete"])(
    "refetches a fresh running-sprint query after %s without invalidating another workspace",
    async (kind) => {
      const { result, queryClient, unmount } = setup();
      expect(result.current.running.data).toEqual([sprint]);
      expect(getRunningSprints).not.toHaveBeenCalled();

      await act(async () => {
        await result.current.mutate(kind);
      });

      await waitFor(() => {
        expect(result.current.running.data).toEqual([]);
      });
      expect(getRunningSprints).toHaveBeenCalledTimes(1);
      expect(getRunningSprints).toHaveBeenCalledWith({
        session: { token: "token" },
        workspaceSlug: "workspace-a",
      });
      expect(
        queryClient.getQueryState(sprintKeys.running("workspace-b"))
          ?.isInvalidated,
      ).toBe(false);
      unmount();
      queryClient.clear();
    },
  );

  it.each<MutationKind>(["create", "update", "delete"])(
    "keeps the running list intact when %s is rejected",
    async (kind) => {
      const error = new Error("Sprint change rejected");
      jest.mocked(createSprintAction).mockRejectedValue(error);
      jest.mocked(updateSprintAction).mockRejectedValue(error);
      jest.mocked(deleteSprintAction).mockRejectedValue(error);
      const { result, queryClient, unmount } = setup();

      await act(async () => {
        await expect(result.current.mutate(kind)).rejects.toThrow(error);
      });

      expect(getRunningSprints).not.toHaveBeenCalled();
      expect(result.current.running.data).toEqual([sprint]);
      expect(
        queryClient.getQueryState(sprintKeys.running("workspace-a"))
          ?.isInvalidated,
      ).toBe(false);
      unmount();
      queryClient.clear();
    },
  );
});
