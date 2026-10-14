"use client";

import type { MutableRefObject, RefObject } from "react";
import { useLayoutEffect, useState } from "react";
import type { StrategyNodeDimensions } from "./strategy-map-layout";

const NODE_SELECTOR = "[data-node-id]";

const collectNodes = (root: Node, nodes: Set<HTMLElement>) => {
  if (!(root instanceof HTMLElement)) return;

  if (root.matches(NODE_SELECTOR)) nodes.add(root);
  root.querySelectorAll<HTMLElement>(NODE_SELECTOR).forEach((node) => {
    nodes.add(node);
  });
};

const getBorderBoxDimensions = (entry: ResizeObserverEntry) => {
  // Older WebKit exposed a single size rather than the current array shape.
  const sizes = entry.borderBoxSize as
    | readonly ResizeObserverSize[]
    | ResizeObserverSize
    | undefined;
  const size = sizes && ("inlineSize" in sizes ? sizes : sizes[0]);
  if (size) return { height: size.blockSize, width: size.inlineSize };

  // A contentRect excludes padding and borders. Preserve connector endpoints
  // on browsers without borderBoxSize by measuring only the resized node.
  const node = entry.target as HTMLElement;
  return { height: node.offsetHeight, width: node.offsetWidth };
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

    let disposed = false;
    const observedNodes = new Set<HTMLElement>();
    const publishDimensions = (
      measurements: Iterable<readonly [HTMLElement, StrategyNodeDimensions]>,
    ) => {
      let nextDimensions = dimensionsRef.current;

      for (const [node, next] of measurements) {
        const nodeId = node.dataset.nodeId;
        if (!nodeId || !observedNodes.has(node)) continue;

        const previous = Object.hasOwn(nextDimensions, nodeId)
          ? nextDimensions[nodeId]
          : undefined;
        if (previous?.height === next.height && previous.width === next.width) {
          continue;
        }

        if (nextDimensions === dimensionsRef.current) {
          nextDimensions = { ...nextDimensions };
        }
        nextDimensions[nodeId] = next;
      }

      if (nextDimensions === dimensionsRef.current) return;
      dimensionsRef.current = nextDimensions;
      setDimensions(nextDimensions);
    };
    const resizeObserver = new ResizeObserver((entries) => {
      if (disposed) return;
      publishDimensions(
        entries
          .filter((entry) => observedNodes.has(entry.target as HTMLElement))
          .map((entry) => [
            entry.target as HTMLElement,
            getBorderBoxDimensions(entry),
          ]),
      );
    });
    const syncNodes = (nodes: Iterable<HTMLElement>) => {
      const addedNodes: HTMLElement[] = [];

      for (const node of nodes) {
        if (!canvas.contains(node)) {
          if (observedNodes.delete(node)) resizeObserver.unobserve(node);
          // Retain offscreen dimensions so existing edges keep their geometry.
          continue;
        }
        if (observedNodes.has(node)) continue;

        observedNodes.add(node);
        resizeObserver.observe(node, { box: "border-box" });
        const nodeId = node.dataset.nodeId;
        if (nodeId && !Object.hasOwn(dimensionsRef.current, nodeId)) {
          addedNodes.push(node);
        }
      }

      publishDimensions(
        addedNodes.map((node) => [
          node,
          { height: node.offsetHeight, width: node.offsetWidth },
        ]),
      );
    };
    const mutationObserver = new MutationObserver((records) => {
      if (disposed) return;
      const changedNodes = new Set<HTMLElement>();
      records.forEach((record) => {
        record.addedNodes.forEach((node) => {
          collectNodes(node, changedNodes);
        });
        record.removedNodes.forEach((node) => {
          collectNodes(node, changedNodes);
        });
      });
      // Inspect final containment once, so moving a node within the canvas does
      // not unsubscribe, remeasure, or publish an intermediate removal.
      syncNodes(changedNodes);
    });

    mutationObserver.observe(canvas, { childList: true, subtree: true });
    syncNodes(canvas.querySelectorAll<HTMLElement>(NODE_SELECTOR));
    setDimensions(dimensionsRef.current);

    return () => {
      disposed = true;
      mutationObserver.disconnect();
      resizeObserver.disconnect();
      observedNodes.clear();
    };
  }, [canvasRef, dimensionsRef]);

  return dimensions;
};
