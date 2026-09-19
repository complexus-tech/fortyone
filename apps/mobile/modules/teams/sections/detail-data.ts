import type { FeedbackItem, IntakeItem } from "./types";
import { useQuery } from "@tanstack/react-query";
import {
  feedbackKeys,
  intakeKeys,
  storyKeys,
  homeKeys,
  searchKeys,
} from "@/constants/keys";
import { readEntity, writeEntity } from "@/modules/entity-details/http";
import { useDetailMutation } from "@/modules/entity-details/use-detail-mutation";

export type FeedbackCommand =
  | {
      type: "status";
      status: FeedbackItem["status"];
      explanation?: string | null;
    }
  | { type: "plan"; teamId: string; storyId?: string }
  | { type: "comment"; body: string; parentId?: string }
  | { type: "read" | "unread" | "trash" | "restore" }
  | { type: "merge"; targetItemId: string };
export type FeedbackResult = { storyId?: string; target?: FeedbackItem };
export function useFeedbackCommands(id: string) {
  return useDetailMutation(
    async (command: FeedbackCommand) => {
      const base = `feedback/items/${encodeURIComponent(id)}`;
      switch (command.type) {
        case "status":
          return writeEntity<FeedbackResult>("put", `${base}/status`, {
            status: command.status,
            roadmapSummary: command.explanation ?? null,
          });
        case "plan":
          return writeEntity<FeedbackResult>("post", `${base}/story`, {
            teamId: command.teamId,
            storyId: command.storyId,
          });
        case "comment":
          return writeEntity<FeedbackResult>("post", `${base}/comments`, {
            body: command.body,
            parentId: command.parentId,
          });
        case "read":
        case "unread":
          return writeEntity<FeedbackResult>(
            "put",
            `${base}/${command.type}`,
            {},
          );
        case "trash":
          return writeEntity<FeedbackResult>("delete", base);
        case "restore":
          return writeEntity<FeedbackResult>("post", `${base}/restore`, {});
        case "merge":
          return writeEntity<FeedbackResult>("post", `${base}/merge`, {
            targetItemId: command.targetItemId,
          });
      }
    },
    [feedbackKeys.all, storyKeys.all, homeKeys.all, searchKeys.all],
    (command) => {
      const now = new Date().toISOString();
      switch (command.type) {
        case "status":
          return { entityId: id, patch: { status: command.status } };
        case "read":
          return { entityId: id, patch: { readAt: now } };
        case "unread":
          return { entityId: id, patch: { readAt: null } };
        case "trash":
          return { entityId: id, patch: { deletedAt: now } };
        case "restore":
          return { entityId: id, patch: { deletedAt: null } };
        default:
          return null;
      }
    },
  );
}
export type IntakePatch = Partial<
  Pick<
    IntakeItem,
    | "title"
    | "description"
    | "statusId"
    | "priority"
    | "assigneeId"
    | "objectiveId"
    | "sprintId"
    | "startDate"
    | "endDate"
    | "labelIds"
  >
>;
type IntakeCommand =
  | { type: "update"; patch: IntakePatch }
  | { type: "accept" | "decline" }
  | {
      type: "comment";
      provider: IntakeItem["provider"];
      body: string;
      idempotencyKey: string;
    };
export function useIntakeCommands(id: string) {
  return useDetailMutation(
    async (command: IntakeCommand) => {
      const base = `integration-requests/${encodeURIComponent(id)}`;
      if (command.type === "update")
        return writeEntity<IntakeItem>("put", base, command.patch);
      if (command.type === "comment") {
        await writeEntity(
          "post",
          `${base}/${command.provider === "github" ? "github-comments" : "comments"}`,
          {
            body: command.body,
            ...(command.provider === "slack"
              ? { idempotencyKey: command.idempotencyKey }
              : {}),
          },
        );
        return undefined;
      }
      return writeEntity<IntakeItem>("post", `${base}/${command.type}`, {});
    },
    [intakeKeys.all, storyKeys.all, homeKeys.all, searchKeys.all],
    (command) => {
      switch (command.type) {
        case "update":
          return { entityId: id, patch: command.patch };
        case "accept":
          return { entityId: id, patch: { status: "accepted" } };
        case "decline":
          return { entityId: id, patch: { status: "declined" } };
        default:
          return null;
      }
    },
  );
}
type ProviderComment = {
  id: string;
  authorName: string;
  authorAvatar?: string;
  body: string;
  createdAt: string;
  deliveryStatus?: string;
};
type GitHubComment = {
  id: number;
  userLogin: string;
  userAvatar?: string;
  body: string;
  createdAt: string;
};
export function useIntakeComments(item: IntakeItem) {
  return useQuery({
    queryKey: [...intakeKeys.detail(item.id), "comments", item.provider],
    enabled: item.provider === "github" || item.provider === "slack",
    queryFn: async ({ signal }): Promise<ProviderComment[]> => {
      const base = `integration-requests/${encodeURIComponent(item.id)}`;
      if (item.provider === "github") {
        const comments = await readEntity<GitHubComment[]>(
          `${base}/github-comments`,
          signal,
        );
        return comments.map((comment) => ({
          ...comment,
          id: String(comment.id),
          authorName: comment.userLogin,
          authorAvatar: comment.userAvatar,
        }));
      }
      const thread = await readEntity<{ comments: ProviderComment[] }>(
        `${base}/thread`,
        signal,
      );
      return thread.comments ?? [];
    },
  });
}
