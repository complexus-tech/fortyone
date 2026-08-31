export type WalkthroughPosition =
  | "top"
  | "top-start"
  | "top-end"
  | "bottom"
  | "bottom-start"
  | "bottom-end"
  | "center"
  | "left"
  | "right";

export interface WalkthroughTargetPosition {
  top: number;
  left: number;
  width: number;
  height: number;
}

export interface WalkthroughPanelSize {
  width: number;
  height: number;
}

interface ViewportSize {
  width: number;
  height: number;
}

interface PanelCoordinates {
  top: number;
  left: number;
}

const VIEWPORT_PADDING = 16;
const WALKTHROUGH_GAP = 16;

const placementCandidates: Record<
  WalkthroughPosition,
  readonly WalkthroughPosition[]
> = {
  top: ["top", "bottom", "top-start", "top-end", "bottom-start", "bottom-end"],
  "top-start": [
    "top-start",
    "top-end",
    "bottom-start",
    "bottom-end",
    "top",
    "bottom",
  ],
  "top-end": [
    "top-end",
    "top-start",
    "bottom-end",
    "bottom-start",
    "top",
    "bottom",
  ],
  bottom: [
    "bottom",
    "top",
    "bottom-start",
    "bottom-end",
    "top-start",
    "top-end",
  ],
  "bottom-start": [
    "bottom-start",
    "bottom-end",
    "top-start",
    "top-end",
    "bottom",
    "top",
  ],
  "bottom-end": [
    "bottom-end",
    "bottom-start",
    "top-end",
    "top-start",
    "bottom",
    "top",
  ],
  center: ["center"],
  left: ["left", "right", "top", "bottom"],
  right: ["right", "left", "top", "bottom"],
};

const getPanelCoordinates = (
  targetPosition: WalkthroughTargetPosition,
  panelSize: WalkthroughPanelSize,
  viewport: ViewportSize,
  position: WalkthroughPosition,
): PanelCoordinates => {
  const centeredLeft =
    targetPosition.left + targetPosition.width / 2 - panelSize.width / 2;
  const centeredTop =
    targetPosition.top + targetPosition.height / 2 - panelSize.height / 2;

  switch (position) {
    case "top":
      return {
        top: targetPosition.top - WALKTHROUGH_GAP - panelSize.height,
        left: centeredLeft,
      };
    case "top-start":
      return {
        top: targetPosition.top - WALKTHROUGH_GAP - panelSize.height,
        left: targetPosition.left,
      };
    case "top-end":
      return {
        top: targetPosition.top - WALKTHROUGH_GAP - panelSize.height,
        left: targetPosition.left + targetPosition.width - panelSize.width,
      };
    case "bottom":
      return {
        top: targetPosition.top + targetPosition.height + WALKTHROUGH_GAP,
        left: centeredLeft,
      };
    case "bottom-start":
      return {
        top: targetPosition.top + targetPosition.height + WALKTHROUGH_GAP,
        left: targetPosition.left,
      };
    case "bottom-end":
      return {
        top: targetPosition.top + targetPosition.height + WALKTHROUGH_GAP,
        left: targetPosition.left + targetPosition.width - panelSize.width,
      };
    case "left":
      return {
        top: centeredTop,
        left: targetPosition.left - WALKTHROUGH_GAP - panelSize.width,
      };
    case "right":
      return {
        top: centeredTop,
        left: targetPosition.left + targetPosition.width + WALKTHROUGH_GAP,
      };
    case "center":
      return {
        top: (viewport.height - panelSize.height) / 2,
        left: (viewport.width - panelSize.width) / 2,
      };
  }
};

const getOverflow = (
  coordinates: PanelCoordinates,
  panelSize: WalkthroughPanelSize,
  viewport: ViewportSize,
) =>
  Math.max(0, VIEWPORT_PADDING - coordinates.top) +
  Math.max(0, VIEWPORT_PADDING - coordinates.left) +
  Math.max(
    0,
    coordinates.top + panelSize.height - (viewport.height - VIEWPORT_PADDING),
  ) +
  Math.max(
    0,
    coordinates.left + panelSize.width - (viewport.width - VIEWPORT_PADDING),
  );

const clamp = (value: number, minimum: number, maximum: number) =>
  Math.min(Math.max(value, minimum), maximum);

const clampToViewport = (
  coordinates: PanelCoordinates,
  panelSize: WalkthroughPanelSize,
  viewport: ViewportSize,
): PanelCoordinates => ({
  top: clamp(
    coordinates.top,
    VIEWPORT_PADDING,
    Math.max(
      VIEWPORT_PADDING,
      viewport.height - panelSize.height - VIEWPORT_PADDING,
    ),
  ),
  left: clamp(
    coordinates.left,
    VIEWPORT_PADDING,
    Math.max(
      VIEWPORT_PADDING,
      viewport.width - panelSize.width - VIEWPORT_PADDING,
    ),
  ),
});

export const getWalkthroughPanelPosition = ({
  panelSize,
  position = "bottom",
  targetPosition,
  viewport,
}: {
  panelSize: WalkthroughPanelSize;
  position?: WalkthroughPosition;
  targetPosition: WalkthroughTargetPosition;
  viewport: ViewportSize;
}): PanelCoordinates => {
  const candidates = placementCandidates[position];
  let bestCoordinates = getPanelCoordinates(
    targetPosition,
    panelSize,
    viewport,
    candidates[0],
  );
  let lowestOverflow = getOverflow(bestCoordinates, panelSize, viewport);

  for (let index = 1; index < candidates.length; index += 1) {
    const candidate = candidates[index];
    const coordinates = getPanelCoordinates(
      targetPosition,
      panelSize,
      viewport,
      candidate,
    );
    const overflow = getOverflow(coordinates, panelSize, viewport);

    if (overflow < lowestOverflow) {
      bestCoordinates = coordinates;
      lowestOverflow = overflow;

      if (lowestOverflow === 0) {
        break;
      }
    }
  }

  return clampToViewport(bestCoordinates, panelSize, viewport);
};
