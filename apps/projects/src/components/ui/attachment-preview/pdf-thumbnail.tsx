"use client";

import type { PDFDocumentLoadingTask, RenderTask } from "pdfjs-dist";
import { useEffect, useRef, useState } from "react";
import { cn } from "lib";
import { queuePdfThumbnail } from "./pdf-thumbnail-queue";
import { CachedAttachmentThumbnail } from "./cached-thumbnail";

type PdfThumbnailProps = {
  url: string;
  fileName: string;
  className?: string;
  thumbnailUrl?: string;
  refreshKey?: number;
};

const THUMBNAIL_WIDTH = 420;
const THUMBNAIL_HEIGHT = THUMBNAIL_WIDTH / 1.6;
const MAX_CANVAS_PIXELS = 1_000_000;

const renderFirstPage = async (
  canvas: HTMLCanvasElement,
  url: string,
  signal: AbortSignal,
) => {
  let loadingTask: PDFDocumentLoadingTask | undefined;
  let rendering: RenderTask | undefined;
  let destruction: Promise<void> | undefined;
  const isCanceled = () => signal.aborted;
  const destroyDocument = () => {
    if (!loadingTask) return Promise.resolve();
    // LoadingTask owns the document and worker. Its destruction also aborts
    // outstanding range requests, including a load that has not resolved yet.
    destruction ??= loadingTask.destroy().catch(() => undefined);
    return destruction;
  };
  const cancel = () => {
    const task = rendering;
    rendering = undefined;
    task?.cancel();
    void destroyDocument();
  };
  signal.addEventListener("abort", cancel, { once: true });

  try {
    if (isCanceled()) return;
    const library = await import("pdfjs-dist/legacy/build/pdf.mjs");
    if (isCanceled()) return;
    const assetRoot = `/pdfjs/${library.version}/`;
    library.GlobalWorkerOptions.workerSrc = `${assetRoot}pdf.worker.min.mjs`;
    loadingTask = library.getDocument({
      url,
      cMapUrl: `${assetRoot}cmaps/`,
      cMapPacked: true,
      standardFontDataUrl: `${assetRoot}standard_fonts/`,
      wasmUrl: `${assetRoot}wasm/`,
      disableAutoFetch: true,
      // PDF.js requires streaming off to avoid fetching the rest of a file in
      // the background. Range-capable preview endpoints load only needed data.
      disableStream: true,
      canvasMaxAreaInBytes: MAX_CANVAS_PIXELS * 4,
      stopAtErrors: true,
    });
    const document = await loadingTask.promise;
    if (isCanceled()) return;
    const page = await document.getPage(1);
    if (isCanceled()) return;
    const base = page.getViewport({ scale: 1 });
    if (
      !Number.isFinite(base.width) ||
      !Number.isFinite(base.height) ||
      base.width <= 0 ||
      base.height <= 0
    ) {
      throw new Error("PDF page has invalid dimensions");
    }
    const viewport = page.getViewport({ scale: THUMBNAIL_WIDTH / base.width });
    const height = Math.min(viewport.height, THUMBNAIL_HEIGHT);
    const outputScale = Math.min(
      window.devicePixelRatio || 1,
      2,
      Math.sqrt(MAX_CANVAS_PIXELS / (viewport.width * height)),
    );
    canvas.width = Math.max(1, Math.floor(viewport.width * outputScale));
    canvas.height = Math.max(1, Math.floor(height * outputScale));
    const context = canvas.getContext("2d", { alpha: false });
    if (!context) throw new Error("Canvas unavailable");
    // The canvas intentionally clips the first page to its top section. This
    // keeps even unusually tall PDFs within a small, predictable pixel budget.
    rendering = page.render({
      canvas,
      canvasContext: context,
      viewport,
      transform: [outputScale, 0, 0, outputScale, 0, 0],
      background: "rgb(255,255,255)",
      annotationMode: library.AnnotationMode.DISABLE,
    });
    await rendering.promise;
    rendering = undefined;
    if (isCanceled()) return;
    page.cleanup();
  } finally {
    signal.removeEventListener("abort", cancel);
    rendering?.cancel();
    await destroyDocument();
  }
};

const Thumbnail = ({ url, fileName, className }: PdfThumbnailProps) => {
  const [status, setStatus] = useState<"loading" | "ready" | "error">(
    "loading",
  );
  const containerRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const container = containerRef.current;
    const canvas = canvasRef.current;
    if (!container || !canvas) return;
    const controller = new AbortController();
    let started = false;
    const start = () => {
      if (started || controller.signal.aborted) return;
      started = true;
      void queuePdfThumbnail(
        () => renderFirstPage(canvas, url, controller.signal),
        controller.signal,
      ).then(
        () => {
          if (!controller.signal.aborted) setStatus("ready");
        },
        () => {
          if (!controller.signal.aborted) setStatus("error");
        },
      );
    };
    const observer =
      typeof IntersectionObserver === "undefined"
        ? null
        : new IntersectionObserver(
            (entries) => {
              if (!entries.some((entry) => entry.isIntersecting)) return;
              observer?.disconnect();
              start();
            },
            { rootMargin: "240px" },
          );
    if (observer) observer.observe(container);
    else start();
    return () => {
      observer?.disconnect();
      controller.abort();
      canvas.width = 0;
      canvas.height = 0;
    };
  }, [url]);

  return (
    <div
      aria-busy={status === "loading"}
      aria-label={
        status === "error"
          ? `PDF preview unavailable for ${fileName}`
          : `First page of ${fileName}`
      }
      className={cn(
        "relative aspect-[8/5] w-full overflow-hidden bg-white",
        className,
      )}
      ref={containerRef}
      role="img"
    >
      <canvas
        aria-hidden
        className="absolute top-0 left-0 block h-auto w-full"
        ref={canvasRef}
        style={{ visibility: status === "ready" ? "visible" : "hidden" }}
      />
      {status !== "ready" ? (
        <span className="bg-surface-muted text-text-muted absolute inset-0 flex items-center justify-center px-4 text-center">
          {status === "error"
            ? "PDF preview unavailable · Open to view"
            : "Loading PDF preview…"}
        </span>
      ) : null}
    </div>
  );
};

export const PdfThumbnail = (props: PdfThumbnailProps) =>
  props.thumbnailUrl ? (
    <CachedAttachmentThumbnail
      className={props.className}
      fileName={props.fileName}
      refreshKey={props.refreshKey}
      url={props.thumbnailUrl}
    />
  ) : (
    <Thumbnail key={props.url} {...props} />
  );
