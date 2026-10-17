import { useState } from "react";
import { ChevronLeftIcon, ChevronRightIcon, MinusIcon, PlusIcon } from "icons";
import { Button, Input, Tooltip } from "ui";

export type PdfFit = "page" | "width";

export const PdfControls = ({
  page,
  total,
  onPageChange,
  zoom,
  onZoomChange,
  fit,
  onFitChange,
}: {
  page: number;
  total: number;
  onPageChange: (page: number) => void;
  zoom: number;
  onZoomChange: (zoom: number) => void;
  fit: PdfFit;
  onFitChange: (fit: PdfFit) => void;
}) => {
  const [pageInput, setPageInput] = useState({ page, value: String(page) });
  const pageValue = pageInput.page === page ? pageInput.value : String(page);
  const selectPage = (nextPage: number) => {
    setPageInput({ page: nextPage, value: String(nextPage) });
    onPageChange(nextPage);
  };
  const applyPage = () => {
    const requested = Number(pageValue);
    if (Number.isInteger(requested) && requested >= 1 && requested <= total)
      selectPage(requested);
    else setPageInput({ page, value: String(page) });
  };
  return (
    <div
      aria-label="PDF page and zoom controls"
      className="border-border flex shrink-0 flex-wrap items-center justify-center gap-2 border-b px-3 py-2"
      data-preview-local-keys
      role="group"
    >
      <Tooltip title="Previous page">
        <Button
          aria-label="Previous PDF page"
          asIcon
          color="tertiary"
          disabled={page === 1}
          onClick={() => {
            selectPage(page - 1);
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <ChevronLeftIcon className="size-4" />
        </Button>
      </Tooltip>
      <form
        className="flex items-center gap-2"
        onSubmit={(event) => {
          event.preventDefault();
          applyPage();
        }}
      >
        <span className="text-text-muted">Page</span>
        <Input
          aria-label="PDF page number"
          className="w-12 px-1 text-center tabular-nums"
          inputMode="numeric"
          maxLength={6}
          onBlur={applyPage}
          onChange={(event) => {
            setPageInput({ page, value: event.target.value });
          }}
          size="sm"
          value={pageValue}
        />
        <span className="text-text-muted tabular-nums">of {total}</span>
      </form>
      <Tooltip title="Next page">
        <Button
          aria-label="Next PDF page"
          asIcon
          color="tertiary"
          disabled={page === total}
          onClick={() => {
            selectPage(page + 1);
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <ChevronRightIcon className="size-4" />
        </Button>
      </Tooltip>
      <span aria-hidden className="bg-border mx-1 h-4 w-px" />
      <Tooltip title="Zoom out">
        <Button
          aria-label="Zoom PDF out"
          asIcon
          color="tertiary"
          disabled={zoom <= 0.5}
          onClick={() => {
            onZoomChange(Math.max(0.5, zoom - 0.25));
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <MinusIcon className="size-4" />
        </Button>
      </Tooltip>
      <span aria-label="PDF zoom" className="w-12 text-center tabular-nums">
        {Math.round(zoom * 100)}%
      </span>
      <Tooltip title="Zoom in">
        <Button
          aria-label="Zoom PDF in"
          asIcon
          color="tertiary"
          disabled={zoom >= 3}
          onClick={() => {
            onZoomChange(Math.min(3, zoom + 0.25));
          }}
          size="sm"
          type="button"
          variant="naked"
        >
          <PlusIcon className="size-4" />
        </Button>
      </Tooltip>
      <Button
        aria-pressed={fit === "page" && zoom === 1}
        color="tertiary"
        onClick={() => {
          onFitChange("page");
        }}
        size="sm"
        type="button"
        variant="naked"
      >
        Fit page
      </Button>
      <Button
        aria-pressed={fit === "width" && zoom === 1}
        color="tertiary"
        onClick={() => {
          onFitChange("width");
        }}
        size="sm"
        type="button"
        variant="naked"
      >
        Fit width
      </Button>
    </div>
  );
};
