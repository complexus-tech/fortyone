import type { StoryPriority } from "@/shared/story/types";
import type { AppNotification } from "../types";

type NotificationIndicator =
  | { kind: "calendar" | "comment" | "mention" | "chat" | "status" }
  | { kind: "priority"; priority: StoryPriority };

type IndicatorNotification = {
  type: AppNotification["type"];
  message: {
    template: string;
    variables?: AppNotification["message"]["variables"] | null;
  };
};

const PRIORITIES: StoryPriority[] = [
  "No Priority",
  "Low",
  "Medium",
  "High",
  "Urgent",
];

export const getNotificationIndicator = ({
  type,
  message,
}: IndicatorNotification): NotificationIndicator | null => {
  if (type === "mention") return { kind: "mention" };
  if (type === "story_comment" || type === "comment_reply") {
    return { kind: "comment" };
  }
  if (type === "chat_message") return { kind: "chat" };
  if (type !== "story_update") return null;

  const messageVariables = message.variables ?? {};
  const variables = Object.values(messageVariables);
  const fields = variables
    .filter((variable) => variable.type === "field")
    .map((variable) => variable.value.trim().toLowerCase());

  // Classify the persisted event, never words in a user's name, title, or
  // comment. Older events use producer-owned templates without field metadata.
  const template = message.template.trim().replace(/: \{reason\}$/, "");
  if (
    variables.some((variable) => variable.type === "date") ||
    fields.some((field) => field === "deadline" || field.includes("date")) ||
    template === "{actor} changed the start date"
  ) {
    return { kind: "calendar" };
  }

  if (
    fields.includes("status") ||
    template === "{actor} moved the task to {value}" ||
    template === "{actor} moved the task from {previous_value} to {value}"
  ) {
    return { kind: "status" };
  }

  if (
    fields.includes("priority") ||
    template === "{actor} changed priority to {value}"
  ) {
    const value = Object.hasOwn(messageVariables, "value")
      ? messageVariables.value.value.trim().toLowerCase()
      : undefined;
    const priority = PRIORITIES.find(
      (candidate) => candidate.toLowerCase() === value,
    );
    return priority ? { kind: "priority", priority } : null;
  }

  return null;
};
