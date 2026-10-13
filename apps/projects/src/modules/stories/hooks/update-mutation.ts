import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useParams } from "next/navigation";
import { useAnalytics, useTerminology, useWorkspacePath } from "@/hooks";
import { objectiveKeys } from "@/shared/objectives/keys";
import type { WipStorySnapshot } from "@/shared/story/wip-capacity";
import {
  affectsWipCapacity,
  captureWipStorySnapshots,
  isWipCapacityQuery,
  refreshWipCapacity,
} from "@/shared/story/wip-capacity";
import { storyKeys } from "../constants";
import type { DetailedStory } from "../types";
import { bulkUpdateAction } from "../actions/bulk-update-stories";
import { loadWipCapacity } from "../public/wip-capacity";
import {
  applyBulkStoryUpdate,
  restoreBulkStoryUpdate,
} from "../utils/bulk-optimistic";
import type { BulkStoryCacheSnapshot } from "../utils/bulk-optimistic";
import {
  assertBulkStoryUpdateSucceeded,
  BulkStoryUpdateFailure,
} from "./bulk-update-result";

type BulkUpdateVariables = {
  storyIds: string[];
  payload: Partial<DetailedStory>;
};

type BulkUpdateContext = {
  previousQueryStates: BulkStoryCacheSnapshot;
  wipSnapshots: WipStorySnapshot[];
};

export const useBulkUpdateStoriesMutation = () => {
  const queryClient = useQueryClient();
  const { storyId } = useParams<{ storyId?: string }>();
  const { workspaceSlug } = useWorkspacePath();
  const { analytics } = useAnalytics();
  const { getTermDisplay } = useTerminology();

  const mutation = useMutation({
    mutationFn: async ({ storyIds, payload }: BulkUpdateVariables) => {
      const response = await bulkUpdateAction(
        { storyIds: Array.from(new Set(storyIds)), updates: payload },
        workspaceSlug,
      );

      if (response.error?.message) {
        throw new Error(response.error.message);
      }
      if (!response.data) {
        throw new Error("The bulk update returned no result");
      }

      return assertBulkStoryUpdateSucceeded(response.data);
    },

    onMutate: async ({ storyIds, payload }): Promise<BulkUpdateContext> => {
      const wipSnapshots = await captureWipStorySnapshots(
        queryClient,
        workspaceSlug,
        storyIds,
        payload,
      );
      await queryClient.cancelQueries({
        queryKey: storyKeys.all(workspaceSlug),
      });
      const previousQueryStates = applyBulkStoryUpdate({
        queryClient,
        workspaceSlug,
        storyIds,
        payload,
        parentStoryId: storyId,
      });

      return { previousQueryStates, wipSnapshots };
    },

    onError: (error, variables, context) => {
      if (context) {
        restoreBulkStoryUpdate(queryClient, context.previousQueryStates);
      }

      const itemFailure =
        error instanceof BulkStoryUpdateFailure ? error : null;
      const hasPersistedChanges = Boolean(
        itemFailure?.successfulStoryIds.length,
      );
      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
        predicate: (query) =>
          !hasPersistedChanges || !isWipCapacityQuery(query, workspaceSlug),
      });
      queryClient.invalidateQueries({
        queryKey: objectiveKeys.list(workspaceSlug),
      });

      if (hasPersistedChanges && affectsWipCapacity(variables.payload)) {
        void refreshWipCapacity({
          queryClient,
          workspaceSlug,
          loadCapacity: loadWipCapacity,
          snapshots: context?.wipSnapshots,
          payload: variables.payload,
          successfulStoryIds: itemFailure?.successfulStoryIds,
        });
      }
      const retryStoryIds =
        itemFailure &&
        itemFailure.failedStoryIds.length === itemFailure.failedCount
          ? itemFailure.failedStoryIds
          : variables.storyIds;
      const failureTitle = itemFailure
        ? `Failed to update ${itemFailure.failedCount} of ${itemFailure.totalCount} ${getTermDisplay(
            "storyTerm",
            {
              variant: itemFailure.totalCount === 1 ? "singular" : "plural",
            },
          )}`
        : `Failed to update ${getTermDisplay("storyTerm", { variant: "plural" })}`;

      toast.error(failureTitle, {
        description: error.message || "Your changes were not saved",
        action: {
          label: "Retry",
          onClick: () => {
            mutation.mutate({
              ...variables,
              storyIds: retryStoryIds,
            });
          },
        },
      });
    },

    onSuccess: (_result, { storyIds, payload }, context) => {
      const uniqueStoryIds = Array.from(new Set(storyIds));
      analytics.track("stories_bulk_updated", {
        storyIds: uniqueStoryIds,
        count: uniqueStoryIds.length,
        ...payload,
      });

      queryClient.invalidateQueries({
        queryKey: storyKeys.all(workspaceSlug),
        predicate: (query) =>
          !affectsWipCapacity(payload) ||
          !isWipCapacityQuery(query, workspaceSlug),
      });
      if (affectsWipCapacity(payload)) {
        void refreshWipCapacity({
          queryClient,
          workspaceSlug,
          loadCapacity: loadWipCapacity,
          snapshots: context.wipSnapshots,
          payload,
          successfulStoryIds: uniqueStoryIds,
        });
      }
    },
  });

  return mutation;
};
