import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { integrationRequestKeys } from "@/constants/keys";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { postIntegrationRequestCommentAction } from "../actions/post-comment";
import type {
  IntegrationRequestComment,
  IntegrationRequestThreadActivity,
} from "../types";

export const usePostIntegrationRequestComment = (requestId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const queryKey = integrationRequestKeys.thread(workspaceSlug, requestId);

  return useMutation({
    mutationFn: async ({
      body,
      idempotencyKey,
    }: {
      body: string;
      idempotencyKey: string;
    }) => {
      const response = await postIntegrationRequestCommentAction(
        requestId,
        body,
        idempotencyKey,
        workspaceSlug,
      );
      if (response.error?.message || !response.data) {
        throw new Error(response.error?.message ?? "Comment was not returned");
      }
      return response.data;
    },
    onMutate: async ({ body }) => {
      await queryClient.cancelQueries({ queryKey });
      const previous =
        queryClient.getQueryData<IntegrationRequestThreadActivity>(queryKey);
      const optimistic: IntegrationRequestComment = {
        id: `optimistic-${Date.now()}`,
        threadId: previous?.thread.id ?? "",
        direction: "outbound",
        authorUserId: session?.user.id,
        authorName: session?.user.name ?? "You",
        authorAvatar: session?.user.image ?? undefined,
        deliveryStatus: "sending",
        body,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      if (previous) {
        queryClient.setQueryData<IntegrationRequestThreadActivity>(queryKey, {
          ...previous,
          comments: [...previous.comments, optimistic],
        });
      }
      return { previous, optimistic };
    },
    onError: (error, _input, context) => {
      if (context?.previous)
        queryClient.setQueryData(queryKey, context.previous);
      toast.error("Failed to post to Slack", { description: error.message });
    },
    onSuccess: (comment, _input, context) => {
      queryClient.setQueryData<IntegrationRequestThreadActivity>(
        queryKey,
        (data) =>
          data
            ? {
                ...data,
                comments: data.comments.map((cached) =>
                  cached.id === context.optimistic.id ? comment : cached,
                ),
              }
            : data,
      );
    },
    onSettled: () => {
      void queryClient.invalidateQueries({ queryKey });
    },
    retry: 2,
  });
};
