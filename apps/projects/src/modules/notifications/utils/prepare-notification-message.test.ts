/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import type { AppNotification } from "../types";
import { prepareNotificationMessage } from "./prepare-notification-message";
import { renderTemplate } from "./render-template";

const ACTOR: NonNullable<AppNotification["actor"]> = {
  id: "actor-id",
  username: "hector",
  fullName: "Hector Smith",
  avatarUrl: "",
  isActive: true,
  isSystem: false,
};

describe("prepareNotificationMessage", () => {
  it("fills an existing actor placeholder without adding another prefix", () => {
    const message = {
      template: "{actor} changed priority to {value}",
      variables: { value: { value: "High", type: "value" } },
    };
    const result = prepareNotificationMessage(message, ACTOR, "story_update");

    expect(result.template).toBe(message.template);
    expect(renderTemplate(result).text).toBe(
      "Hector Smith changed priority to High",
    );
    expect(message.variables).not.toHaveProperty("actor");
  });

  it.each(["hector", "Hector Smith"])(
    "preserves a historical template with the literal actor name %s",
    (name) => {
      const message = {
        template: `${name} changed priority to High`,
        variables: {},
      };

      expect(prepareNotificationMessage(message, ACTOR, "story_update")).toBe(
        message,
      );
    },
  );

  it("adds Maya to the subjectless action returned by the API", () => {
    const message = {
      template: "scheduled this task to {scheduled_for}",
      variables: {
        scheduled_for: { value: "Wednesday at 11:00", type: "date" },
      },
    };
    const maya = {
      ...ACTOR,
      fullName: "Maya",
      username: "maya",
      isSystem: true,
    };

    expect(
      renderTemplate(prepareNotificationMessage(message, maya, "story_update"))
        .text,
    ).toBe("Maya scheduled this task to Wednesday at 11:00");
  });

  it("matches a literal actor prefix only at a complete name boundary", () => {
    const message = { template: "scheduled this task", variables: {} };
    const actor = { ...ACTOR, fullName: "schedule", username: "schedule" };

    expect(
      renderTemplate(prepareNotificationMessage(message, actor, "story_update"))
        .text,
    ).toBe("schedule scheduled this task");
  });

  it("preserves the actor supplied with the original message", () => {
    const message = {
      template: "{actor} updated the story",
      variables: { actor: { value: "Original name", type: "actor" } },
    };

    expect(prepareNotificationMessage(message, ACTOR, "story_update")).toBe(
      message,
    );
  });

  it("accepts null variables when filling a missing actor placeholder", () => {
    const message = { template: "{actor} updated the story", variables: null };

    expect(
      renderTemplate(prepareNotificationMessage(message, ACTOR, "story_update"))
        .text,
    ).toBe("Hector Smith updated the story");
  });

  it.each(["Hector Smith: Please review the launch plan", "Please review"])(
    "preserves the producer's complete chat preview: %s",
    (value) => {
      const message = {
        template: "{message}",
        variables: { message: { value, type: "plain_text" } },
      };

      expect(prepareNotificationMessage(message, ACTOR, "chat_message")).toBe(
        message,
      );
    },
  );

  it("does not add an actor to complete strategy guidance", () => {
    const message = {
      template: "Strategy guidance is ready to review.",
      variables: null,
    };

    expect(prepareNotificationMessage(message, ACTOR, "strategy_update")).toBe(
      message,
    );
  });

  it("preserves the message when no actor is available", () => {
    const message = { template: "Updated the story", variables: null };

    expect(prepareNotificationMessage(message, undefined, "story_update")).toBe(
      message,
    );
  });
});
