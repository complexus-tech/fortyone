/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getNotificationDetailsPath,
  getObjectiveDetailsPath,
  getSingleSearchParam,
  isNotificationEntityType,
} from "./notification-destination";

describe("notification destinations", () => {
  it("opens a document notification at its specific discussion", () => {
    const threadId = "55555555-5555-4555-8555-555555555555";
    const path = getNotificationDetailsPath({
      entityId: "document-1",
      entityType: "document",
      notificationId: "notification-1",
      message: {
        template: "Mentioned you",
        variables: { thread_id: { value: threadId } },
      },
    });
    const query = new URL(path, "https://fortyone.app").searchParams;
    expect(query.get("thread")).toBe(threadId);
    expect(query.get("entityType")).toBe("document");
  });
  it("builds an in-app notification detail path from the API contract", () => {
    expect(
      getNotificationDetailsPath({
        entityId: "objective-1",
        entityType: "objective",
        notificationId: "notification-1",
      }),
    ).toBe(
      "/notifications/notification-1?entityId=objective-1&entityType=objective",
    );
  });

  it("builds an overview deep link for a verified parent objective", () => {
    expect(
      getObjectiveDetailsPath({
        keyResultId: "key result/1",
        objectiveId: "objective/1",
        teamId: "team/1",
      }),
    ).toBe(
      "/teams/team%2F1/objectives/objective%2F1?tab=overview&keyResultId=key+result%2F1",
    );
  });

  it("preserves the displayed timestamp precision in a detail link", () => {
    const observedCreatedAt = "2026-10-03T20:00:00.123456789+02:00";
    const path = getNotificationDetailsPath({
      entityId: "story-1",
      entityType: "story",
      notificationId: "notification-1",
      observedCreatedAt,
    });
    expect(
      new URL(path, "https://fortyone.app").searchParams.get(
        "observedCreatedAt",
      ),
    ).toBe(observedCreatedAt);
  });

  it("rejects ambiguous and unsupported search parameters", () => {
    expect(
      getSingleSearchParam(["objective-1", "objective-2"]),
    ).toBeUndefined();
    expect(getSingleSearchParam("  objective-1  ")).toBe("objective-1");
    expect(isNotificationEntityType("key_result")).toBe(true);
    expect(isNotificationEntityType("feedback")).toBe(false);
  });

  it("targets the specific comment while retaining the story and notification version", () => {
    const commentId = "55555555-5555-4555-8555-555555555555";
    const path = getNotificationDetailsPath({
      entityId: "story-1",
      entityType: "story",
      notificationId: "notification-1",
      observedCreatedAt: "2026-10-16T09:00:00Z",
      message: {
        template: "{actor} replied: {content}",
        variables: { commentId: { value: commentId, type: "comment" } },
      },
    });
    const query = new URL(path, "https://fortyone.app").searchParams;
    expect(query.get("comment")).toBe(commentId);
    expect(query.get("entityId")).toBe("story-1");
    expect(query.get("observedCreatedAt")).toBe("2026-10-16T09:00:00Z");
  });

  it.each([
    undefined,
    "",
    "not-a-uuid",
    "00000000-0000-0000-0000-000000000000",
  ])("keeps older or invalid comment targets on the story (%s)", (value) => {
    const path = getNotificationDetailsPath({
      entityId: "story-1",
      entityType: "story",
      notificationId: "notification-1",
      message: value
        ? { template: "A comment", variables: { commentId: { value } } }
        : undefined,
    });
    expect(
      new URL(path, "https://fortyone.app").searchParams.has("comment"),
    ).toBe(false);
  });

  it("does not apply comment metadata to other resource types", () => {
    const path = getNotificationDetailsPath({
      entityId: "objective-1",
      entityType: "objective",
      notificationId: "notification-1",
      message: {
        template: "Updated objective",
        variables: {
          commentId: { value: "55555555-5555-4555-8555-555555555555" },
        },
      },
    });
    expect(
      new URL(path, "https://fortyone.app").searchParams.has("comment"),
    ).toBe(false);
  });
});
