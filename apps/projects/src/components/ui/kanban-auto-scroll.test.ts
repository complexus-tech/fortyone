/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import {
  canAutoScrollKanbanColumn,
  getKanbanAutoScrollVelocity,
} from "./kanban-auto-scroll";

const baseInput = {
  clientWidth: 800,
  edgeSize: 100,
  maxVelocity: 800,
  scrollLeft: 400,
  scrollWidth: 1_600,
  viewportLeft: 100,
  viewportRight: 900,
};

describe("getKanbanAutoScrollVelocity", () => {
  it.each([
    { pointerX: 100, expected: -800 },
    { pointerX: 900, expected: 800 },
    { pointerX: 150, expected: -400 },
    { pointerX: 850, expected: 400 },
    { pointerX: 175, expected: -125 },
    { pointerX: 825, expected: 125 },
  ])(
    "returns a symmetric capped velocity at pointer position $pointerX",
    ({ expected, pointerX }) => {
      expect(
        getKanbanAutoScrollVelocity({ ...baseInput, pointerX }),
      ).toBeCloseTo(expected);
    },
  );

  it.each([200, 500, 800])(
    "returns zero inside the dead zone at pointer position %s",
    (pointerX) => {
      expect(getKanbanAutoScrollVelocity({ ...baseInput, pointerX })).toBe(0);
    },
  );

  it("caps velocity when the pointer moves beyond the viewport", () => {
    expect(getKanbanAutoScrollVelocity({ ...baseInput, pointerX: -100 })).toBe(
      -800,
    );
    expect(getKanbanAutoScrollVelocity({ ...baseInput, pointerX: 1_100 })).toBe(
      800,
    );
  });

  it("stops at each bound while keeping the opposite direction available", () => {
    expect(
      getKanbanAutoScrollVelocity({
        ...baseInput,
        pointerX: 100,
        scrollLeft: 0,
      }),
    ).toBe(0);
    expect(
      getKanbanAutoScrollVelocity({
        ...baseInput,
        pointerX: 900,
        scrollLeft: 0,
      }),
    ).toBe(800);
    expect(
      getKanbanAutoScrollVelocity({
        ...baseInput,
        pointerX: 900,
        scrollLeft: 800,
      }),
    ).toBe(0);
    expect(
      getKanbanAutoScrollVelocity({
        ...baseInput,
        pointerX: 100,
        scrollLeft: 800,
      }),
    ).toBe(-800);
  });

  it("does not scroll when the container has no horizontal overflow", () => {
    expect(
      getKanbanAutoScrollVelocity({
        ...baseInput,
        pointerX: 900,
        scrollWidth: baseInput.clientWidth,
      }),
    ).toBe(0);
  });
});

describe("canAutoScrollKanbanColumn", () => {
  it("keeps dnd-kit auto-scroll scoped to vertical columns", () => {
    const column = document.createElement("div");
    const board = document.createElement("div");
    column.setAttribute("data-kanban-column-scroll", "");
    board.setAttribute("data-kanban-board-scroll", "");

    expect(canAutoScrollKanbanColumn(column)).toBe(true);
    expect(canAutoScrollKanbanColumn(board)).toBe(false);
    expect(canAutoScrollKanbanColumn(document.documentElement)).toBe(false);
  });
});
