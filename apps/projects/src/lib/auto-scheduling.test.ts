/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  canLockAutoSchedulingStatus,
  canToggleAutoSchedulingLock,
  deriveAutoSchedulingStatus,
  getAutoSchedulingHelper,
  getNewStoryAutoSchedulingEnabled,
  isMayaAssigneeSelection,
} from "./auto-scheduling";

describe("auto-scheduling state", () => {
  it("keeps human-assigned work off until someone explicitly enables it", () => {
    expect(
      deriveAutoSchedulingStatus({
        assigneeId: "user-1",
        autoSchedulingEnabled: false,
        estimatedDurationMinutes: 60,
      }),
    ).toBe("off");
  });

  it("guides an enabled story through its missing inputs", () => {
    expect(deriveAutoSchedulingStatus({ autoSchedulingEnabled: true })).toBe(
      "needs_owner",
    );
    expect(
      deriveAutoSchedulingStatus({
        assigneeId: "user-1",
        autoSchedulingEnabled: true,
      }),
    ).toBe("needs_time");
    expect(
      deriveAutoSchedulingStatus({
        assigneeId: "user-1",
        autoSchedulingEnabled: true,
        estimatedDurationMinutes: 90,
      }),
    ).toBe("planning");
  });

  it("gives locking precedence over transient planning state", () => {
    expect(
      deriveAutoSchedulingStatus({
        autoSchedulingEnabled: true,
        autoSchedulingLocked: true,
        autoSchedulingStatus: "scheduled",
      }),
    ).toBe("locked");
  });

  it("keeps actionable Maya states visible while blocks are locked", () => {
    expect(
      deriveAutoSchedulingStatus({
        autoSchedulingEnabled: true,
        autoSchedulingLocked: true,
        autoSchedulingStatus: "at_risk",
      }),
    ).toBe("at_risk");
    expect(
      deriveAutoSchedulingStatus({
        autoSchedulingEnabled: true,
        autoSchedulingLocked: true,
        autoSchedulingStatus: "cannot_fit",
      }),
    ).toBe("cannot_fit");
    expect(
      deriveAutoSchedulingStatus({
        autoSchedulingEnabled: true,
        autoSchedulingLocked: true,
        autoSchedulingStatus: "needs_time",
      }),
    ).toBe("needs_time");
  });

  it("only offers locking for scheduled or already locked work", () => {
    expect(canLockAutoSchedulingStatus("scheduled")).toBe(true);
    expect(canLockAutoSchedulingStatus("locked")).toBe(true);
    expect(canLockAutoSchedulingStatus("at_risk")).toBe(false);
    expect(canLockAutoSchedulingStatus("cannot_fit")).toBe(false);
  });

  it("keeps unlock available from a persisted risk state", () => {
    expect(canToggleAutoSchedulingLock("at_risk", true)).toBe(true);
    expect(canToggleAutoSchedulingLock("cannot_fit", true)).toBe(true);
    expect(canToggleAutoSchedulingLock("at_risk", false)).toBe(false);
  });

  it("surfaces a stored reason instead of generic helper copy", () => {
    expect(
      getAutoSchedulingHelper("at_risk", "A customer call moved this block."),
    ).toBe("A customer call moved this block.");
  });

  it("only treats an explicit Maya assignment as an enable signal", () => {
    expect(isMayaAssigneeSelection("maya-1", "maya-1")).toBe(true);
    expect(isMayaAssigneeSelection("human-1", "maya-1")).toBe(false);
    expect(isMayaAssigneeSelection(null, "maya-1")).toBe(false);
  });

  it("keeps human-assigned creation off unless the user explicitly enabled it", () => {
    expect(
      getNewStoryAutoSchedulingEnabled({
        currentEnabled: true,
        hasExplicitChoice: false,
        mayaAssigneeId: "maya-1",
        selectedAssigneeId: "human-1",
      }),
    ).toBe(false);
    expect(
      getNewStoryAutoSchedulingEnabled({
        currentEnabled: true,
        hasExplicitChoice: true,
        mayaAssigneeId: "maya-1",
        selectedAssigneeId: "human-1",
      }),
    ).toBe(true);
    expect(
      getNewStoryAutoSchedulingEnabled({
        currentEnabled: false,
        hasExplicitChoice: true,
        mayaAssigneeId: "maya-1",
        selectedAssigneeId: "maya-1",
      }),
    ).toBe(true);
  });
});
