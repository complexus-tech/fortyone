import type { ApiResponse } from "@/types";
import { auth } from "@/auth";
import { patch, post, put, remove } from "@/lib/http";
import type {
  ChecklistCommand,
  ChecklistMutationResult,
} from "../types/checklist";

export const updateChecklistAction = async (
  storyId: string,
  command: ChecklistCommand,
  workspaceSlug: string,
): Promise<ChecklistMutationResult> => {
  const session = await auth();
  if (!session) throw new Error("Please sign in to update this checklist.");

  const ctx = { session, workspaceSlug };
  const path = `stories/${storyId}/checklist`;
  const options = { retry: 0 };
  let response: ApiResponse<ChecklistMutationResult>;

  switch (command.type) {
    case "create":
      response = await post<
        { title: string },
        ApiResponse<ChecklistMutationResult>
      >(path, { title: command.title }, ctx, options);
      break;
    case "update":
      response = await patch<
        typeof command.updates,
        ApiResponse<ChecklistMutationResult>
      >(`${path}/${command.itemId}`, command.updates, ctx, options);
      break;
    case "delete":
      response = await remove<ApiResponse<ChecklistMutationResult>>(
        `${path}/${command.itemId}`,
        ctx,
        options,
      );
      break;
    case "reorder":
      response = await put<
        { itemIds: string[] },
        ApiResponse<ChecklistMutationResult>
      >(`${path}/order`, { itemIds: command.itemIds }, ctx, options);
      break;
    case "clear":
      response = await remove<ApiResponse<ChecklistMutationResult>>(
        `${path}/items`,
        ctx,
        options,
      );
      break;
    case "convert":
      response = await post<
        Record<string, never>,
        ApiResponse<ChecklistMutationResult>
      >(`${path}/${command.itemId}/convert`, {}, ctx, options);
      break;
  }

  if (response.error) throw new Error(response.error.message);
  if (command.type !== "delete" && command.type !== "clear" && !response.data) {
    throw new Error(
      "The checklist response was incomplete. Refresh to check your changes.",
    );
  }
  return response.data ?? null;
};
