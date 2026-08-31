/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getWalkthroughPanelPosition } from "./walkthrough-position";

describe("getWalkthroughPanelPosition", () => {
  it("mirrors a bottom-start panel before it overflows the right viewport edge", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 400, width: 432 },
      position: "bottom-start",
      targetPosition: { height: 96, left: 560, top: 4, width: 100 },
      viewport: { height: 920, width: 860 },
    });

    expect(position).toEqual({ left: 228, top: 116 });
    expect(position.left).toBeGreaterThanOrEqual(16);
    expect(position.left + 432).toBeLessThanOrEqual(860 - 16);
  });

  it("moves a top-aligned panel below its target when there is no room above it", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 240, width: 320 },
      position: "top-start",
      targetPosition: { height: 48, left: 100, top: 24, width: 120 },
      viewport: { height: 700, width: 1000 },
    });

    expect(position).toEqual({ left: 100, top: 88 });
    expect(position.top).toBeGreaterThanOrEqual(16);
  });

  it("keeps a far-right create action panel inside the viewport", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 360, width: 432 },
      position: "bottom-start",
      targetPosition: { height: 60, left: 744, top: 20, width: 60 },
      viewport: { height: 900, width: 860 },
    });

    expect(position).toEqual({ left: 372, top: 96 });
    expect(position.left + 432).toBeLessThanOrEqual(860 - 16);
  });

  it("preserves a right placement when the preferred side has room", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 240, width: 320 },
      position: "right",
      targetPosition: { height: 48, left: 80, top: 300, width: 180 },
      viewport: { height: 900, width: 1200 },
    });

    expect(position).toEqual({ left: 276, top: 204 });
  });

  it("centers a variable-height fallback panel in the viewport", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 476, width: 432 },
      position: "center",
      targetPosition: { height: 0, left: 0, top: 0, width: 0 },
      viewport: { height: 920, width: 1440 },
    });

    expect(position).toEqual({ left: 504, top: 222 });
  });

  it("pins an oversized panel to the safe leading inset", () => {
    const position = getWalkthroughPanelPosition({
      panelSize: { height: 700, width: 900 },
      position: "bottom-start",
      targetPosition: { height: 48, left: 980, top: 680, width: 80 },
      viewport: { height: 720, width: 1100 },
    });

    expect(position).toEqual({ left: 160, top: 16 });
  });
});
