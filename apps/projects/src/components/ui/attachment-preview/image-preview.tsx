"use client";
/* eslint-disable jsx-a11y/no-noninteractive-tabindex -- The scrollable media region must be focusable for keyboard panning. */

import { useRef, useState } from "react";
import { MinusIcon, PlusIcon } from "icons";
import { Button, Tooltip } from "ui";
import { PreviewError, PreviewLoading } from "./preview-state";
import type { PreviewAttachment } from "./types";
import { useElementSize } from "./use-element-size";

const MIN_ZOOM = 1;
const MAX_ZOOM = 4;

export const ImagePreview = ({
  attachment,
}: {
  attachment: PreviewAttachment;
}) => {
  const [zoom, setZoom] = useState(MIN_ZOOM);
  const [status, setStatus] = useState<"loading" | "ready" | "failed">(
    "loading",
  );
  const [attempt, setAttempt] = useState(0);
  const viewport = useRef<HTMLDivElement>(null);
  const viewportSize = useElementSize(viewport);
  const [naturalSize, setNaturalSize] = useState({ width: 0, height: 0 });
  const fittedScale =
    naturalSize.width && naturalSize.height
      ? Math.min(
          1,
          Math.max(1, viewportSize.width - 32) / naturalSize.width,
          Math.max(1, viewportSize.height - 32) / naturalSize.height,
        )
      : 1;
  const width = naturalSize.width * fittedScale * zoom;
  const height = naturalSize.height * fittedScale * zoom;
  const drag = useRef<{
    pointerId: number;
    x: number;
    y: number;
    left: number;
    top: number;
  } | null>(null);

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      <div
        aria-label="Image controls"
        className="border-border flex shrink-0 items-center justify-center gap-2 border-b px-3 py-2"
        role="group"
      >
        <Tooltip title="Zoom out">
          <Button
            aria-label="Zoom image out"
            asIcon
            color="tertiary"
            disabled={zoom <= MIN_ZOOM || status !== "ready"}
            onClick={() => {
              setZoom((current) => Math.max(MIN_ZOOM, current - 0.5));
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            <MinusIcon className="size-4" />
          </Button>
        </Tooltip>
        <span aria-label="Image zoom" className="w-12 text-center tabular-nums">
          {Math.round(zoom * 100)}%
        </span>
        <Tooltip title="Zoom in">
          <Button
            aria-label="Zoom image in"
            asIcon
            color="tertiary"
            disabled={zoom >= MAX_ZOOM || status !== "ready"}
            onClick={() => {
              setZoom((current) => Math.min(MAX_ZOOM, current + 0.5));
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            <PlusIcon className="size-4" />
          </Button>
        </Tooltip>
        <Button
          color="tertiary"
          disabled={zoom === MIN_ZOOM}
          onClick={() => {
            setZoom(MIN_ZOOM);
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          Fit image
        </Button>
      </div>
      <div
        aria-label={
          zoom > MIN_ZOOM
            ? "Image preview. Scroll or drag to pan."
            : "Image preview"
        }
        className="bg-surface-muted/40 relative min-h-0 flex-1 overflow-auto overscroll-contain"
        data-preview-local-keys={zoom > MIN_ZOOM ? "" : undefined}
        onLostPointerCapture={() => {
          drag.current = null;
        }}
        onPointerDown={(event) => {
          if (zoom === MIN_ZOOM || event.button !== 0 || !viewport.current)
            return;
          drag.current = {
            pointerId: event.pointerId,
            x: event.clientX,
            y: event.clientY,
            left: viewport.current.scrollLeft,
            top: viewport.current.scrollTop,
          };
          event.currentTarget.setPointerCapture(event.pointerId);
        }}
        onPointerMove={(event) => {
          const start = drag.current;
          if (
            !start ||
            start.pointerId !== event.pointerId ||
            !viewport.current
          )
            return;
          viewport.current.scrollLeft = start.left + start.x - event.clientX;
          viewport.current.scrollTop = start.top + start.y - event.clientY;
        }}
        onPointerUp={(event) => {
          if (event.currentTarget.hasPointerCapture(event.pointerId))
            event.currentTarget.releasePointerCapture(event.pointerId);
          drag.current = null;
        }}
        ref={viewport}
        role="region"
        style={{
          cursor: zoom > MIN_ZOOM ? "grab" : undefined,
          touchAction: zoom > MIN_ZOOM ? "none" : undefined,
        }}
        tabIndex={0}
      >
        {status === "loading" ? (
          <div className="absolute inset-0 flex">
            <PreviewLoading />
          </div>
        ) : null}
        {status === "failed" ? (
          <PreviewError
            message="This image could not be loaded. Retry or download the original file."
            onRetry={() => {
              setStatus("loading");
              setAttempt((current) => current + 1);
            }}
          />
        ) : (
          <div
            className="flex items-center justify-center p-4"
            style={{
              width: Math.max(viewportSize.width, width + 32),
              height: Math.max(viewportSize.height, height + 32),
              minWidth: "100%",
              minHeight: "100%",
            }}
          >
            {/* eslint-disable-next-line @next/next/no-img-element -- Private and local blob URLs must bypass public image optimization. */}
            <img
              alt={attachment.fileName}
              className="shrink-0 object-contain select-none"
              draggable={false}
              key={attempt}
              onError={() => {
                setStatus("failed");
              }}
              onLoad={(event) => {
                setNaturalSize({
                  width: event.currentTarget.naturalWidth,
                  height: event.currentTarget.naturalHeight,
                });
                setStatus("ready");
              }}
              src={attachment.url}
              style={{
                width: width || undefined,
                height: height || undefined,
                visibility: status === "ready" ? "visible" : "hidden",
              }}
            />
          </div>
        )}
      </div>
    </div>
  );
};
