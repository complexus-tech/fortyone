/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */

import { useState } from "react";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { AttachmentPreviewDialog } from "./attachment-preview-dialog";
import { getAdjacentPreviewId, getPreviewAttachments } from "./types";
import type { PreviewAttachment } from "./types";

const mockPdfRender = jest.fn();
const originalResizeObserver = globalThis.ResizeObserver;
jest.mock("./pdf-preview", () => ({
  PdfPreview: ({ attachment }: { attachment: PreviewAttachment }) => {
    mockPdfRender(attachment.id);
    return (
      <div data-testid="active-pdf">
        <input aria-label="PDF page number" defaultValue="1" />
      </div>
    );
  },
}));

const image = (id: number): PreviewAttachment => ({
  id: `image-${id}`,
  fileName: `Image ${id}.png`,
  mediaType: "image/png",
  sizeBytes: 1000,
  url: `/image-${id}.png`,
});
const images = Array.from({ length: 8 }, (_, index) => image(index + 1));
const pdf: PreviewAttachment = {
  id: "pdf",
  fileName: "Guide.pdf",
  mediaType: "application/pdf",
  sizeBytes: 3000,
  url: "/guide.pdf",
};
const audio: PreviewAttachment = {
  id: "audio",
  fileName: "Voice.m4a",
  mediaType: "audio/mp4",
  sizeBytes: 5000,
  url: "/voice.m4a",
};
const doc: PreviewAttachment = {
  id: "word",
  fileName: "Notes.docx",
  mediaType:
    "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
  sizeBytes: 2500,
  url: "/notes.docx",
};

const Harness = ({
  files = images,
  initialId = files[0]?.id ?? null,
}: {
  files?: PreviewAttachment[];
  initialId?: string | null;
}) => {
  const [selectedId, setSelectedId] = useState(initialId);
  const [open, setOpen] = useState(true);
  return (
    <AttachmentPreviewDialog
      attachments={files}
      onClose={() => {
        setOpen(false);
      }}
      onSelect={setSelectedId}
      open={open}
      selectedId={selectedId}
    />
  );
};

beforeEach(() => {
  globalThis.ResizeObserver = jest.fn(() => ({
    observe: jest.fn(),
    disconnect: jest.fn(),
    unobserve: jest.fn(),
  })) as unknown as typeof ResizeObserver;
});

afterEach(() => {
  cleanup();
  mockPdfRender.mockClear();
  globalThis.ResizeObserver = originalResizeObserver;
});

describe("controlled attachment preview", () => {
  it("opens the clicked image and allows all eight attachments with finite endpoints", () => {
    render(<Harness initialId="image-4" />);
    expect(screen.getByText("4 of 8")).toBeInTheDocument();
    fireEvent.load(screen.getByAltText("Image 4.png"));
    expect(
      screen.getByRole("img", { name: "Image 4.png" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("link", { name: "Download Image 4.png" }),
    ).toHaveAttribute("download", "Image 4.png");
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "ArrowRight" });
    expect(screen.getByText("5 of 8")).toBeInTheDocument();
    for (let index = 0; index < 3; index += 1)
      fireEvent.click(screen.getByRole("button", { name: "Next attachment" }));
    expect(screen.getByText("8 of 8")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Next attachment" }),
    ).toBeDisabled();
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "ArrowRight" });
    expect(screen.getByText("8 of 8")).toBeInTheDocument();
    for (let index = 0; index < 7; index += 1)
      fireEvent.click(
        screen.getByRole("button", { name: "Previous attachment" }),
      );
    expect(
      screen.getByRole("button", { name: "Previous attachment" }),
    ).toBeDisabled();
    expect(screen.getByText("1 of 8")).toBeInTheDocument();
  });

  it("hides file navigation for a single PDF while retaining its independent page control", async () => {
    render(<Harness files={[pdf]} />);
    expect(
      await screen.findByRole("textbox", { name: "PDF page number" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("group", { name: "Attachment navigation" }),
    ).not.toBeInTheDocument();
    expect(screen.queryByText("1 of 1")).not.toBeInTheDocument();
  });

  it("never mounts a PDF while viewing an image, loads only the active PDF, and unmounts it on close", async () => {
    render(
      <Harness
        files={[image(1), pdf, { ...pdf, id: "pdf-2", fileName: "Other.pdf" }]}
      />,
    );
    expect(mockPdfRender).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Next attachment" }));
    await screen.findByTestId("active-pdf");
    expect(mockPdfRender).toHaveBeenCalledWith("pdf");
    expect(mockPdfRender).not.toHaveBeenCalledWith("pdf-2");
    fireEvent.click(
      screen.getByRole("button", { name: "Close attachment preview" }),
    );
    await waitFor(() => {
      expect(screen.queryByTestId("active-pdf")).not.toBeInTheDocument();
    });
  });

  it("does not steal arrows from PDF inputs or other focused controls", async () => {
    render(<Harness files={[pdf, image(1)]} />);
    const pageInput = await screen.findByRole("textbox", {
      name: "PDF page number",
    });
    fireEvent.keyDown(pageInput, { key: "ArrowRight" });
    expect(screen.getByText("1 of 2")).toBeInTheDocument();
    fireEvent.keyDown(screen.getByRole("button", { name: "Next attachment" }), {
      key: "ArrowRight",
    });
    expect(screen.getByText("1 of 2")).toBeInTheDocument();
    fireEvent.keyDown(screen.getByRole("dialog"), { key: "ArrowRight" });
    expect(screen.getByText("2 of 2")).toBeInTheDocument();
  });

  it("excludes audio while retaining unsupported document fallback entries in order", () => {
    render(<Harness files={[image(1), audio, doc]} />);
    expect(screen.getByText("1 of 2")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Next attachment" }));
    expect(screen.getByText("2 of 2")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Download file" })).toHaveAttribute(
      "href",
      doc.url,
    );
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
    expect(
      getPreviewAttachments([image(1), audio, doc]).map((item) => item.id),
    ).toEqual(["image-1", "word"]);
    expect(getAdjacentPreviewId([image(1), doc], "missing", 1)).toBeNull();
  });

  it("does not open a modal for an audio selection", () => {
    render(<Harness files={[audio]} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("renders a parent error without mounting stale content and offers retry", () => {
    const retry = jest.fn();
    render(
      <AttachmentPreviewDialog
        attachments={[pdf]}
        error="Access to this file was removed."
        onClose={jest.fn()}
        onRetry={retry}
        onSelect={jest.fn()}
        open
        selectedId={pdf.id}
      />,
    );
    expect(screen.getByRole("alert")).toHaveTextContent(
      "Access to this file was removed.",
    );
    expect(mockPdfRender).not.toHaveBeenCalled();
    expect(screen.queryByRole("link")).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Retry preview" }));
    expect(retry).toHaveBeenCalledTimes(1);
  });

  it("shows unavailable rather than silently selecting a different file after deletion", () => {
    render(<Harness files={[image(1)]} initialId="deleted" />);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "This attachment is no longer available",
    );
    expect(screen.queryByRole("img")).not.toBeInTheDocument();
  });
});
