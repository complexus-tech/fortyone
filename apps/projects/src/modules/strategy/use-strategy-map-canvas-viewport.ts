"use client";

import type { RefObject } from "react";
import { useCallback, useLayoutEffect, useRef, useState } from "react";
import {
  createStrategyMapViewport,
  getInitialStrategyMapViewport,
  getRetainedStrategyMapViewport,
  type StrategyMapViewport,
} from "./strategy-map-visibility";
import {
  STRATEGY_MAP_MAX_ZOOM,
  STRATEGY_MAP_MIN_ZOOM,
} from "./strategy-map-zoom";

export { useStrategyMapCanvasNodeDimensions } from "./use-strategy-map-node-dimensions";

type ZoomAnchor = {
  viewportX: number;
  viewportY: number;
  worldX: number;
  worldY: number;
};

type ZoomRequest = { anchor: ZoomAnchor; zoom: number };
type CameraSnapshot = {
  actualScrollLeft: number;
  actualScrollTop: number;
  clientHeight: number;
  clientWidth: number;
  scrollLeft: number;
  scrollTop: number;
  zoom: number;
};

type UseStrategyMapCanvasViewportOptions = {
  layoutWidth: number;
  onZoomChange: (zoom: number) => void;
  resetSignal: number;
  viewportRef: RefObject<HTMLDivElement | null>;
  zoom: number;
};

export const useStrategyMapCanvasViewport = ({
  layoutWidth,
  onZoomChange,
  resetSignal,
  viewportRef,
  zoom,
}: UseStrategyMapCanvasViewportOptions) => {
  const previousResetSignalRef = useRef(resetSignal);
  const cameraRef = useRef<CameraSnapshot | null>(null);
  const renderedZoomRef = useRef(zoom);
  const desiredZoomRef = useRef(zoom);
  const onZoomChangeRef = useRef(onZoomChange);
  const pendingWheelRef = useRef<ZoomRequest | null>(null);
  const publishedWheelRef = useRef<ZoomRequest | null>(null);
  const wheelFrameRef = useRef<number | null>(null);
  const [viewport, setViewport] = useState<StrategyMapViewport>(() =>
    getRetainedStrategyMapViewport(
      null,
      getInitialStrategyMapViewport(layoutWidth, zoom),
    ),
  );

  const captureCamera = useCallback(() => {
    const element = viewportRef.current;
    if (!element) return null;
    const previous = cameraRef.current;
    const actualScrollLeft = element.scrollLeft;
    const actualScrollTop = element.scrollTop;
    const sameScale = previous?.zoom === renderedZoomRef.current;
    const snapshot = {
      actualScrollLeft,
      actualScrollTop,
      clientHeight: element.clientHeight,
      clientWidth: element.clientWidth,
      scrollLeft: sameScale
        ? previous.scrollLeft + actualScrollLeft - previous.actualScrollLeft
        : actualScrollLeft,
      scrollTop: sameScale
        ? previous.scrollTop + actualScrollTop - previous.actualScrollTop
        : actualScrollTop,
      zoom: renderedZoomRef.current,
    };
    cameraRef.current = snapshot;
    return snapshot;
  }, [viewportRef]);

  const positionCamera = useCallback(
    (left: number, top: number) => {
      const element = viewportRef.current;
      if (!element) return;
      const scrollLeft = Math.max(
        0,
        Math.min(left, element.scrollWidth - element.clientWidth),
      );
      const scrollTop = Math.max(
        0,
        Math.min(top, element.scrollHeight - element.clientHeight),
      );
      element.scrollLeft = scrollLeft;
      element.scrollTop = scrollTop;
      // WebKit can round native scroll offsets to whole pixels. Retain the
      // fractional remainder instead of feeding that error into every pinch.
      cameraRef.current = {
        actualScrollLeft: element.scrollLeft,
        actualScrollTop: element.scrollTop,
        clientHeight: element.clientHeight,
        clientWidth: element.clientWidth,
        scrollLeft,
        scrollTop,
        zoom: renderedZoomRef.current,
      };
    },
    [viewportRef],
  );

  const updateViewport = useCallback(() => {
    const camera = captureCamera();
    if (!camera) return;
    const requiredViewport = createStrategyMapViewport(camera);
    // Native scrolling moves the same cards and paths together. React only
    // replenishes their shared render window before its safety margin runs out.
    setViewport((current) =>
      getRetainedStrategyMapViewport(current, requiredViewport),
    );
  }, [captureCamera]);

  useLayoutEffect(() => {
    onZoomChangeRef.current = onZoomChange;
  }, [onZoomChange]);

  useLayoutEffect(() => {
    const element = viewportRef.current;
    if (!element) return;
    // Preserve the last camera BEFORE React resizes the scaled scroll surface.
    // Reading scrollLeft here for the old zoom would use browser-clamped values.
    const previousCamera = cameraRef.current;
    const shouldReset =
      !previousCamera || previousResetSignalRef.current !== resetSignal;
    previousResetSignalRef.current = resetSignal;
    renderedZoomRef.current = zoom;

    if (shouldReset) {
      if (wheelFrameRef.current !== null) {
        cancelAnimationFrame(wheelFrameRef.current);
        wheelFrameRef.current = null;
      }
      pendingWheelRef.current = null;
      publishedWheelRef.current = null;
      desiredZoomRef.current = zoom;
      positionCamera((layoutWidth * zoom - element.clientWidth) / 2, 0);
    } else if (previousCamera.zoom !== zoom) {
      const wheelRequest = publishedWheelRef.current;
      const isWheelZoom = wheelRequest?.zoom === zoom;
      const anchor = isWheelZoom
        ? wheelRequest.anchor
        : {
            viewportX: previousCamera.clientWidth / 2,
            viewportY: previousCamera.clientHeight / 2,
            worldX:
              (previousCamera.scrollLeft + previousCamera.clientWidth / 2) /
              previousCamera.zoom,
            worldY:
              (previousCamera.scrollTop + previousCamera.clientHeight / 2) /
              previousCamera.zoom,
          };
      positionCamera(
        anchor.worldX * zoom - anchor.viewportX,
        anchor.worldY * zoom - anchor.viewportY,
      );
      publishedWheelRef.current = null;

      if (!isWheelZoom) {
        // An explicit toolbar action supersedes any uncommitted wheel input.
        if (wheelFrameRef.current !== null) {
          cancelAnimationFrame(wheelFrameRef.current);
          wheelFrameRef.current = null;
        }
        pendingWheelRef.current = null;
      }
      if (!pendingWheelRef.current) desiredZoomRef.current = zoom;
    }

    updateViewport();
  }, [
    layoutWidth,
    positionCamera,
    resetSignal,
    updateViewport,
    viewportRef,
    zoom,
  ]);

  useLayoutEffect(() => {
    const element = viewportRef.current;
    if (!element) return;
    let viewportFrame: number | null = null;
    const scheduleViewportUpdate = () => {
      // Keep a pre-resize anchor even when React hasn't published the next
      // culling window yet. No React work is needed for each scroll event.
      captureCamera();
      if (viewportFrame !== null) return;
      viewportFrame = requestAnimationFrame(() => {
        viewportFrame = null;
        updateViewport();
      });
    };
    const observer = new ResizeObserver(scheduleViewportUpdate);
    observer.observe(element);
    element.addEventListener("scroll", scheduleViewportUpdate, {
      passive: true,
    });

    const handleWheel = (event: WheelEvent) => {
      if (!event.ctrlKey && !event.metaKey) return;
      event.preventDefault();
      let deltaUnit = 1;
      if (event.deltaMode === WheelEvent.DOM_DELTA_LINE) deltaUnit = 16;
      if (event.deltaMode === WheelEvent.DOM_DELTA_PAGE) {
        deltaUnit = element.clientHeight;
      }
      const nextZoom = Math.max(
        STRATEGY_MAP_MIN_ZOOM,
        Math.min(
          STRATEGY_MAP_MAX_ZOOM,
          desiredZoomRef.current * Math.exp(-event.deltaY * deltaUnit * 0.002),
        ),
      );
      if (nextZoom === desiredZoomRef.current) return;
      const bounds = element.getBoundingClientRect();
      const viewportX = event.clientX - bounds.left;
      const viewportY = event.clientY - bounds.top;
      const camera = captureCamera();
      if (!camera) return;
      pendingWheelRef.current = {
        anchor: {
          viewportX,
          viewportY,
          worldX: (camera.scrollLeft + viewportX) / camera.zoom,
          worldY: (camera.scrollTop + viewportY) / camera.zoom,
        },
        zoom: nextZoom,
      };
      desiredZoomRef.current = nextZoom;
      if (wheelFrameRef.current !== null) return;
      wheelFrameRef.current = requestAnimationFrame(() => {
        wheelFrameRef.current = null;
        const request = pendingWheelRef.current;
        pendingWheelRef.current = null;
        if (!request) return;
        publishedWheelRef.current = request;
        onZoomChangeRef.current(request.zoom);
      });
    };
    element.addEventListener("wheel", handleWheel, { passive: false });

    return () => {
      if (viewportFrame !== null) cancelAnimationFrame(viewportFrame);
      if (wheelFrameRef.current !== null) {
        cancelAnimationFrame(wheelFrameRef.current);
        wheelFrameRef.current = null;
      }
      pendingWheelRef.current = null;
      observer.disconnect();
      element.removeEventListener("scroll", scheduleViewportUpdate);
      element.removeEventListener("wheel", handleWheel);
    };
  }, [captureCamera, updateViewport, viewportRef]);

  return viewport;
};
