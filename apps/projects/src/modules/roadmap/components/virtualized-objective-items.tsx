"use client";

import type { CSSProperties, ReactNode, RefObject } from "react";
import { useMemo } from "react";
import { Box } from "ui";
import { cn } from "lib";
import { useRoadmapVirtualizer } from "../use-roadmap-virtualizer";

type VirtualizedObjectiveItemsProps<T> = {
  axis?: "horizontal" | "vertical";
  className?: string;
  estimatedSize: number;
  getItemKey: (item: T) => string;
  itemClassName?: string;
  items: T[];
  onItemFocus?: (item: T) => void;
  overscan?: number;
  pinnedKeys?: readonly string[];
  renderItem: (item: T, index: number) => ReactNode;
  scrollElementRef: RefObject<HTMLElement | null>;
  style?: CSSProperties;
};

export const VirtualizedObjectiveItems = <T,>({
  axis = "vertical",
  className,
  estimatedSize,
  getItemKey,
  itemClassName,
  items,
  onItemFocus,
  overscan,
  pinnedKeys,
  renderItem,
  scrollElementRef,
  style,
}: VirtualizedObjectiveItemsProps<T>) => {
  const itemKeys = useMemo(
    () => items.map((item) => getItemKey(item)),
    [getItemKey, items],
  );
  const {
    items: virtualItems,
    measureElement,
    totalSize,
  } = useRoadmapVirtualizer({
    axis,
    estimatedSize,
    itemKeys,
    overscan,
    pinnedKeys,
    scrollElementRef,
  });
  const isVertical = axis === "vertical";

  return (
    <Box
      aria-label="Objectives"
      className={cn("relative", className)}
      role="list"
      style={{
        ...style,
        ...(isVertical ? { height: totalSize } : { width: totalSize }),
      }}
    >
      {virtualItems.map((virtualItem) => {
        const item = items[virtualItem.index];

        return (
          <Box
            aria-posinset={virtualItem.index + 1}
            aria-setsize={items.length}
            className={cn(
              "absolute",
              isVertical ? "inset-x-0 top-0" : "inset-y-0 left-0",
              itemClassName,
            )}
            data-virtual-index={virtualItem.index}
            key={virtualItem.key}
            onFocusCapture={() => {
              onItemFocus?.(item);
            }}
            ref={(element) => {
              measureElement(virtualItem.key, element);
            }}
            role="listitem"
            style={{
              transform: isVertical
                ? `translateY(${virtualItem.start}px)`
                : `translateX(${virtualItem.start}px)`,
              ...(isVertical ? undefined : { width: virtualItem.size }),
            }}
          >
            {renderItem(item, virtualItem.index)}
          </Box>
        );
      })}
    </Box>
  );
};
