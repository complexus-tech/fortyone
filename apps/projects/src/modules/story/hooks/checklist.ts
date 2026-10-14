import {
  useIsMutating,
  useMutation,
  useQuery,
  useQueryClient,
} from "@tanstack/react-query";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks";
import { useSession } from "@/lib/auth/client";
import { storyKeys } from "@/modules/stories/constants";
import { loadWipCapacity } from "@/modules/stories/public/wip-capacity";
import { refreshWipCapacity } from "@/shared/story/wip-capacity";
import type {
  ChecklistCommand,
  ChecklistMutationResult,
  StoryChecklistItem,
} from "../types/checklist";
import { updateChecklistAction } from "../actions/checklist";
import { getStoryChecklist } from "../queries/get-checklist";

type ChecklistWrite = {
  settled: Promise<void>;
  releaseReads: () => void;
};

const activeChecklistWrites = new WeakMap<
  object,
  Map<string, ChecklistWrite>
>();

export const useStoryChecklist = (storyId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const queryKey = storyKeys.checklist(workspaceSlug, storyId);

  return useQuery({
    queryKey,
    queryFn: async ({ signal }) => {
      // Realtime and focus refreshes must not overwrite an in-flight change.
      await activeChecklistWrites
        .get(queryClient)
        ?.get(JSON.stringify(queryKey))?.settled;
      if (signal.aborted) {
        throw new DOMException(
          "The checklist refresh was cancelled.",
          "AbortError",
        );
      }
      return getStoryChecklist(storyId, { workspaceSlug, session }, signal);
    },
    enabled: Boolean(session && storyId),
    staleTime: 30_000,
  });
};

const applyOptimisticCommand = (
  items: StoryChecklistItem[],
  command: ChecklistCommand,
): StoryChecklistItem[] => {
  switch (command.type) {
    case "update":
      return items.map((item) =>
        item.id === command.itemId ? { ...item, ...command.updates } : item,
      );
    case "delete":
      return items.filter((item) => item.id !== command.itemId);
    case "clear":
      return [];
    case "reorder": {
      const byId = new Map(items.map((item) => [item.id, item]));
      // A concurrent addition must never disappear from the local checklist.
      if (
        command.itemIds.length !== items.length ||
        new Set(command.itemIds).size !== items.length ||
        command.itemIds.some((id) => !byId.has(id))
      ) {
        return items;
      }
      return command.itemIds.map((id, position) => ({
        ...byId.get(id)!,
        position,
      }));
    }
    default:
      return items;
  }
};

export const useChecklistMutation = (storyId: string) => {
  const queryClient = useQueryClient();
  const { workspaceSlug } = useWorkspacePath();
  const queryKey = storyKeys.checklist(workspaceSlug, storyId);
  const mutationKey = [...queryKey, "mutation"];
  const writeKey = JSON.stringify(queryKey);
  const pendingCount = useIsMutating({ mutationKey });

  const mutation = useMutation<
    ChecklistMutationResult,
    Error,
    ChecklistCommand,
    {
      previousItems: StoryChecklistItem[] | undefined;
      optimisticItemId?: string;
    }
  >({
    mutationKey,
    retry: false,
    mutationFn: (command) =>
      updateChecklistAction(storyId, command, workspaceSlug),
    onMutate: async (command) => {
      // Story details can be mounted in more than one surface at a time.
      // Reject overlapping writes before applying optimism or sending a request.
      const activeWrites =
        activeChecklistWrites.get(queryClient) ??
        new Map<string, ChecklistWrite>();
      if (activeWrites.has(writeKey)) {
        throw new Error("Wait for the current checklist change to finish.");
      }
      let releaseReads!: () => void;
      const settled = new Promise<void>((resolve) => {
        releaseReads = resolve;
      });
      activeWrites.set(writeKey, { settled, releaseReads });
      activeChecklistWrites.set(queryClient, activeWrites);
      try {
        await queryClient.cancelQueries({ queryKey });
      } catch (error) {
        releaseReads();
        activeWrites.delete(writeKey);
        throw error;
      }
      const previousItems =
        queryClient.getQueryData<StoryChecklistItem[]>(queryKey);
      if (command.type === "create") {
        const items = previousItems ?? [];
        const now = new Date().toISOString();
        const optimisticItem: StoryChecklistItem = {
          id: crypto.randomUUID(),
          storyId,
          title: command.title,
          completed: false,
          position: Math.max(-1, ...items.map((item) => item.position)) + 1,
          convertedAt: null,
          convertedStory: null,
          createdAt: now,
          updatedAt: now,
        };
        queryClient.setQueryData(queryKey, [...items, optimisticItem]);
        return { previousItems, optimisticItemId: optimisticItem.id };
      }
      if (previousItems) {
        queryClient.setQueryData(
          queryKey,
          applyOptimisticCommand(previousItems, command),
        );
      }
      return { previousItems };
    },
    onSuccess: (result, command, context) => {
      if (Array.isArray(result)) {
        queryClient.setQueryData(queryKey, result);
      } else if (result) {
        queryClient.setQueryData<StoryChecklistItem[]>(
          queryKey,
          (items = []) =>
            command.type === "create"
              ? [
                  ...items.filter(
                    (item) =>
                      item.id !== context.optimisticItemId &&
                      item.id !== result.id,
                  ),
                  result,
                ]
              : items.map((item) => (item.id === result.id ? result : item)),
        );
      }
      if (command.type === "convert") {
        void queryClient.invalidateQueries({
          queryKey: storyKeys.all(workspaceSlug),
        });
        void queryClient.invalidateQueries({
          queryKey: storyKeys.total(workspaceSlug),
        });
        const createdStory = Array.isArray(result)
          ? null
          : result?.convertedStory;
        if (createdStory?.statusId) {
          void refreshWipCapacity({
            queryClient,
            workspaceSlug,
            loadCapacity: loadWipCapacity,
            showCapacityWarning: false,
            createdStory: {
              ...createdStory,
              statusId: createdStory.statusId,
              parentId: null,
              archivedAt: null,
              deletedAt: null,
            },
          });
        }
      }
    },
    onError: (error, _command, context) => {
      if (context) {
        queryClient.setQueryData(queryKey, context.previousItems ?? []);
      }
      toast.error("Unable to save checklist changes", {
        description: `${error.message} Refresh to check the latest checklist before trying again.`,
      });
    },
    onSettled: (_data, _error, _command, context) => {
      if (!context) return;
      activeChecklistWrites.get(queryClient)?.get(writeKey)?.releaseReads();
      // Refresh committed history even if realtime is unavailable or the save
      // response was lost after the server committed the change.
      void queryClient.invalidateQueries({
        queryKey: storyKeys.activitiesInfinite(workspaceSlug, storyId),
        refetchType: "all",
      });
      return queryClient.invalidateQueries({ queryKey }).finally(() => {
        activeChecklistWrites.get(queryClient)?.delete(writeKey);
      });
    },
  });

  return { ...mutation, isPending: mutation.isPending || pendingCount > 0 };
};
