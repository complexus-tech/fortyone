export type VirtualItem = {
  index: number;
  key: string;
  size: number;
  start: number;
};

export type VirtualLayout = {
  items: VirtualItem[];
  totalSize: number;
};

const findFirstItemEndingAfter = (items: VirtualItem[], offset: number) => {
  let lowerBound = 0;
  let upperBound = items.length;

  while (lowerBound < upperBound) {
    const middle = Math.floor((lowerBound + upperBound) / 2);
    const item = items[middle];

    if (item.start + item.size <= offset) {
      lowerBound = middle + 1;
    } else {
      upperBound = middle;
    }
  }

  return lowerBound;
};

const findFirstItemStartingAtOrAfter = (
  items: VirtualItem[],
  offset: number,
) => {
  let lowerBound = 0;
  let upperBound = items.length;

  while (lowerBound < upperBound) {
    const middle = Math.floor((lowerBound + upperBound) / 2);

    if (items[middle].start < offset) {
      lowerBound = middle + 1;
    } else {
      upperBound = middle;
    }
  }

  return lowerBound;
};

export const getVirtualLayout = ({
  itemKeys,
  estimatedSize,
  measuredSizes = new Map<string, number>(),
  scrollOffset,
  viewportSize,
  overscan = 4,
  pinnedKeys = [],
}: {
  itemKeys: string[];
  estimatedSize: number;
  measuredSizes?: ReadonlyMap<string, number>;
  scrollOffset: number;
  viewportSize: number;
  overscan?: number;
  pinnedKeys?: readonly string[];
}): VirtualLayout => {
  if (itemKeys.length === 0) {
    return { items: [], totalSize: 0 };
  }

  const allItems: VirtualItem[] = [];
  let totalSize = 0;

  for (const [index, key] of itemKeys.entries()) {
    const measuredSize = measuredSizes.get(key);
    const size =
      measuredSize !== undefined && measuredSize > 0
        ? measuredSize
        : estimatedSize;

    allItems.push({ index, key, size, start: totalSize });
    totalSize += size;
  }

  const visibleStart = Math.max(0, scrollOffset);
  const visibleEnd = visibleStart + Math.max(1, viewportSize);
  const firstVisibleIndex = Math.min(
    itemKeys.length - 1,
    findFirstItemEndingAfter(allItems, visibleStart),
  );
  const lastVisibleIndex = Math.max(
    firstVisibleIndex,
    findFirstItemStartingAtOrAfter(allItems, visibleEnd) - 1,
  );
  const startIndex = Math.max(0, firstVisibleIndex - overscan);
  const endIndex = Math.min(itemKeys.length, lastVisibleIndex + overscan + 1);
  const pinnedKeySet = new Set(pinnedKeys);

  return {
    items: allItems.filter(
      ({ index, key }) =>
        (index >= startIndex && index < endIndex) || pinnedKeySet.has(key),
    ),
    totalSize,
  };
};
