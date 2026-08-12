/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  getNotificationDetailsPath,
  getObjectiveDetailsPath,
  getSingleSearchParam,
  isNotificationEntityType,
} from "./notification-destination";

describe("notification destinations", () => {
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

  it("rejects ambiguous and unsupported search parameters", () => {
    expect(
      getSingleSearchParam(["objective-1", "objective-2"]),
    ).toBeUndefined();
    expect(getSingleSearchParam("  objective-1  ")).toBe("objective-1");
    expect(isNotificationEntityType("key_result")).toBe(true);
    expect(isNotificationEntityType("feedback")).toBe(false);
  });
});
