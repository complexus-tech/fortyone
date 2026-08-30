"use client";

import type { MutableRefObject, RefObject } from "react";
import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import type { StrategyNodeDimensions } from "./strategy-map-layout";
import {
  createStrategyMapViewport,
  getInitialStrategyMapViewport,
  type StrategyMapViewport,
} from "./strategy-map-visibility";

type ZoomAnchor = {
  viewportX: number;
  viewportY: number;
  worldX: number;
  worldY: number;
};

type UseStrategyMapCanvasViewportOptions = {
  layoutWidth: number;
  onZoomChange: (zoom: number) => void;
  resetSignal: number;
  viewportRef: RefObject<HTMLDivElement | null>;
  zoom: number;
};

const areViewportsEqual = (
  current: StrategyMapViewport,
  next: StrategyMapViewport,
) =>
  current.height === next.height &&
  current.left === next.left &&
  current.top === next.top &&
  current.width === next.width;

export const useStrategyMapCanvasViewport = ({
  layoutWidth,
  onZoomChange,
  resetSignal,
  viewportRef,
  zoom,
}: UseStrategyMapCanvasViewportOptions) => {
  const hasPositionedViewportRef = useRef(false);
  const previousResetSignalRef = useRef(resetSignal);
  const previousZoomRef = useRef(zoom);
  const pendingZoomAnchorRef = useRef<ZoomAnchor | null>(null);
  const [viewport, setViewport] = useState<StrategyMapViewport>(() =>
    getInitialStrategyMapViewport(layoutWidth, zoom),
  );
  const updateViewport = useCallback(() => {
    const element = viewportRef.current;
    const nextViewport = element
      ? createStrategyMapViewport({
          clientHeight: element.clientHeight,
          clientWidth: element.clientWidth,
          scrollLeft: element.scrollLeft,
          scrollTop: element.scrollTop,
          zoom,
        })
      : getInitialStrategyMapViewport(layoutWidth, zoom);

    setViewport((current) =>
      areViewportsEqual(current, nextViewport) ? current : nextViewport,
    );
  }, [layoutWidth, viewportRef, zoom]);

  useLayoutEffect(() => {
    const element = viewportRef.current;
    if (!element) return;

    let frameId: number | null = null;
    const scheduleViewportUpdate = () => {
      if (frameId !== null) return;
      frameId = requestAnimationFrame(() => {
        frameId = null;
        updateViewport();
      });
    };
    const observer = new ResizeObserver(scheduleViewportUpdate);

    observer.observe(element);
    element.addEventListener("scroll", scheduleViewportUpdate, {
      passive: true,
    });
    updateViewport();

    return () => {
      if (frameId !== null) cancelAnimationFrame(frameId);
      observer.disconnect();
      element.removeEventListener("scroll", scheduleViewportUpdate);
    };
  }, [updateViewport, viewportRef]);

  useLayoutEffect(() => {
    const viewportElement = viewportRef.current;
    if (!viewportElement) return;

    const shouldReset =
      !hasPositionedViewportRef.current ||
      previousResetSignalRef.current !== resetSignal;
    previousResetSignalRef.current = resetSignal;
    if (!shouldReset) return;

    viewportElement.scrollLeft = Math.max(
      0,
      (layoutWidth * zoom - viewportElement.clientWidth) / 2,
    );
    viewportElement.scrollTop = 0;
    hasPositionedViewportRef.current = true;
    updateViewport();
  }, [layoutWidth, resetSignal, updateViewport, viewportRef, zoom]);

  useLayoutEffect(() => {
    const viewportElement = viewportRef.current;
    const previousZoom = previousZoomRef.current;
    previousZoomRef.current = zoom;
    if (
      !viewportElement ||
      previousZoom === zoom ||
      !hasPositionedViewportRef.current
    ) {
      return;
    }

    const anchor = pendingZoomAnchorRef.current;
    pendingZoomAnchorRef.current = null;
    if (anchor) {
      viewportElement.scrollLeft = anchor.worldX * zoom - anchor.viewportX;
      viewportElement.scrollTop = anchor.worldY * zoom - anchor.viewportY;
      updateViewport();
      return;
    }

    const ratio = zoom / previousZoom;
    viewportElement.scrollLeft =
      (viewportElement.scrollLeft + viewportElement.clientWidth / 2) * ratio -
      viewportElement.clientWidth / 2;
    viewportElement.scrollTop =
      (viewportElement.scrollTop + viewportElement.clientHeight / 2) * ratio -
      viewportElement.clientHeight / 2;
    updateViewport();
  }, [updateViewport, viewportRef, zoom]);

  useEffect(() => {
    const viewportElement = viewportRef.current;
    if (!viewportElement) return;

    const handleWheel = (event: WheelEvent) => {
      if (!event.ctrlKey && !event.metaKey) return;
      event.preventDefault();

      const nextZoom = Math.max(
        0.5,
        Math.min(
          1.6,
          Number((zoom * Math.exp(-event.deltaY * 0.002)).toFixed(2)),
        ),
      );
      if (nextZoom === zoom) return;

      const bounds = viewportElement.getBoundingClientRect();
      const viewportX = event.clientX - bounds.left;
      const viewportY = event.clientY - bounds.top;
      pendingZoomAnchorRef.current = {
        viewportX,
        viewportY,
        worldX: (viewportElement.scrollLeft + viewportX) / zoom,
        worldY: (viewportElement.scrollTop + viewportY) / zoom,
      };
      onZoomChange(nextZoom);
    };

    viewportElement.addEventListener("wheel", handleWheel, { passive: false });
    return () => {
      viewportElement.removeEventListener("wheel", handleWheel);
    };
  }, [onZoomChange, viewportRef, zoom]);

  return viewport;
};

export const useStrategyMapCanvasNodeDimensions = ({
  canvasRef,
  dimensionsRef,
}: {
  canvasRef: RefObject<HTMLDivElement | null>;
  dimensionsRef: MutableRefObject<Record<string, StrategyNodeDimensions>>;
}) => {
  const [dimensions, setDimensions] = useState<
    Record<string, StrategyNodeDimensions>
  >({});

  useLayoutEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const measureNodes = (nodes: Iterable<HTMLElement>) => {
      const nextDimensions = { ...dimensionsRef.current };
      const changedNodeIds: string[] = [];

      for (const node of nodes) {
        const nodeId = node.dataset.nodeId;
        if (!nodeId) continue;

        const next = { height: node.offsetHeight, width: node.offsetWidth };
        const previous = Object.hasOwn(nextDimensions, nodeId)
          ? nextDimensions[nodeId]
          : undefined;
        if (
          previous &&
          previous.height === next.height &&
          previous.width === next.width
        ) {
          continue;
        }

        nextDimensions[nodeId] = next;
        changedNodeIds.push(nodeId);
      }

      if (changedNodeIds.length === 0) return;
      dimensionsRef.current = nextDimensions;
      setDimensions(nextDimensions);
    };
    const observer = new ResizeObserver((entries) => {
      measureNodes(entries.map((entry) => entry.target as HTMLElement));
    });
    const observedNodes = new Set<HTMLElement>();
    const syncObservedNodes = () => {
      const nodes = new Set(
        canvas.querySelectorAll<HTMLElement>("[data-node-id]"),
      );

      observedNodes.forEach((node) => {
        if (nodes.has(node)) return;
        observer.unobserve(node);
        observedNodes.delete(node);
      });
      nodes.forEach((node) => {
        if (observedNodes.has(node)) return;
        observer.observe(node);
        observedNodes.add(node);
      });
      measureNodes(nodes);
    };
    const mutationObserver = new MutationObserver(syncObservedNodes);

    mutationObserver.observe(canvas, { childList: true, subtree: true });
    syncObservedNodes();

    return () => {
      mutationObserver.disconnect();
      observer.disconnect();
    };
  }, [canvasRef, dimensionsRef]);

  return dimensions;
};
