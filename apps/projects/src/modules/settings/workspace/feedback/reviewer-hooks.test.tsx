/* global beforeEach, describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { act, renderHook, waitFor } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { updateFeedbackBoardReviewer } from "./actions";
import { getFeedbackBoardReviewers } from "./queries";
import {
  useFeedbackBoardReviewers,
  useUpdateFeedbackBoardReviewerMutation,
} from "./hooks";
import type { FeedbackReviewer } from "./types";

jest.mock("sonner", () => ({
  toast: {
    error: jest.fn(),
    success: jest.fn(),
  },
}));

jest.mock("@/lib/auth/client", () => ({
  useSession: jest.fn(),
}));

jest.mock("@/hooks", () => ({
  useWorkspacePath: jest.fn(),
}));

jest.mock("./actions", () => ({
  createFeedbackBoard: jest.fn(),
  updateFeedbackBoardReviewer: jest.fn(),
  updateFeedbackPortal: jest.fn(),
}));

jest.mock("./queries", () => ({
  getFeedbackBoardReviewers: jest.fn(),
  getFeedbackPortals: jest.fn(),
}));

const mockUseSession = jest.mocked(useSession);
const mockUseWorkspacePath = jest.mocked(useWorkspacePath);
const mockGetFeedbackBoardReviewers = jest.mocked(getFeedbackBoardReviewers);
const mockUpdateFeedbackBoardReviewer = jest.mocked(
  updateFeedbackBoardReviewer,
);

const reviewer: FeedbackReviewer = {
  avatarUrl: "https://cdn.example.com/tariro.jpg",
  email: "tariro@example.com",
  emailFrequency: "off",
  name: "Tariro Ncube",
  role: "member",
  userId: "user-2",
};

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      mutations: { retry: false },
      queries: { retry: false },
    },
  });

const wrapperFor = (queryClient: QueryClient) =>
  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
    );
  };

describe("feedback reviewer hooks", () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockUseSession.mockReturnValue({
      data: { token: "token" },
    } as unknown as ReturnType<typeof useSession>);
    mockUseWorkspacePath.mockReturnValue({
      workspaceSlug: "city-roads",
    } as ReturnType<typeof useWorkspacePath>);
  });

  it("does not fetch reviewers until its dialog is open", async () => {
    const queryClient = createQueryClient();
    mockGetFeedbackBoardReviewers.mockResolvedValue([reviewer]);

    const { result, rerender } = renderHook(
      ({ enabled }) => useFeedbackBoardReviewers("board-1", enabled),
      {
        initialProps: { enabled: false },
        wrapper: wrapperFor(queryClient),
      },
    );

    expect(mockGetFeedbackBoardReviewers).not.toHaveBeenCalled();

    rerender({ enabled: true });

    await waitFor(() => {
      expect(result.current.data).toEqual([reviewer]);
    });
    expect(mockGetFeedbackBoardReviewers).toHaveBeenCalledTimes(1);
  });

  it("rolls an optimistic frequency change back when the API rejects it", async () => {
    const queryClient = createQueryClient();
    const reviewersKey = feedbackKeys.reviewers("city-roads", "board-1");
    queryClient.setQueryData(reviewersKey, [reviewer]);

    type ReviewerResponse = Awaited<
      ReturnType<typeof updateFeedbackBoardReviewer>
    >;
    let resolveUpdate: (response: ReviewerResponse) => void = () => undefined;
    mockUpdateFeedbackBoardReviewer.mockImplementation(
      () =>
        new Promise<ReviewerResponse>((resolve) => {
          resolveUpdate = resolve;
        }),
    );

    const { result } = renderHook(
      () => useUpdateFeedbackBoardReviewerMutation("board-1"),
      { wrapper: wrapperFor(queryClient) },
    );

    act(() => {
      result.current.mutate({
        input: { emailFrequency: "weekly" },
        userId: reviewer.userId,
      });
    });

    await waitFor(() => {
      expect(
        queryClient.getQueryData<FeedbackReviewer[]>(reviewersKey)?.[0]
          ?.emailFrequency,
      ).toBe("weekly");
    });

    act(() => {
      resolveUpdate({ error: { message: "Could not update reviewer" } });
    });

    await waitFor(() => {
      expect(
        queryClient.getQueryData<FeedbackReviewer[]>(reviewersKey)?.[0]
          ?.emailFrequency,
      ).toBe("off");
    });
    expect(toast.error).toHaveBeenCalledWith("Failed to update reviewer", {
      description: "Could not update reviewer",
    });
  });
});
