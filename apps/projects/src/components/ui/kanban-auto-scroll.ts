const clamp = (value: number, min: number, max: number) =>
  Math.min(max, Math.max(min, value));

const smoothstep = (progress: number) =>
  progress * progress * (3 - 2 * progress);

export const KANBAN_AUTO_SCROLL_EDGE_SIZE_PX = 96;
export const KANBAN_AUTO_SCROLL_MAX_VELOCITY_PX_PER_SECOND = 480;

export const canAutoScrollKanbanColumn = (element: Element) =>
  element.hasAttribute("data-kanban-column-scroll");

export type KanbanAutoScrollInput = {
  clientWidth: number;
  edgeSize: number;
  maxVelocity: number;
  pointerX: number;
  scrollLeft: number;
  scrollWidth: number;
  viewportLeft: number;
  viewportRight: number;
};

export const getKanbanAutoScrollVelocity = ({
  clientWidth,
  edgeSize,
  maxVelocity,
  pointerX,
  scrollLeft,
  scrollWidth,
  viewportLeft,
  viewportRight,
}: KanbanAutoScrollInput) => {
  const viewportWidth = Math.max(0, viewportRight - viewportLeft);
  const effectiveEdgeSize = Math.min(Math.max(0, edgeSize), viewportWidth / 2);
  const maxScrollLeft = Math.max(0, scrollWidth - clientWidth);
  const speedCap = Math.max(0, maxVelocity);

  if (effectiveEdgeSize === 0 || maxScrollLeft === 0 || speedCap === 0) {
    return 0;
  }

  const leftProgress = clamp(
    (viewportLeft + effectiveEdgeSize - pointerX) / effectiveEdgeSize,
    0,
    1,
  );
  const rightProgress = clamp(
    (pointerX - (viewportRight - effectiveEdgeSize)) / effectiveEdgeSize,
    0,
    1,
  );

  if (leftProgress > 0) {
    return scrollLeft <= 0 ? 0 : -speedCap * smoothstep(leftProgress);
  }

  if (rightProgress > 0) {
    return scrollLeft >= maxScrollLeft
      ? 0
      : speedCap * smoothstep(rightProgress);
  }

  return 0;
};
