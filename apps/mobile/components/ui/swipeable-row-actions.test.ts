import assert from "node:assert/strict";
import test from "node:test";
import {
  captureSwipeAction,
  resolveSwipeAction,
} from "./swipeable-row-actions";

const read = { id: "read", accessibilityLabel: "Mark as read" };
const unread = { id: "unread", accessibilityLabel: "Mark as unread" };
const remove = { id: "delete", accessibilityLabel: "Delete notification" };

test("a native menu choice targets its identity, not its position after actions reorder", () => {
  const selected = captureSwipeAction(remove, "session:notification");
  const current = {
    actions: [remove, read],
    active: true,
    scope: selected.scope,
  };
  assert.equal(resolveSwipeAction(current, selected), remove);
});

test("an old read choice cannot become unread when a notification updates", () => {
  const selected = captureSwipeAction(read, "session:notification");
  assert.equal(
    resolveSwipeAction(
      { actions: [unread, remove], active: true, scope: selected.scope },
      selected,
    ),
    undefined,
  );
});

test("menu choices are rejected after recycling, session changes, unmount or a pending action", () => {
  const selected = captureSwipeAction(
    remove,
    "first-session:first-notification",
  );
  const current = {
    actions: [read, remove],
    active: true,
    scope: selected.scope,
  };
  for (const changed of [
    { ...current, scope: "second-session:first-notification" },
    { ...current, scope: "first-session:second-notification" },
    { ...current, active: false },
    { ...current, disabled: true },
    { ...current, actions: [read, { ...remove, disabled: true }] },
    { ...current, actions: [read] },
  ])
    assert.equal(resolveSwipeAction(changed, selected), undefined);
});

test("the existing single Complete action remains supported without an explicit id", () => {
  const complete = { accessibilityLabel: "Mark as done" };
  const selected = captureSwipeAction(complete);
  assert.equal(
    resolveSwipeAction({ actions: [complete], active: true }, selected),
    complete,
  );
  assert.equal(
    resolveSwipeAction(
      { actions: [{ accessibilityLabel: "Mark as closed" }], active: true },
      selected,
    ),
    undefined,
  );
});
