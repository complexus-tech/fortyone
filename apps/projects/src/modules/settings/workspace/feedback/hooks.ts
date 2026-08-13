import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { feedbackKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks";
import { DURATION_FROM_MILLISECONDS } from "@/constants/time";
import {
  createFeedbackUpdate,
  createFeedbackBoard,
  createFeedbackWidgetSigningSecret,
  deleteFeedbackUpdate,
  deleteFeedbackBoard,
  publishFeedbackUpdate,
  rotateFeedbackWidgetSigningSecret,
  unpublishFeedbackUpdate,
  updateFeedbackUpdate,
  updateFeedbackBoardReviewer,
  updateFeedbackPortal,
  updateFeedbackWidgetSettings,
} from "./actions";
import {
  getFeedbackBoardReviewers,
  getFeedbackPortals,
  getFeedbackUpdateCandidates,
  getFeedbackUpdates,
  getFeedbackWidgetSettings,
} from "./queries";
import type {
  FeedbackReviewer,
  FeedbackWidgetSigningSecret,
  UpdateFeedbackReviewerInput,
} from "./types";

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

export const useFeedbackWidgetSettings = (portalId: string) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.widgetSettings(workspaceSlug, portalId),
    queryFn: () =>
      getFeedbackWidgetSettings(portalId, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session && portalId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
};

export const useFeedbackUpdates = () => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  return useQuery({
    queryKey: feedbackKeys.updates(workspaceSlug),
    queryFn: () => getFeedbackUpdates({ session: session!, workspaceSlug }),
    enabled: Boolean(session),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
  });
};

export const useFeedbackUpdateCandidates = (
  portalId: string,
  search: string,
) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const normalizedSearch = search.trim();

  return useQuery({
    queryKey: feedbackKeys.updateCandidates(
      workspaceSlug,
      portalId,
      normalizedSearch,
    ),
    queryFn: () =>
      getFeedbackUpdateCandidates(portalId, normalizedSearch, {
        session: session!,
        workspaceSlug,
      }),
    enabled: Boolean(session && portalId),
    staleTime: DURATION_FROM_MILLISECONDS.MINUTE,
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

export const useUpdateFeedbackWidgetSettingsMutation = (portalId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = feedbackKeys.widgetSettings(workspaceSlug, portalId);

  return useMutation({
    mutationFn: async (
      input: Parameters<typeof updateFeedbackWidgetSettings>[1],
    ) => {
      const response = await updateFeedbackWidgetSettings(
        portalId,
        input,
        workspaceSlug,
      );
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ?? "Widget settings were not returned",
        );
      }
      return response.data;
    },
    onSuccess: (settings) => {
      queryClient.setQueryData(queryKey, settings);
      toast.success("Widget settings saved");
    },
    onError: (error) => {
      toast.error("Failed to save widget settings", {
        description: error.message,
      });
    },
  });
};

const useFeedbackWidgetSecretMutation = (
  portalId: string,
  operation: "create" | "rotate",
  onReveal: (secret: FeedbackWidgetSigningSecret) => void,
) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = feedbackKeys.widgetSettings(workspaceSlug, portalId);

  return useMutation({
    mutationFn: async () => {
      const response = await (operation === "create"
        ? createFeedbackWidgetSigningSecret(portalId, workspaceSlug)
        : rotateFeedbackWidgetSigningSecret(portalId, workspaceSlug));
      if (response.error?.message || !response.data) {
        throw new Error(
          response.error?.message ?? "A signing secret was not returned",
        );
      }
      return response.data;
    },
    onSuccess: (result) => {
      const { signingSecret: _, ...settings } = result;
      queryClient.setQueryData(queryKey, settings);
      onReveal(result);
      toast.success(
        operation === "create"
          ? "Signing secret created"
          : "Signing secret rotated",
      );
    },
    onError: (error) => {
      toast.error("Failed to update signing secret", {
        description: error.message,
      });
    },
  });
};

export const useCreateFeedbackWidgetSecretMutation = (
  portalId: string,
  onReveal: (secret: FeedbackWidgetSigningSecret) => void,
) => useFeedbackWidgetSecretMutation(portalId, "create", onReveal);

export const useRotateFeedbackWidgetSecretMutation = (
  portalId: string,
  onReveal: (secret: FeedbackWidgetSigningSecret) => void,
) => useFeedbackWidgetSecretMutation(portalId, "rotate", onReveal);

export const useCreateFeedbackUpdateMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (input: Parameters<typeof createFeedbackUpdate>[0]) => {
      const response = await createFeedbackUpdate(input, workspaceSlug);
      if (response.error?.message || !response.data) {
        throw new Error(response.error?.message ?? "Update was not returned");
      }
      return response.data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.updates(workspaceSlug),
      });
      toast.success("Update draft created");
    },
    onError: (error) => {
      toast.error("Failed to create update", { description: error.message });
    },
  });
};

export const useUpdateFeedbackUpdateMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      input,
      updateId,
    }: {
      input: Parameters<typeof updateFeedbackUpdate>[1];
      updateId: string;
    }) => {
      const response = await updateFeedbackUpdate(
        updateId,
        input,
        workspaceSlug,
      );
      if (response.error?.message || !response.data) {
        throw new Error(response.error?.message ?? "Update was not returned");
      }
      return response.data;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.updates(workspaceSlug),
      });
      toast.success("Update saved");
    },
    onError: (error) => {
      toast.error("Failed to save update", { description: error.message });
    },
  });
};

export const useDeleteFeedbackUpdateMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async (updateId: string) => {
      const response = await deleteFeedbackUpdate(updateId, workspaceSlug);
      if (response.error?.message) throw new Error(response.error.message);
      return updateId;
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.updates(workspaceSlug),
      });
      toast.success("Update deleted");
    },
    onError: (error) => {
      toast.error("Failed to delete update", { description: error.message });
    },
  });
};

export const usePublishFeedbackUpdateMutation = () => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();

  return useMutation({
    mutationFn: async ({
      publish,
      updateId,
    }: {
      publish: boolean;
      updateId: string;
    }) => {
      const response = await (publish
        ? publishFeedbackUpdate(updateId, workspaceSlug)
        : unpublishFeedbackUpdate(updateId, workspaceSlug));
      if (response.error?.message || !response.data) {
        throw new Error(response.error?.message ?? "Update was not returned");
      }
      return response.data;
    },
    onSuccess: (update) => {
      void queryClient.invalidateQueries({
        queryKey: feedbackKeys.updates(workspaceSlug),
      });
      toast.success(
        update.publishedAt ? "Update published" : "Update unpublished",
      );
    },
    onError: (error) => {
      toast.error("Failed to change publication status", {
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
