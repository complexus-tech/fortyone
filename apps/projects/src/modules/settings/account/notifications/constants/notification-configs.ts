import { type NotificationType } from "@/modules/notifications/types";

export type NotificationConfig = {
  type: NotificationType;
  title: string;
  description: string;
};

export const notificationConfigs: NotificationConfig[] = [
  {
    type: "story_update",
    title: "Story updates",
    description: "Get notified when a story you're involved with is updated",
  },
  {
    type: "comment_reply",
    title: "Comments",
    description: "Get notified when someone comments on your stories",
  },
  {
    type: "mention",
    title: "Mentions",
    description: "Get notified when someone mentions you in a comment or story",
  },
  {
    type: "key_result_update",
    title: "Key Result updates",
    description: "Get notified about updates to key results",
  },
  {
    type: "story_comment",
    title: "Story comments",
    description: "Get notified when comments are added to stories",
  },
  {
    type: "objective_update",
    title: "Objective updates",
    description: "Get notified when objectives are updated",
  },
];
