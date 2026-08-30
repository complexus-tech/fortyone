"use client";

import { useMemo } from "react";
import { getGanttVirtualRows } from "./base-gantt-utils";

export type GanttRowWindow = {
  itemCount: number;
  rows: {
    index: number;
    size: number;
    start: number;
  }[];
  totalSize: number;
  virtualized: boolean;
};

export type RenderedGanttRow<T> = GanttRowWindow["rows"][number] & {
  item: T;
};

export const useGanttRowWindow = <T extends { id: string }>({
  activeInteractionId,
  focusedItemId,
  items,
  pinnedItemIds,
  rowHeight,
  scrollTop,
  viewportHeight,
  virtualizeRows,
}: {
  activeInteractionId: string | null;
  focusedItemId: string | null;
  items: T[];
  pinnedItemIds: readonly string[];
  rowHeight: number | string;
  scrollTop: number;
  viewportHeight: number;
  virtualizeRows: boolean;
}) => {
  const canVirtualizeRows =
    virtualizeRows && typeof rowHeight === "number" && rowHeight > 0;
  const pinnedRowIndices = useMemo(() => {
    const pinnedIds = new Set(
      [activeInteractionId, focusedItemId, ...pinnedItemIds].filter(
        (itemId): itemId is string => Boolean(itemId),
      ),
    );

    return items.flatMap((item, index) =>
      pinnedIds.has(item.id) ? [index] : [],
    );
  }, [activeInteractionId, focusedItemId, items, pinnedItemIds]);
  const virtualLayout = useMemo(
    () =>
      canVirtualizeRows
        ? getGanttVirtualRows({
            itemCount: items.length,
            pinnedIndices: pinnedRowIndices,
            rowHeight,
            scrollTop,
            viewportHeight,
          })
        : {
            rows: items.map((_, index) => ({
              index,
              size: typeof rowHeight === "number" ? rowHeight : 0,
              start: 0,
            })),
            totalSize:
              typeof rowHeight === "number" ? items.length * rowHeight : 0,
          },
    [
      canVirtualizeRows,
      items,
      pinnedRowIndices,
      rowHeight,
      scrollTop,
      viewportHeight,
    ],
  );
  const rowWindow = useMemo<GanttRowWindow>(
    () => ({
      itemCount: items.length,
      rows: virtualLayout.rows,
      totalSize: virtualLayout.totalSize,
      virtualized: canVirtualizeRows,
    }),
    [canVirtualizeRows, items.length, virtualLayout],
  );
  const renderedRows = useMemo(
    () => rowWindow.rows.map((row) => ({ ...row, item: items[row.index] })),
    [items, rowWindow.rows],
  );
  const renderedItems = useMemo(
    () => renderedRows.map(({ item }) => item),
    [renderedRows],
  );

  return { renderedItems, renderedRows, rowWindow };
};
