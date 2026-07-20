import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import {
  createFeedbackBoard,
  deleteFeedbackBoard,
  updateFeedbackBoardReviewer,
  updateFeedbackPortal,
} from "./actions";
import { getFeedbackBoardReviewers, getFeedbackPortals } from "./queries";
import type { FeedbackReviewer, UpdateFeedbackReviewerInput } from "./types";

export const useFeedbackPortals = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.portals(workspaceSlug),
    queryFn: () => getFeedbackPortals({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE * 5,
  });
};

export const useFeedbackBoardReviewers = (
  boardId: string,
  enabled: boolean,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.reviewers(workspaceSlug, boardId),
    queryFn: () =>
      getFeedbackBoardReviewers(boardId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session && boardId && enabled),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
};

export const useUpdateFeedbackPortalMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: ({
      portalId,
      input,
    }: {
      portalId: string;
      input: Parameters<typeof updateFeedbackPortal>[1];
    }) => updateFeedbackPortal(portalId, input, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to update portal", {
          description: response.error.message,
        });
        return;
      }
      toast.success("Portal updated");
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.portals(workspaceSlug),
      });
    },
    onError: (error) => {
      toast.error("Failed to update portal", {
        description: error.message,
      });
    },
  });
};

export const useCreateFeedbackBoardMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: (input: Parameters<typeof createFeedbackBoard>[0]) =>
      createFeedbackBoard(input, workspaceSlug),
    onSuccess: (response) => {
      if (response.error?.message) {
        toast.error("Failed to create board", {
          description: response.error.message,
        });
        return;
      }
      toast.success("Board created");
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.portals(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.teamSummaries(workspaceSlug),
      });
    },
    onError: (error) => {
      toast.error("Failed to create board", {
        description: error.message,
      });
    },
  });
};

export const useDeleteFeedbackBoardMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (boardId: string) => {
      const response = await deleteFeedbackBoard(boardId, workspaceSlug);
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      return boardId;
    },
    onSuccess: () => {
      toast.success("Board deleted");
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.portals(workspaceSlug),
      });
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.teamSummaries(workspaceSlug),
      });
    },
    onError: (error) => {
      toast.error("Failed to delete board", {
        description: error.message,
      });
    },
  });
};

export const useUpdateFeedbackBoardReviewerMutation = (boardId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const reviewersKey = feedbackKeys.reviewers(workspaceSlug, boardId);

  return useMutation({
    mutationFn: async ({
      userId,
      input,
    }: {
      userId: string;
      input: UpdateFeedbackReviewerInput;
    }) => {
      const response = await updateFeedbackBoardReviewer(
        boardId,
        userId,
        input,
        workspaceSlug,
      );
      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data) {
        throw new Error("The reviewer preference could not be updated");
      }
      return response.data;
    },
    onMutate: async ({ userId, input }) => {
      await queryClient.cancelQueries({ queryKey: reviewersKey });
      const reviewers =
        queryClient.getQueryData<FeedbackReviewer[]>(reviewersKey);
      const previousReviewer = reviewers?.find(
        (reviewer) => reviewer.userId === userId,
      );

      queryClient.setQueryData<FeedbackReviewer[]>(reviewersKey, (current) =>
        current?.map((reviewer) =>
          reviewer.userId === userId
            ? { ...reviewer, emailFrequency: input.emailFrequency }
            : reviewer,
        ),
      );

      return { previousReviewer };
    },
    onError: (error, { userId }, context) => {
      const previousReviewer = context?.previousReviewer;
      if (previousReviewer) {
        queryClient.setQueryData<FeedbackReviewer[]>(reviewersKey, (current) =>
          current?.map((reviewer) =>
            reviewer.userId === userId ? previousReviewer : reviewer,
          ),
        );
      }
      toast.error("Failed to update reviewer", {
        description: error.message || "Your changes were not saved",
      });
    },
    onSuccess: (reviewer) => {
      queryClient.setQueryData<FeedbackReviewer[]>(reviewersKey, (current) =>
        current?.map((candidate) =>
          candidate.userId === reviewer.userId ? reviewer : candidate,
        ),
      );
      toast.success("Reviewer updated");
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey: reviewersKey });
    },
  });
};
