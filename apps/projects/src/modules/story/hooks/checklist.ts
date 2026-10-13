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

const activeChecklistWrites = new WeakMap<object, Set<string>>();

export const useStoryChecklist = (storyId: string) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();

  return useQuery({
    queryKey: storyKeys.checklist(workspaceSlug, storyId),
    queryFn: ({ signal }) =>
      getStoryChecklist(storyId, { workspaceSlug, session }, signal),
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
    { previousItems: StoryChecklistItem[] | undefined }
  >({
    mutationKey,
    retry: false,
    mutationFn: (command) =>
      updateChecklistAction(storyId, command, workspaceSlug),
    onMutate: async (command) => {
      // Story details can be mounted in more than one surface at a time.
      // Reject overlapping writes before applying optimism or sending a request.
      const activeWrites =
        activeChecklistWrites.get(queryClient) ?? new Set<string>();
      if (activeWrites.has(writeKey)) {
        throw new Error("Wait for the current checklist change to finish.");
      }
      activeWrites.add(writeKey);
      activeChecklistWrites.set(queryClient, activeWrites);
      try {
        await queryClient.cancelQueries({ queryKey });
      } catch (error) {
        activeWrites.delete(writeKey);
        throw error;
      }
      const previousItems =
        queryClient.getQueryData<StoryChecklistItem[]>(queryKey);
      if (previousItems) {
        queryClient.setQueryData(
          queryKey,
          applyOptimisticCommand(previousItems, command),
        );
      }
      return { previousItems };
    },
    onSuccess: (result, command) => {
      if (Array.isArray(result)) {
        queryClient.setQueryData(queryKey, result);
      } else if (result) {
        queryClient.setQueryData<StoryChecklistItem[]>(
          queryKey,
          (items = []) =>
            command.type === "create"
              ? [...items.filter((item) => item.id !== result.id), result]
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
        toast.success("Checklist item converted to a story");
        const createdStory = Array.isArray(result)
          ? null
          : result?.convertedStory;
        if (createdStory?.statusId) {
          void refreshWipCapacity({
            queryClient,
            workspaceSlug,
            loadCapacity: loadWipCapacity,
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
      if (context?.previousItems) {
        queryClient.setQueryData(queryKey, context.previousItems);
      }
      toast.error("Unable to save checklist changes", {
        description: `${error.message} Refresh to check the latest checklist before trying again.`,
      });
    },
    onSettled: (_data, _error, _command, context) => {
      if (!context) return;
      return queryClient.invalidateQueries({ queryKey }).finally(() => {
        activeChecklistWrites.get(queryClient)?.delete(writeKey);
      });
    },
  });

  return { ...mutation, isPending: mutation.isPending || pendingCount > 0 };
};
