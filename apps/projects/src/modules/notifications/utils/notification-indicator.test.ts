/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { AppNotification } from "../types";
import { getNotificationIndicator } from "./notification-indicator";

const storyUpdate = (
  template: string,
  variables: AppNotification["message"]["variables"] = {},
) => ({
  type: "story_update" as const,
  message: { template, variables },
});

describe("notification indicators", () => {
  it.each([null, undefined])(
    "accepts absent runtime variables (%s)",
    (variables) => {
      expect(
        getNotificationIndicator({
          type: "story_update",
          message: { template: "{actor} changed the start date", variables },
        }),
      ).toEqual({ kind: "calendar" });
      expect(
        getNotificationIndicator({
          type: "story_update",
          message: {
            template: "{actor} changed priority to {value}",
            variables,
          },
        }),
      ).toBeNull();
    },
  );

  it.each(["Low", "Medium", "High", "Urgent", "No Priority"])(
    "uses the actual %s priority from the notification",
    (priority) => {
      expect(
        getNotificationIndicator(
          storyUpdate("{actor} changed priority to {value}", {
            value: { value: priority, type: "value" },
          }),
        ),
      ).toEqual({ kind: "priority", priority });
    },
  );

  it("resolves typed priority fields and normalizes their value", () => {
    expect(
      getNotificationIndicator(
        storyUpdate("{actor} updated {field} to {value}", {
          field: { value: "Priority", type: "field" },
          value: { value: " low ", type: "value" },
        }),
      ),
    ).toEqual({ kind: "priority", priority: "Low" });
  });

  it.each<AppNotification["message"]["variables"]>([
    {},
    { value: { value: "Unknown", type: "value" } },
  ])("does not invent a priority for an incomplete event", (variables) => {
    expect(
      getNotificationIndicator(
        storyUpdate("{actor} changed priority to {value}", variables),
      ),
    ).toBeNull();
  });

  it.each([
    "{actor} moved the task to {value}",
    "{actor} moved the task from {previous_value} to {value}",
    "{actor} moved the task from {previous_value} to {value}: {reason}",
  ])("recognizes a status event: %s", (template) => {
    expect(
      getNotificationIndicator(
        storyUpdate(template, {
          value: { value: "In progress", type: "value" },
        }),
      ),
    ).toEqual({ kind: "status" });
  });

  it("recognizes a legacy status field", () => {
    expect(
      getNotificationIndicator(
        storyUpdate("{actor} updated {field}", {
          field: { value: "Status", type: "field" },
        }),
      ),
    ).toEqual({ kind: "status" });
  });

  it.each([
    storyUpdate("{actor} changed the start date"),
    storyUpdate("{actor} removed the {field}", {
      field: { value: "deadline", type: "field" },
    }),
    storyUpdate("{actor} set the {field} to {value}", {
      value: { value: "8 Sep", type: "date" },
    }),
    storyUpdate("moved this task to {scheduled_for}: {reason}", {
      scheduled_for: { value: "Wednesday at 11:00", type: "date" },
    }),
  ])("recognizes date and scheduling events", (notification) => {
    expect(getNotificationIndicator(notification)).toEqual({
      kind: "calendar",
    });
  });

  it.each([
    ["story_comment", "comment"],
    ["comment_reply", "comment"],
    ["mention", "mention"],
    ["chat_message", "chat"],
  ] as const)("uses the event type for %s", (type, kind) => {
    expect(
      getNotificationIndicator({
        type,
        message: {
          template: "{actor} replied: {content}",
          variables: {
            content: {
              value: "Change the priority, status, and deadline",
              type: "text",
            },
            date: { value: "8 Sep", type: "date" },
          },
        },
      }),
    ).toEqual({ kind });
  });

  it.each([
    "{actor} assigned you a task",
    "{actor} renamed the story to {value}",
  ])("ignores user-authored words when classifying %s", (template) => {
    expect(
      getNotificationIndicator(
        storyUpdate(template, {
          actor: { value: "Priority deadline status", type: "actor" },
          value: {
            value: "Fix priority status before the deadline",
            type: "value",
          },
        }),
      ),
    ).toBeNull();
  });
});
