type Point = { x: number; y: number };
type Phase = "closed" | "dragging" | "open" | "closing";

const TAP_SLOP = 10;

/**
 * A swipe and its closing animation must not become a detail tap. Keep the
 * current press cancelled even if reduced motion finishes closing before its
 * delayed onPress arrives. Only a new touch can restore that touch intent;
 * accessibility activation is independent and only allowed once fully closed.
 */
export function createSwipeRowInteraction() {
  let phase: Phase = "closed";
  let pressCancelled = false;
  let pressOrigin: Point | undefined;

  return {
    touchStarted(point: Point) {
      pressOrigin = point;
      pressCancelled = phase !== "closed";
    },
    pressMoved(point: Point) {
      if (
        pressOrigin &&
        (Math.abs(point.x - pressOrigin.x) >= TAP_SLOP ||
          Math.abs(point.y - pressOrigin.y) >= TAP_SLOP)
      ) {
        pressCancelled = true;
      }
    },
    cancelPress() {
      pressCancelled = true;
    },
    swipeStarted() {
      phase = "dragging";
      pressCancelled = true;
    },
    willOpen() {
      phase = "open";
      pressCancelled = true;
    },
    willClose() {
      phase = "closing";
      pressCancelled = true;
    },
    requestClose() {
      if (phase === "closed") return false;
      phase = "closing";
      pressCancelled = true;
      return true;
    },
    didClose() {
      phase = "closed";
    },
    canOpenDetails(source: "touch" | "activation" = "touch") {
      return phase === "closed" && (source === "activation" || !pressCancelled);
    },
  };
}
