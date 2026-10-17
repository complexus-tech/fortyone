"use client";
/* eslint-disable jsx-a11y/no-noninteractive-tabindex -- The scrollable PDF region must be focusable for keyboard scrolling and text selection. */

import type {
  PDFDocumentProxy,
  PDFDocumentLoadingTask,
  PDFPageProxy,
  RenderTask,
} from "pdfjs-dist";
import type * as PdfLibraryModule from "pdfjs-dist/legacy/build/pdf.mjs";
import { useEffect, useRef, useState } from "react";
import type { PreviewAttachment } from "./types";
import { PreviewError, PreviewLoading } from "./preview-state";
import { PdfControls } from "./pdf-controls";
import type { PdfFit } from "./pdf-controls";
import { useElementSize } from "./use-element-size";
import "./pdf-text-layer.css";

type PdfLibrary = typeof PdfLibraryModule;
type PdfSource = { document: PDFDocumentProxy; library: PdfLibrary };

const MAX_CANVAS_PIXELS = 8_000_000;
const MAX_CANVAS_EDGE = 4096;

const documentError = (error: unknown) => {
  if (error instanceof Error && error.name === "PasswordException") {
    return "This PDF is password protected. Download it to open it in your PDF app.";
  }
  return "This PDF could not be loaded. It may be unavailable or damaged. Retry or download the original file.";
};

export const PdfPreview = ({
  attachment,
}: {
  attachment: PreviewAttachment;
}) => {
  const [source, setSource] = useState<PdfSource | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [attempt, setAttempt] = useState(0);
  const [pageNumber, setPageNumber] = useState(1);
  const [zoom, setZoom] = useState(1);
  const [fit, setFit] = useState<PdfFit>("page");
  const [rendering, setRendering] = useState(true);
  const viewportRef = useRef<HTMLDivElement>(null);
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const textRef = useRef<HTMLDivElement>(null);
  const pageRequestRef = useRef<{
    document: PDFDocumentProxy;
    pageNumber: number;
  } | null>(null);
  const size = useElementSize(viewportRef);

  useEffect(() => {
    let cancelled = false;
    const isCancelled = () => cancelled;
    let loadingTask: PDFDocumentLoadingTask | undefined;
    setError(null);
    setSource(null);
    const load = async () => {
      try {
        const library = await import("pdfjs-dist/legacy/build/pdf.mjs");
        if (isCancelled()) return;
        const assetRoot = `/pdfjs/${library.version}/`;
        library.GlobalWorkerOptions.workerSrc = `${assetRoot}pdf.worker.min.mjs`;
        loadingTask = library.getDocument({
          url: attachment.url,
          cMapUrl: `${assetRoot}cmaps/`,
          cMapPacked: true,
          standardFontDataUrl: `${assetRoot}standard_fonts/`,
          wasmUrl: `${assetRoot}wasm/`,
          // A maxImageSize limit silently omits oversized embedded images. Keep
          // complete pages and bound raster output instead; source decoding may
          // still use more memory than the final canvas budget.
          canvasMaxAreaInBytes: MAX_CANVAS_PIXELS * 4,
          stopAtErrors: true,
        });
        const document = await loadingTask.promise;
        if (!isCancelled()) setSource({ document, library });
      } catch (failure) {
        if (!isCancelled()) setError(documentError(failure));
      }
    };
    void load();
    return () => {
      cancelled = true;
      // Teardown may race the worker finishing an already-aborted load.
      void loadingTask?.destroy().catch(() => undefined);
    };
  }, [attachment.url, attempt]);

  useEffect(() => {
    if (!source || !size.width || !size.height) return;
    const canvas = canvasRef.current;
    const textContainer = textRef.current;
    if (!canvas || !textContainer) return;
    let cancelled = false;
    const isCancelled = () => cancelled;
    let renderTask: RenderTask | undefined;
    let textLayer: InstanceType<PdfLibrary["TextLayer"]> | undefined;
    let activePage: PDFPageProxy | undefined;
    let completion: Promise<unknown> | undefined;
    const request = { document: source.document, pageNumber };
    pageRequestRef.current = request;
    const releasePage = (page: PDFPageProxy) => {
      const current = pageRequestRef.current;
      // A new zoom/resize can request this same page before getPage resolves.
      // Leave its resources intact until we switch pages or close the viewer.
      if (
        current?.document !== request.document ||
        current.pageNumber !== request.pageNumber
      )
        page.cleanup();
    };
    setRendering(true);
    textContainer.replaceChildren();
    const render = async () => {
      try {
        const page = await source.document.getPage(pageNumber);
        if (isCancelled()) {
          releasePage(page);
          return;
        }
        activePage = page;
        const base = page.getViewport({ scale: 1 });
        const availableWidth = Math.max(1, size.width - 32);
        const availableHeight = Math.max(1, size.height - 32);
        const fittedScale =
          fit === "width"
            ? availableWidth / base.width
            : Math.min(
                availableWidth / base.width,
                availableHeight / base.height,
              );
        const viewport = page.getViewport({ scale: fittedScale * zoom });
        const outputScale = Math.min(
          window.devicePixelRatio || 1,
          2,
          Math.sqrt(MAX_CANVAS_PIXELS / (viewport.width * viewport.height)),
          MAX_CANVAS_EDGE / Math.max(viewport.width, viewport.height),
        );
        canvas.width = Math.max(1, Math.floor(viewport.width * outputScale));
        canvas.height = Math.max(1, Math.floor(viewport.height * outputScale));
        canvas.style.width = `${viewport.width}px`;
        canvas.style.height = `${viewport.height}px`;
        const context = canvas.getContext("2d");
        if (!context) throw new Error("Canvas unavailable");
        renderTask = page.render({
          canvas,
          canvasContext: context,
          viewport,
          transform: [outputScale, 0, 0, outputScale, 0, 0],
          background: "rgb(255,255,255)",
          annotationMode: source.library.AnnotationMode.DISABLE,
        });
        const renderText = async () => {
          const textContent = await page.getTextContent();
          if (isCancelled()) return;
          textContainer.style.setProperty(
            "--total-scale-factor",
            String(viewport.scale * viewport.userUnit),
          );
          textLayer = new source.library.TextLayer({
            textContentSource: textContent,
            container: textContainer,
            viewport,
          });
          await textLayer.render();
        };
        // Attach both rejection handlers immediately, including during text extraction.
        const operations = [renderTask.promise, renderText()];
        completion = Promise.allSettled(operations);
        await Promise.all(operations);
        if (!isCancelled()) setRendering(false);
      } catch {
        if (!isCancelled()) {
          setError(
            "This PDF page could not be displayed. Retry or download the original file.",
          );
          setRendering(false);
        }
      }
    };
    void render();
    return () => {
      cancelled = true;
      renderTask?.cancel();
      textLayer?.cancel();
      if (pageRequestRef.current === request) pageRequestRef.current = null;
      void (completion ?? Promise.resolve()).then(() => {
        if (activePage) releasePage(activePage);
      });
      textContainer.replaceChildren();
      canvas.width = 0;
      canvas.height = 0;
    };
  }, [source, pageNumber, zoom, fit, size.width, size.height]);

  return (
    <div className="flex min-h-0 flex-1 flex-col">
      {source && !error ? (
        <PdfControls
          fit={fit}
          onFitChange={(value) => {
            setFit(value);
            setZoom(1);
          }}
          onPageChange={(value) => {
            setPageNumber(value);
            viewportRef.current?.scrollTo({ top: 0, left: 0 });
          }}
          onZoomChange={setZoom}
          page={pageNumber}
          total={source.document.numPages}
          zoom={zoom}
        />
      ) : null}
      <div
        aria-label="PDF page preview"
        className="bg-surface-muted/40 relative min-h-0 flex-1 overflow-auto overscroll-contain p-4"
        data-preview-local-keys
        ref={viewportRef}
        role="region"
        tabIndex={0}
      >
        {error ? (
          <PreviewError
            message={error}
            onRetry={() => {
              setError(null);
              setPageNumber(1);
              setAttempt((value) => value + 1);
            }}
          />
        ) : (
          <>
            {!source || rendering ? (
              <div className="pointer-events-none absolute inset-0 z-10 flex">
                <PreviewLoading
                  label={source ? "Rendering page…" : "Loading PDF…"}
                />
              </div>
            ) : null}
            <div
              aria-busy={!source || rendering}
              aria-label={`Page ${pageNumber} of ${source?.document.numPages ?? "…"} of ${attachment.fileName}`}
              className="relative mx-auto w-max bg-white shadow-sm"
              role="region"
              style={{
                visibility: source && !rendering ? "visible" : "hidden",
              }}
            >
              <canvas aria-hidden ref={canvasRef} />
              <div
                className="attachment-pdf-text-layer textLayer"
                ref={textRef}
              />
            </div>
          </>
        )}
      </div>
    </div>
  );
};
