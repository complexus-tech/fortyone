"use client";

import type { RefObject } from "react";
import { useCallback, useEffect, useRef } from "react";
import { useDndMonitor } from "@dnd-kit/core";
import {
  getKanbanAutoScrollVelocity,
  KANBAN_AUTO_SCROLL_EDGE_SIZE_PX,
  KANBAN_AUTO_SCROLL_MAX_VELOCITY_PX_PER_SECOND,
} from "./kanban-auto-scroll";

const MAX_FRAME_DURATION_MS = 32;

type ScrollMetrics = {
  clientWidth: number;
  edgeSize: number;
  scrollWidth: number;
  viewportLeft: number;
  viewportRight: number;
};

const getEventClientX = (event: unknown) => {
  const pointerEvent = event as Event & { clientX?: unknown };
  if (typeof pointerEvent.clientX === "number") {
    return pointerEvent.clientX;
  }

  const touchEvent = event as Event & {
    changedTouches?: TouchList;
    touches?: TouchList;
  };
  return (
    touchEvent.touches?.item(0)?.clientX ??
    touchEvent.changedTouches?.item(0)?.clientX ??
    null
  );
};

export const useKanbanAutoScroll = (
  scrollContainerRef: RefObject<HTMLDivElement | null>,
) => {
  const activeRef = useRef(false);
  const frameIdRef = useRef<number | null>(null);
  const lastFrameTimeRef = useRef<number | null>(null);
  const pointerIsDownRef = useRef(false);
  const pointerXRef = useRef<number | null>(null);
  const scrollMetricsRef = useRef<ScrollMetrics | null>(null);

  const stopAutoScroll = useCallback(() => {
    activeRef.current = false;
    pointerIsDownRef.current = false;
    pointerXRef.current = null;
    lastFrameTimeRef.current = null;
    scrollMetricsRef.current = null;

    if (frameIdRef.current !== null) {
      cancelAnimationFrame(frameIdRef.current);
      frameIdRef.current = null;
    }
  }, []);

  const runAutoScrollFrame = useCallback(
    (timestamp: number) => {
      frameIdRef.current = null;
      if (!activeRef.current) return;

      const scrollContainer = scrollContainerRef.current;
      const scrollMetrics = scrollMetricsRef.current;
      const pointerX = pointerXRef.current;
      if (!scrollContainer || !scrollMetrics || pointerX === null) {
        stopAutoScroll();
        return;
      }

      const velocity = getKanbanAutoScrollVelocity({
        ...scrollMetrics,
        maxVelocity: KANBAN_AUTO_SCROLL_MAX_VELOCITY_PX_PER_SECOND,
        pointerX,
        scrollLeft: scrollContainer.scrollLeft,
      });

      if (velocity === 0) {
        lastFrameTimeRef.current = null;
        return;
      }

      const previousTimestamp = lastFrameTimeRef.current ?? timestamp;
      const frameDuration = clampFrameDuration(timestamp - previousTimestamp);
      lastFrameTimeRef.current = timestamp;

      if (frameDuration > 0) {
        const maxScrollLeft = Math.max(
          0,
          scrollMetrics.scrollWidth - scrollMetrics.clientWidth,
        );
        scrollContainer.scrollLeft = Math.min(
          maxScrollLeft,
          Math.max(
            0,
            scrollContainer.scrollLeft + (velocity * frameDuration) / 1_000,
          ),
        );
      }

      frameIdRef.current = requestAnimationFrame(runAutoScrollFrame);
    },
    [scrollContainerRef, stopAutoScroll],
  );

  const scheduleAutoScroll = useCallback(() => {
    if (!activeRef.current || frameIdRef.current !== null) return;

    frameIdRef.current = requestAnimationFrame(runAutoScrollFrame);
  }, [runAutoScrollFrame]);

  const startAutoScroll = useCallback(() => {
    const pointerX = pointerXRef.current;
    const scrollContainer = scrollContainerRef.current;
    if (pointerX === null || !scrollContainer) return;

    stopAutoScroll();
    const rect = scrollContainer.getBoundingClientRect();
    scrollMetricsRef.current = {
      clientWidth: scrollContainer.clientWidth,
      edgeSize: Math.min(
        KANBAN_AUTO_SCROLL_EDGE_SIZE_PX,
        scrollContainer.clientWidth * 0.2,
      ),
      scrollWidth: scrollContainer.scrollWidth,
      viewportLeft: rect.left,
      viewportRight: rect.right,
    };
    activeRef.current = true;
    pointerXRef.current = pointerX;
    scheduleAutoScroll();
  }, [scheduleAutoScroll, scrollContainerRef, stopAutoScroll]);

  useDndMonitor({
    onDragCancel: stopAutoScroll,
    onDragEnd: stopAutoScroll,
    onDragStart: startAutoScroll,
  });

  useEffect(() => {
    const handlePointerDown = (event: PointerEvent) => {
      pointerIsDownRef.current = true;
      pointerXRef.current = event.clientX;
    };
    const handlePointerMove = (event: PointerEvent) => {
      if (activeRef.current || pointerIsDownRef.current) {
        pointerXRef.current = event.clientX;
        scheduleAutoScroll();
      }
    };
    const handlePointerUp = () => {
      pointerIsDownRef.current = false;
      if (!activeRef.current) pointerXRef.current = null;
    };
    const handleTouchStart = (event: TouchEvent) => {
      const pointerX = getEventClientX(event);
      if (pointerX === null) return;

      pointerIsDownRef.current = true;
      pointerXRef.current = pointerX;
      scheduleAutoScroll();
    };
    const handleTouchMove = (event: TouchEvent) => {
      if (!activeRef.current && !pointerIsDownRef.current) return;

      const pointerX = getEventClientX(event);
      if (pointerX !== null) {
        pointerXRef.current = pointerX;
        scheduleAutoScroll();
      }
    };

    window.addEventListener("blur", stopAutoScroll);
    window.addEventListener("pointercancel", handlePointerUp);
    window.addEventListener("pointerdown", handlePointerDown, {
      passive: true,
    });
    window.addEventListener("pointermove", handlePointerMove, {
      passive: true,
    });
    window.addEventListener("pointerup", handlePointerUp);
    window.addEventListener("touchcancel", handlePointerUp);
    window.addEventListener("touchend", handlePointerUp);
    window.addEventListener("touchstart", handleTouchStart, { passive: true });
    window.addEventListener("touchmove", handleTouchMove, { passive: true });

    return () => {
      window.removeEventListener("blur", stopAutoScroll);
      window.removeEventListener("pointercancel", handlePointerUp);
      window.removeEventListener("pointerdown", handlePointerDown);
      window.removeEventListener("pointermove", handlePointerMove);
      window.removeEventListener("pointerup", handlePointerUp);
      window.removeEventListener("touchcancel", handlePointerUp);
      window.removeEventListener("touchend", handlePointerUp);
      window.removeEventListener("touchstart", handleTouchStart);
      window.removeEventListener("touchmove", handleTouchMove);
      stopAutoScroll();
    };
  }, [scheduleAutoScroll, stopAutoScroll]);
};

const clampFrameDuration = (duration: number) =>
  Math.min(MAX_FRAME_DURATION_MS, Math.max(0, duration));
