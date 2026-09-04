"use client";

import type { ComponentProps, RefObject } from "react";
import { cn } from "lib";
import { Box, Flex } from "ui";
import { walkthroughTargets } from "@/shared/walkthrough/targets";
import { CanvasConnections } from "./strategy-map-canvas-renderers";
import { StrategyCanvasNodes } from "./strategy-map-canvas-nodes";

type StrategyMapCanvasSurfaceProps = {
  canvasConnections: ComponentProps<typeof CanvasConnections>;
  canvasNodes: ComponentProps<typeof StrategyCanvasNodes>;
  canvasRef: RefObject<HTMLDivElement | null>;
  isPanning: boolean;
  layout: { height: number; width: number };
  onViewportPointerCancel: ComponentProps<"div">["onPointerCancel"];
  onViewportPointerDown: ComponentProps<"div">["onPointerDown"];
  onViewportPointerMove: ComponentProps<"div">["onPointerMove"];
  onViewportPointerUp: ComponentProps<"div">["onPointerUp"];
  viewportRef: RefObject<HTMLDivElement | null>;
  zoom: number;
};

export const StrategyMapCanvasSurface = ({
  canvasConnections,
  canvasNodes,
  canvasRef,
  isPanning,
  layout,
  onViewportPointerCancel,
  onViewportPointerDown,
  onViewportPointerMove,
  onViewportPointerUp,
  viewportRef,
  zoom,
}: StrategyMapCanvasSurfaceProps) => (
  <Box className="relative h-full overflow-hidden">
    <div
      className={cn(
        "bg-surface-muted/20 dark:bg-background h-full overflow-auto overscroll-none",
        isPanning ? "cursor-grabbing" : "cursor-grab",
      )}
      onPointerCancel={onViewportPointerCancel}
      onPointerDown={onViewportPointerDown}
      onPointerMove={onViewportPointerMove}
      onPointerUp={onViewportPointerUp}
      ref={viewportRef}
      style={{ touchAction: "none" }}
    >
      <div
        className="relative"
        style={{
          height: layout.height * zoom,
          width: layout.width * zoom,
        }}
      >
        <div
          className="bg-surface-muted/15 dark:bg-background absolute top-0 left-0 origin-top-left"
          ref={canvasRef}
          style={{
            backgroundImage:
              "radial-gradient(var(--color-border-strong) 1.15px, transparent 1.15px)",
            backgroundSize: "22px 22px",
            height: layout.height,
            transform: `scale(${zoom})`,
            width: layout.width,
          }}
        >
          <CanvasConnections {...canvasConnections} />
          <StrategyCanvasNodes {...canvasNodes} />
        </div>
      </div>
    </div>

    <Flex
      align="center"
      className="border-border bg-surface/90 text-text-muted pointer-events-none absolute bottom-4 left-1/2 z-40 -translate-x-1/2 gap-2 rounded-lg border px-3.5 py-2 text-sm shadow-lg backdrop-blur"
      data-canvas-control
      data-walkthrough-target={walkthroughTargets.strategyCanvasHelp}
    >
      <span>Drag cards to position</span>
      <span aria-hidden>·</span>
      <span>Drag the canvas to pan</span>
      <span aria-hidden>·</span>
      <span>Right-click for actions</span>
    </Flex>
  </Box>
);
