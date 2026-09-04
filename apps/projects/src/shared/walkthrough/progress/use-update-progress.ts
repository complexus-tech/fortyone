import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { userKeys } from "@/constants/keys";
import { useSession } from "@/lib/auth/client";
import { updateOnboardingTourProgressAction } from "./update-progress";
import type {
  OnboardingTourProgress,
  UpdateOnboardingTourProgress,
} from "./types";

const mergeUnique = (current: string[], updates: string[] | undefined) =>
  Array.from(new Set([...current, ...(updates ?? [])]));

const mergeStatus = (
  current: OnboardingTourProgress["status"],
  next: UpdateOnboardingTourProgress["status"],
) => {
  if (current !== "active") {
    return current;
  }

  return next ?? current;
};

export const useUpdateOnboardingTourProgressMutation = () => {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();

  const mutation = useMutation({
    mutationFn: (payload: UpdateOnboardingTourProgress) =>
      updateOnboardingTourProgressAction(payload, workspaceSlug),

    onMutate: async (payload) => {
      const queryKey = userKeys.onboardingTourProgress(
        session?.user.id ?? "anonymous",
        payload.tourKey,
        payload.tourVersion,
      );
      await queryClient.cancelQueries({ queryKey });

      const previousProgress =
        queryClient.getQueryData<OnboardingTourProgress>(queryKey);

      if (previousProgress) {
        queryClient.setQueryData<OnboardingTourProgress>(queryKey, {
          ...previousProgress,
          completedActionIds: mergeUnique(
            previousProgress.completedActionIds,
            payload.completedActionIds,
          ),
          completedStepIds: mergeUnique(
            previousProgress.completedStepIds,
            payload.completedStepIds,
          ),
          status: mergeStatus(previousProgress.status, payload.status),
        });
      }

      return { previousProgress, queryKey };
    },

    onError: (error, _payload, context) => {
      if (context?.previousProgress) {
        queryClient.setQueryData(context.queryKey, context.previousProgress);
      }

      toast.error("Couldn’t save onboarding progress", {
        description:
          error.message || "Your progress will be retried next time.",
      });
    },

    onSuccess: (progress, _payload, context) => {
      queryClient.setQueryData<OnboardingTourProgress>(
        context.queryKey,
        (currentProgress) => {
          if (!currentProgress) {
            return progress;
          }

          return {
            ...progress,
            completedActionIds: mergeUnique(
              progress.completedActionIds,
              currentProgress.completedActionIds,
            ),
            completedStepIds: mergeUnique(
              progress.completedStepIds,
              currentProgress.completedStepIds,
            ),
            status: mergeStatus(progress.status, currentProgress.status),
          };
        },
      );
    },
  });

  return mutation;
};
