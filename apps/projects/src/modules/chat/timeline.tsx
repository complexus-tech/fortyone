"use client";

import type { ReactNode } from "react";
import type {
  IndexLocationWithAlign,
  VirtuosoHandle,
  VirtuosoProps,
} from "react-virtuoso";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { GroupedVirtuoso, Virtuoso } from "react-virtuoso";
import { MessageDateDivider } from "./message-date-divider";
import { useTimelinePagination } from "./use-timeline-pagination";

const TimelineFooter = () => <div className="h-5" />;
const TIMELINE_COMPONENTS = { Footer: TimelineFooter };
const TimelineHeader = ({ context }: { context?: ReactNode }) => (
  <div>{context}</div>
);
const GROUPED_TIMELINE_COMPONENTS = {
  ...TIMELINE_COMPONENTS,
  Header: TimelineHeader,
};

export const Timeline = <Item extends { id: string }>({
  items,
  renderItem,
  header,
  startAtEnd = false,
  targetItemId,
  scrollToItemId,
  followOwnItemId,
  followNewItems = true,
  isActive = true,
  label = "Message history",
  onVisibleItem,
  getItemDate,
  hasOlder = false,
  isFetching = false,
  paginationError,
  onLoadOlder,
}: {
  items: readonly Item[];
  renderItem: (item: Item, index: number) => ReactNode;
  header: ReactNode;
  startAtEnd?: boolean;
  targetItemId?: string | null;
  scrollToItemId?: string | null;
  followOwnItemId?: string;
  followNewItems?: boolean;
  isActive?: boolean;
  label?: string;
  onVisibleItem?: (item: Item) => void;
  getItemDate?: (item: Item) => string;
  hasOlder?: boolean;
  isFetching?: boolean;
  paginationError?: Error | null;
  onLoadOlder?: () => Promise<unknown>;
}) => {
  const [positionReady, setPositionReady] = useState(
    !startAtEnd && !targetItemId,
  );
  const [windowState, setWindowState] = useState({
    items,
    firstIndex: 1_000_000,
  });
  let firstIndex = windowState.firstIndex;
  if (windowState.items !== items) {
    const oldFirst = windowState.items[0]?.id;
    const newFirst = items[0]?.id;
    if (oldFirst && newFirst && oldFirst !== newFirst) {
      const prepended = items.findIndex((item) => item.id === oldFirst);
      const removed = windowState.items.findIndex(
        (item) => item.id === newFirst,
      );
      firstIndex += Math.max(removed, 0) - Math.max(prepended, 0);
    }
    setWindowState({ items, firstIndex });
  }
  const listRef = useRef<VirtuosoHandle>(null);
  const scrollerRef = useRef<HTMLElement | null>(null);
  const isNearTop = useCallback(() => {
    const scroller = scrollerRef.current;
    return Boolean(scroller?.clientHeight && scroller.scrollTop <= 200);
  }, []);
  const { onNearTopChange } = useTimelinePagination({
    enabled: isActive,
    hasOlder,
    isFetching,
    error: paginationError,
    firstItemId: items[0]?.id,
    onLoadOlder,
    positionReady,
    isNearTop,
  });
  const visibilityFrame = useRef<number | null>(null);
  const focusObserver = useRef<MutationObserver | null>(null);
  const focusedTargetId = useRef<string | null>(null);
  const scrolledTargetId = useRef<string | null>(null);
  const followedItemId = useRef<string | null>(null);
  const previousWindow = useRef({
    firstId: items[0]?.id,
    lastId: items.at(-1)?.id,
  });
  // The intro is a virtual item too, so long threads never keep a heavy root mounted.
  const headerOffset = getItemDate ? 0 : 1;
  const data = useMemo(
    () => (getItemDate ? items : [null, ...items]),
    [getItemDate, items],
  );
  const groups = useMemo(() => {
    const dates: string[] = [];
    const counts: number[] = [];
    const keys: string[] = [];
    if (getItemDate) {
      for (const item of items) {
        const date = getItemDate(item);
        if (date !== dates.at(-1)) {
          dates.push(date);
          counts.push(0);
          keys.push(`date:${date}`);
        }
        counts[counts.length - 1]++;
        keys.push(item.id);
      }
    }
    return { dates, counts, keys };
  }, [getItemDate, items]);
  const renderDateGroup = useCallback(
    (groupIndex: number) => (
      <MessageDateDivider date={groups.dates[groupIndex]} />
    ),
    [groups.dates],
  );
  const targetIndex = useMemo(
    () =>
      targetItemId ? items.findIndex((item) => item.id === targetItemId) : -1,
    [items, targetItemId],
  );
  const inspectVisibleItems = useCallback(() => {
    const scroller = scrollerRef.current;
    if (!scroller || !isActive || !scroller.clientHeight) return;
    // Initial inverse scrolling briefly measures at the top. Only enable paging
    // after the latest messages have actually reached their starting position.
    if (
      !positionReady &&
      startAtEnd &&
      !targetItemId &&
      scroller.querySelector("[data-visible-index]") &&
      scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight <= 24
    ) {
      setPositionReady(true);
    }
    if (!onVisibleItem) return;
    const viewport = scroller.getBoundingClientRect();
    let visible: Item | undefined;
    for (const element of scroller.querySelectorAll<HTMLElement>(
      "[data-visible-index]",
    )) {
      if (getComputedStyle(element).visibility === "hidden") continue;
      const rect = element.getBoundingClientRect();
      const visibleHeight =
        Math.min(rect.bottom, viewport.bottom) -
        Math.max(rect.top, viewport.top);
      if (visibleHeight >= Math.min(rect.height, viewport.height) * 0.6) {
        visible = items[Number(element.dataset.visibleIndex)];
      }
    }
    if (visible) onVisibleItem(visible);
  }, [isActive, items, onVisibleItem, positionReady, startAtEnd, targetItemId]);
  const queueVisibilityCheck = useCallback(() => {
    if (visibilityFrame.current !== null)
      cancelAnimationFrame(visibilityFrame.current);
    visibilityFrame.current = requestAnimationFrame(inspectVisibleItems);
  }, [inspectVisibleItems]);
  useEffect(() => {
    window.addEventListener("focus", queueVisibilityCheck);
    document.addEventListener("visibilitychange", queueVisibilityCheck);
    const observer = new ResizeObserver(queueVisibilityCheck);
    if (scrollerRef.current) observer.observe(scrollerRef.current);
    queueVisibilityCheck();
    return () => {
      observer.disconnect();
      window.removeEventListener("focus", queueVisibilityCheck);
      document.removeEventListener("visibilitychange", queueVisibilityCheck);
      if (visibilityFrame.current !== null)
        cancelAnimationFrame(visibilityFrame.current);
    };
  }, [queueVisibilityCheck]);

  useEffect(() => {
    if (!targetItemId || !isActive) {
      focusedTargetId.current = null;
      scrolledTargetId.current = null;
      return;
    }
    if (targetIndex < 0 || scrolledTargetId.current === targetItemId) return;
    listRef.current?.scrollToIndex({
      index: targetIndex + headerOffset,
      align: "center",
    });
    scrolledTargetId.current = targetItemId;
  }, [targetItemId, targetIndex, isActive, headerOffset]);
  useEffect(() => {
    if (
      !scrollToItemId ||
      !isActive ||
      followedItemId.current === scrollToItemId
    )
      return;
    const index = items.findIndex((item) => item.id === scrollToItemId);
    if (index < 0) return;
    listRef.current?.scrollToIndex({
      index: index + headerOffset,
      align: "end",
    });
    followedItemId.current = scrollToItemId;
  }, [isActive, items, scrollToItemId, headerOffset]);
  useEffect(
    () => () => {
      focusObserver.current?.disconnect();
    },
    [],
  );

  const focusTarget = useCallback(
    (element: HTMLDivElement | null) => {
      focusObserver.current?.disconnect();
      if (
        !element ||
        !isActive ||
        !targetItemId ||
        focusedTargetId.current === targetItemId
      )
        return;
      const focusWhenVisible = () => {
        if (
          !element.getClientRects().length ||
          getComputedStyle(element).visibility === "hidden"
        )
          return false;
        element.focus({ preventScroll: true });
        if (document.activeElement !== element) return false;
        focusedTargetId.current = targetItemId;
        focusObserver.current?.disconnect();
        return true;
      };
      if (focusWhenVisible()) return;
      // Virtuoso hides initial items until measurement finishes. Focus only after
      // that layout becomes visible, rather than guessing how many frames it takes.
      const scroller = element.closest("[data-virtuoso-scroller]");
      if (!scroller) return;
      focusObserver.current = new MutationObserver(focusWhenVisible);
      focusObserver.current.observe(scroller, {
        attributes: true,
        attributeFilter: ["style", "class", "hidden"],
        subtree: true,
      });
    },
    [targetItemId, isActive],
  );
  const followOutput = useCallback(
    (isAtBottom: boolean) => {
      const previous = previousWindow.current;
      const firstId = items[0]?.id;
      const lastId = items.at(-1)?.id;
      previousWindow.current = { firstId, lastId };
      // Prepending history must never trigger follow-output, including while
      // Virtuoso is recalculating its bottom state and restoring the anchor.
      if (firstId !== previous.firstId || lastId === previous.lastId)
        return false;
      return isAtBottom || lastId === followOwnItemId ? "auto" : false;
    },
    [followOwnItemId, items],
  );
  const renderTimelineItem = (index: number, item: Item | null) => (
    <div
      className="flow-root outline-none"
      data-visible-index={item ? index - firstIndex - headerOffset : undefined}
      ref={item?.id === targetItemId ? focusTarget : undefined}
      tabIndex={item?.id === targetItemId ? -1 : undefined}
    >
      {item ? renderItem(item, index - firstIndex - headerOffset) : header}
    </div>
  );
  let initialItem: IndexLocationWithAlign | number = 0;
  if (startAtEnd) initialItem = { index: "LAST", align: "end" };
  if (targetIndex >= 0)
    initialItem = { index: targetIndex + headerOffset, align: "center" };

  const listProps: VirtuosoProps<Item | null, ReactNode> = {
    "aria-label": label,
    atBottomThreshold: 24,
    atTopThreshold: 200,
    atTopStateChange: onNearTopChange,
    className:
      "focus-visible:ring-primary/40 min-h-0 flex-1 overscroll-contain outline-none focus-visible:ring-2 focus-visible:ring-inset",
    components: TIMELINE_COMPONENTS,
    computeItemKey: (_index, item) => item?.id ?? "timeline-header",
    data,
    defaultItemHeight: 96,
    firstItemIndex: firstIndex,
    followOutput: followNewItems ? followOutput : false,
    increaseViewportBy: 300,
    initialTopMostItemIndex: initialItem,
    itemsRendered: queueVisibilityCheck,
    onWheelCapture: () => {
      setPositionReady(true);
    },
    onTouchMove: () => {
      setPositionReady(true);
    },
    onKeyDown: (event) => {
      setPositionReady(true);
      if (event.target !== event.currentTarget) return;
      if (event.key !== "Home" && event.key !== "End") return;
      event.preventDefault();
      // Pixel scrolling has no index retry that could target the new first item
      // after an older page arrives and undo Virtuoso's prepend compensation.
      listRef.current?.scrollTo({
        top: event.key === "Home" ? 0 : scrollerRef.current?.scrollHeight ?? 0,
        behavior: "auto",
      });
    },
    onScroll: queueVisibilityCheck,
    role: "region",
    scrollerRef: (element) => {
      scrollerRef.current = element instanceof HTMLElement ? element : null;
    },
    tabIndex: 0,
  };
  // Initial positioning is a mount-time operation in Virtuoso. Mounting an empty
  // inverse list can finish that operation before the first history page arrives.
  // Keep the intro usable while loading, then initialize with actual messages.
  if (items.length === 0) {
    return (
      <div
        aria-label={label}
        className={`${listProps.className} overflow-y-auto`}
        role="region"
        // eslint-disable-next-line jsx-a11y/no-noninteractive-tabindex -- The scrollable history region must support keyboard scrolling, including a long thread root.
        tabIndex={0}
      >
        {header}
        <TimelineFooter />
      </div>
    );
  }

  return getItemDate ? (
    <GroupedVirtuoso
      {...listProps}
      components={GROUPED_TIMELINE_COMPONENTS}
      computeItemKey={(index) => groups.keys[index - firstIndex] ?? index}
      context={header}
      data={undefined}
      groupContent={renderDateGroup}
      groupCounts={groups.counts}
      itemContent={(index) =>
        renderTimelineItem(index, data[index - firstIndex])
      }
      ref={listRef}
    />
  ) : (
    <Virtuoso {...listProps} itemContent={renderTimelineItem} ref={listRef} />
  );
};
