/* global describe, expect, it -- Jest globals are provided by the projects test runner. */

import { getVirtualLayout } from "./roadmap-virtualization";

const itemKeys = Array.from({ length: 100 }, (_, index) => `item-${index}`);

describe("Roadmap virtualization", () => {
  it("bounds the rendered window around the visible rows", () => {
    const layout = getVirtualLayout({
      estimatedSize: 50,
      itemKeys,
      overscan: 2,
      scrollOffset: 1_000,
      viewportSize: 250,
    });

    expect(layout.totalSize).toBe(5_000);
    expect(layout.items.map(({ index }) => index)).toEqual([
      18, 19, 20, 21, 22, 23, 24, 25, 26,
    ]);
  });

  it("uses measured sizes to preserve the scroll geometry", () => {
    const layout = getVirtualLayout({
      estimatedSize: 50,
      itemKeys: itemKeys.slice(0, 5),
      measuredSizes: new Map([
        ["item-0", 80],
        ["item-1", 30],
      ]),
      overscan: 0,
      scrollOffset: 80,
      viewportSize: 30,
    });

    expect(layout.totalSize).toBe(260);
    expect(layout.items).toEqual([
      { index: 1, key: "item-1", size: 30, start: 80 },
    ]);
  });

  it("keeps a focused or dragged item mounted without mounting the gap", () => {
    const layout = getVirtualLayout({
      estimatedSize: 50,
      itemKeys,
      overscan: 0,
      pinnedKeys: ["item-90"],
      scrollOffset: 0,
      viewportSize: 100,
    });

    expect(layout.items.map(({ index }) => index)).toEqual([0, 1, 90]);
  });

  it("clamps a viewport beyond the end to the final item", () => {
    const layout = getVirtualLayout({
      estimatedSize: 50,
      itemKeys: itemKeys.slice(0, 3),
      overscan: 0,
      scrollOffset: 10_000,
      viewportSize: 500,
    });

    expect(layout.items.map(({ index }) => index)).toEqual([2]);
  });
});
