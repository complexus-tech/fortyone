"use client";

import type { RefObject } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { getVirtualLayout } from "./roadmap-virtualization";

type VirtualAxis = "horizontal" | "vertical";

const getElementSize = (element: HTMLElement, axis: VirtualAxis) =>
  axis === "vertical"
    ? element.getBoundingClientRect().height
    : element.getBoundingClientRect().width;

export const useRoadmapVirtualizer = ({
  axis = "vertical",
  estimatedSize,
  itemKeys,
  overscan,
  pinnedKeys,
  scrollElementRef,
}: {
  axis?: VirtualAxis;
  estimatedSize: number;
  itemKeys: string[];
  overscan?: number;
  pinnedKeys?: readonly string[];
  scrollElementRef: RefObject<HTMLElement | null>;
}) => {
  const measuredElementsRef = useRef(new Map<string, HTMLElement>());
  const resizeObserverRef = useRef<ResizeObserver | null>(null);
  const [measuredSizes, setMeasuredSizes] = useState<
    ReadonlyMap<string, number>
  >(() => new Map());
  const [viewport, setViewport] = useState({
    offset: 0,
    size: estimatedSize * (axis === "vertical" ? 8 : 3),
  });
  const recordMeasuredSize = useCallback((key: string, size: number) => {
    if (size <= 0) return;

    setMeasuredSizes((current) => {
      if (current.get(key) === size) return current;

      const next = new Map(current);
      next.set(key, size);
      return next;
    });
  }, []);

  useEffect(() => {
    const scrollElement = scrollElementRef.current;
    if (!scrollElement) return;

    let frameId: number | undefined;
    const updateViewport = () => {
      if (frameId !== undefined) cancelAnimationFrame(frameId);

      frameId = requestAnimationFrame(() => {
        const nextViewport =
          axis === "vertical"
            ? {
                offset: scrollElement.scrollTop,
                size: scrollElement.clientHeight,
              }
            : {
                offset: scrollElement.scrollLeft,
                size: scrollElement.clientWidth,
              };

        setViewport((current) =>
          current.offset === nextViewport.offset &&
          current.size === nextViewport.size
            ? current
            : nextViewport,
        );
      });
    };

    updateViewport();
    scrollElement.addEventListener("scroll", updateViewport, {
      passive: true,
    });

    const resizeObserver =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(updateViewport);
    resizeObserver?.observe(scrollElement);
    window.addEventListener("resize", updateViewport);

    return () => {
      if (frameId !== undefined) cancelAnimationFrame(frameId);
      resizeObserver?.disconnect();
      scrollElement.removeEventListener("scroll", updateViewport);
      window.removeEventListener("resize", updateViewport);
    };
  }, [axis, scrollElementRef]);

  useEffect(() => {
    if (typeof ResizeObserver === "undefined") return;

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        const element = entry.target as HTMLElement;
        const key = element.dataset.virtualItemKey;
        if (!key) continue;

        const size = getElementSize(element, axis);
        recordMeasuredSize(key, size);
      }
    });

    resizeObserverRef.current = resizeObserver;
    for (const element of measuredElementsRef.current.values()) {
      resizeObserver.observe(element);
    }

    return () => {
      resizeObserver.disconnect();
      resizeObserverRef.current = null;
    };
  }, [axis, recordMeasuredSize]);

  const measureElement = useCallback(
    (key: string, element: HTMLElement | null) => {
      const previousElement = measuredElementsRef.current.get(key);
      if (previousElement && previousElement !== element) {
        resizeObserverRef.current?.unobserve(previousElement);
        measuredElementsRef.current.delete(key);
      }

      if (!element) return;

      element.dataset.virtualItemKey = key;
      measuredElementsRef.current.set(key, element);
      resizeObserverRef.current?.observe(element);

      const size = getElementSize(element, axis);
      recordMeasuredSize(key, size);
    },
    [axis, recordMeasuredSize],
  );

  const layout = useMemo(
    () =>
      getVirtualLayout({
        estimatedSize,
        itemKeys,
        measuredSizes,
        overscan,
        pinnedKeys,
        scrollOffset: viewport.offset,
        viewportSize: viewport.size,
      }),
    [
      estimatedSize,
      itemKeys,
      measuredSizes,
      overscan,
      pinnedKeys,
      viewport.offset,
      viewport.size,
    ],
  );

  return { ...layout, measureElement };
};
