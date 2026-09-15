/* global afterEach, beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { useAnalytics, useWorkspacePath } from "@/hooks";
import {
  calendarKeys,
  developerKeys,
  feedbackKeys,
  labelKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
  userKeys,
} from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { objectiveKeys } from "@/shared/objectives/keys";
import { documentKeys } from "@/shared/documents/keys";
import { deleteTeamAction } from "../actions/delete-team";
import { useDeleteTeamMutation } from "./delete-team-mutation";

jest.mock("@/hooks", () => ({
  useAnalytics: jest.fn(),
  useWorkspacePath: jest.fn(),
}));

jest.mock("sonner", () => ({
  toast: {
    loading: jest.fn(),
    dismiss: jest.fn(),
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("../actions/delete-team", () => ({
  deleteTeamAction: jest.fn(),
}));

const WORKSPACE_SLUG = "test-workspace";
const TEAM_ID = "deleted-team";
const track = jest.fn();
const mockDeleteTeamAction = jest.mocked(deleteTeamAction);

describe("useDeleteTeamMutation", () => {
  let queryClient: QueryClient;

  const wrapper = ({ children }: { children: ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  );

  beforeEach(() => {
    jest.clearAllMocks();
    queryClient = new QueryClient({
      defaultOptions: {
        queries: { retry: false },
        mutations: { retry: false },
      },
    });
    jest.mocked(useWorkspacePath).mockReturnValue({
      workspaceSlug: WORKSPACE_SLUG,
    } as ReturnType<typeof useWorkspacePath>);
    jest.mocked(useAnalytics).mockReturnValue({
      analytics: { track },
    } as unknown as ReturnType<typeof useAnalytics>);
  });

  afterEach(() => {
    queryClient.clear();
  });

  it("treats an API error response as failure and preserves cached data", async () => {
    const teamsKey = teamKeys.lists(WORKSPACE_SLUG);
    const teams = [{ id: TEAM_ID }];
    queryClient.setQueryData(teamsKey, teams);
    mockDeleteTeamAction.mockResolvedValueOnce({
      error: { message: "The team could not be deleted." },
    });
    const onSuccess = jest.fn();
    const { result } = renderHook(useDeleteTeamMutation, { wrapper });

    act(() => {
      result.current.mutate(TEAM_ID, { onSuccess });
    });
    await waitFor(() => {
      expect(result.current.isError).toBe(true);
    });

    expect(result.current.error?.message).toBe(
      "The team could not be deleted.",
    );
    expect(onSuccess).not.toHaveBeenCalled();
    expect(track).not.toHaveBeenCalled();
    expect(toast.success).not.toHaveBeenCalled();
    expect(toast.error).toHaveBeenCalledWith(
      "Failed to delete team",
      expect.objectContaining({
        description: "The team could not be deleted.",
      }),
    );
    expect(queryClient.getQueryData(teamsKey)).toEqual(teams);
    expect(queryClient.getQueryState(teamsKey)?.isInvalidated).toBe(false);
  });

  it("removes the deleted team's caches and refreshes dependencies only in this workspace", async () => {
    const deletedKeys = [
      teamKeys.detail(WORKSPACE_SLUG, TEAM_ID),
      teamKeys.settings(WORKSPACE_SLUG, TEAM_ID),
      storyKeys.team(WORKSPACE_SLUG, TEAM_ID),
      objectiveKeys.team(WORKSPACE_SLUG, TEAM_ID),
      statusKeys.team(WORKSPACE_SLUG, TEAM_ID),
      labelKeys.team(WORKSPACE_SLUG, TEAM_ID),
      sprintKeys.team(WORKSPACE_SLUG, TEAM_ID),
    ];
    const refreshedKeys = [
      teamKeys.lists(WORKSPACE_SLUG),
      teamKeys.public(WORKSPACE_SLUG),
      teamKeys.joined(WORKSPACE_SLUG),
      storyKeys.mine(WORKSPACE_SLUG),
      storyKeys.total(WORKSPACE_SLUG),
      objectiveKeys.list(WORKSPACE_SLUG),
      labelKeys.lists(WORKSPACE_SLUG),
      statusKeys.lists(WORKSPACE_SLUG),
      sprintKeys.running(WORKSPACE_SLUG),
      feedbackKeys.teamSummaries(WORKSPACE_SLUG),
      documentKeys.lists(WORKSPACE_SLUG),
      calendarKeys.schedules(WORKSPACE_SLUG),
      developerKeys.personalTokens(WORKSPACE_SLUG),
      developerKeys.serviceAccountKeys(WORKSPACE_SLUG, "service-account"),
    ];
    const unchangedKeys = [
      teamKeys.detail(WORKSPACE_SLUG, "other-team"),
      teamKeys.settings(WORKSPACE_SLUG, "other-team"),
      teamKeys.lists("other-workspace"),
      storyKeys.mine("other-workspace"),
      objectiveKeys.list("other-workspace"),
      userKeys.profile(),
      developerKeys.oauthApplications(WORKSPACE_SLUG),
      developerKeys.webhooks(WORKSPACE_SLUG),
      developerKeys.personalTokens("other-workspace"),
    ];
    for (const key of [...deletedKeys, ...refreshedKeys, ...unchangedKeys]) {
      queryClient.setQueryData(key, [{ id: "cached" }]);
    }
    mockDeleteTeamAction.mockResolvedValueOnce({});
    const onSuccess = jest.fn();
    const { result } = renderHook(useDeleteTeamMutation, { wrapper });

    act(() => {
      result.current.mutate(TEAM_ID, { onSuccess });
    });
    await waitFor(() => {
      expect(result.current.isSuccess).toBe(true);
    });

    expect(mockDeleteTeamAction).toHaveBeenCalledWith(TEAM_ID, WORKSPACE_SLUG);
    expect(onSuccess).toHaveBeenCalledTimes(1);
    expect(track).toHaveBeenCalledWith("team_deleted", { teamId: TEAM_ID });
    for (const key of deletedKeys) {
      expect(queryClient.getQueryState(key)).toBeUndefined();
    }
    for (const key of refreshedKeys) {
      expect(queryClient.getQueryState(key)?.isInvalidated).toBe(true);
    }
    for (const key of unchangedKeys) {
      expect(queryClient.getQueryData(key)).toEqual([{ id: "cached" }]);
      expect(queryClient.getQueryState(key)?.isInvalidated).toBe(false);
    }
  });
});
