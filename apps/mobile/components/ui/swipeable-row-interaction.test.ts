import assert from "node:assert/strict";
import test from "node:test";
import { createSwipeRowInteraction } from "./swipeable-row-interaction";
import {
  captureSwipeAction,
  resolveSwipeAction,
} from "./swipeable-row-actions";

const origin = { x: 200, y: 100 };

test("a fresh closed-row tap still opens details", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.pressMoved({ x: 203, y: 102 });
  assert.equal(interaction.canOpenDetails(), true);
});

test("touch movement cancels the press before the native swipe callback arrives", () => {
  for (const destination of [
    { x: 175, y: 100 },
    { x: 200, y: 125 },
  ]) {
    const interaction = createSwipeRowInteraction();
    interaction.touchStarted(origin);
    interaction.pressMoved(destination);
    assert.equal(interaction.canOpenDetails(), false);
    // Dragging back to the starting position must not turn it back into a tap.
    interaction.pressMoved(origin);
    assert.equal(interaction.canOpenDetails(), false);
  }
});

test("opening drag, release and open rail never become a notification press", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.swipeStarted();
  assert.equal(interaction.canOpenDetails(), false);
  interaction.willOpen();
  assert.equal(interaction.canOpenDetails(), false);
});

test("tapping an open row only closes it, including reduced-motion event order", () => {
  const interaction = createSwipeRowInteraction();
  interaction.willOpen();
  interaction.touchStarted(origin);
  interaction.willClose();
  assert.equal(interaction.canOpenDetails(), false);
  // RNGH restores pointer events when the close spring starts. Reduced motion
  // may deliver its completion before RN delivers onPress for the same touch.
  interaction.didClose();
  assert.equal(interaction.canOpenDetails(), false);
  interaction.touchStarted(origin);
  assert.equal(interaction.canOpenDetails(), true);
});

test("a short aborted swipe stays cancelled after the close callback", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.swipeStarted();
  interaction.willClose();
  interaction.didClose();
  assert.equal(interaction.canOpenDetails(), false);
  interaction.touchStarted(origin);
  assert.equal(interaction.canOpenDetails(), true);
});

test("starting a new touch during a closing animation does not navigate", () => {
  const interaction = createSwipeRowInteraction();
  interaction.willOpen();
  interaction.willClose();
  interaction.touchStarted(origin);
  interaction.didClose();
  assert.equal(interaction.canOpenDetails(), false);
});

test("a full swipe reveals actions without executing or opening anything", () => {
  const interaction = createSwipeRowInteraction();
  let actionCalls = 0;
  const action = {
    accessibilityLabel: "Delete notification",
    onPress: () => actionCalls++,
  };
  interaction.touchStarted(origin);
  interaction.pressMoved({ x: -1000, y: 100 });
  interaction.swipeStarted();
  interaction.willOpen();
  assert.equal(interaction.canOpenDetails(), false);
  assert.equal(actionCalls, 0);
  // Actions still require an explicit selection through the existing token gate.
  resolveSwipeAction(
    { actions: [action], active: true },
    captureSwipeAction(action),
  )?.onPress();
  assert.equal(actionCalls, 1);
});

test("an action tap executes only the selected action and cannot open details", () => {
  const interaction = createSwipeRowInteraction();
  let readCalls = 0;
  let deleteCalls = 0;
  const actions = [
    {
      id: "read",
      accessibilityLabel: "Mark as read",
      onPress: () => readCalls++,
    },
    {
      id: "delete",
      accessibilityLabel: "Delete notification",
      onPress: () => deleteCalls++,
    },
  ];
  interaction.willOpen();
  interaction.cancelPress();
  interaction.willClose();
  interaction.didClose();
  resolveSwipeAction(
    { actions, active: true, scope: "session:notification" },
    captureSwipeAction(actions[0], "session:notification"),
  )?.onPress();
  assert.equal(readCalls, 1);
  assert.equal(deleteCalls, 0);
  assert.equal(interaction.canOpenDetails(), false);
});

test("a native action menu cannot release a detail press after dismissal", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.cancelPress();
  interaction.willClose();
  interaction.didClose();
  assert.equal(interaction.canOpenDetails(), false);
  interaction.touchStarted(origin);
  assert.equal(interaction.canOpenDetails(), true);
});

test("a cancelled touch stays cancelled when Pressability reactivates its press rect", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.pressMoved({ x: 160, y: 100 });
  // Pressability may emit onPressIn again here, but only onTouchStart resets us.
  interaction.pressMoved(origin);
  assert.equal(interaction.canOpenDetails(), false);
});

test("closing an already closed row needs no animation completion to recover", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.cancelPress(); // Opening a native long-press menu.
  assert.equal(interaction.requestClose(), false);
  assert.equal(interaction.canOpenDetails(), false);
  // Cancel the menu without any native swipe-close callback.
  interaction.touchStarted(origin);
  assert.equal(interaction.canOpenDetails(), true);
  interaction.pressMoved({ x: 188, y: 100 }); // Below the native swipe threshold.
  assert.equal(interaction.requestClose(), false);
  assert.equal(interaction.canOpenDetails(), false);
  interaction.touchStarted(origin);
  assert.equal(interaction.canOpenDetails(), true);
});

test("accessibility activation recovers after cancelled touch without rearming it", () => {
  const interaction = createSwipeRowInteraction();
  interaction.touchStarted(origin);
  interaction.cancelPress();
  assert.equal(interaction.canOpenDetails("activation"), true);
  assert.equal(interaction.canOpenDetails("touch"), false);
  interaction.swipeStarted();
  assert.equal(interaction.canOpenDetails("activation"), false);
  interaction.willOpen();
  assert.equal(interaction.canOpenDetails("activation"), false);
  interaction.willClose();
  assert.equal(interaction.canOpenDetails("activation"), false);
  interaction.didClose();
  assert.equal(interaction.canOpenDetails("activation"), true);
  assert.equal(interaction.canOpenDetails("touch"), false);
});
