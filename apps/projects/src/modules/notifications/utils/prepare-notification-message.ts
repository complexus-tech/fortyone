import type { AppNotification } from "../types";
import { htmlToPlainText } from "lib/src/html-to-plain-text";

type NotificationMessage = AppNotification["message"];

// These action phrases come from scheduleTransitionMessage and the Maya
// templates whose leading name is removed by normalizeSystemActorMessage.
const SUBJECTLESS_ACTION_PREFIXES = [
  "assigned you this task",
  "reassigned this task to",
  "updated this task's schedule",
  "needs a time estimate before scheduling this task",
  "flagged this task's schedule as at risk",
  "could not fit this task into the current schedule",
  "scheduled this task",
  "moved this task",
];

const startsWithPhrase = (template: string, phrase: string): boolean =>
  Boolean(phrase) &&
  template.startsWith(phrase) &&
  /^(?:\s|[.,:;!?]|$)/.test(template.slice(phrase.length));

export const prepareNotificationMessage = (
  message: NotificationMessage,
  actor: AppNotification["actor"],
  type: AppNotification["type"],
): NotificationMessage => {
  if (!actor || Object.hasOwn(message.variables ?? {}, "actor")) {
    return message;
  }

  const template = htmlToPlainText(message.template);
  const hasActorPlaceholder = template.includes("{actor}");

  if (!hasActorPlaceholder) {
    const hasLiteralActor = [actor.fullName, actor.username].some((name) =>
      startsWithPhrase(template, name),
    );
    const isSubjectlessAction = SUBJECTLESS_ACTION_PREFIXES.some((prefix) =>
      startsWithPhrase(template, prefix),
    );

    // Chat previews and complete sentences already own their subject. Only
    // supplement the producer's known subjectless action templates.
    if (type === "chat_message" || hasLiteralActor || !isSubjectlessAction) {
      return message;
    }
  }

  return {
    ...message,
    template: hasActorPlaceholder
      ? message.template
      : `{actor} ${message.template}`,
    variables: {
      ...message.variables,
      actor: {
        value: actor.fullName || actor.username || "Someone",
        type: "actor",
      },
    },
  };
};
