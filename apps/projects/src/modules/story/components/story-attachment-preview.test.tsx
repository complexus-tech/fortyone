/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { ReactNode } from "react";
import { fireEvent, render, screen, within } from "@testing-library/react";
import type { StoryAttachment } from "../types";
import { StoryAttachmentPreview } from "./story-attachment-preview";

jest.mock("player.style/sutro/react", () => ({
  __esModule: true,
  default: ({ children }: { children: ReactNode }) => <div>{children}</div>,
}));
jest.mock("@/hooks/owner", () => ({
  useIsAdminOrOwner: () => ({ isAdminOrOwner: false }),
}));
jest.mock("@/components/ui", () => ({ ConfirmDialog: () => null }));
jest.mock("./attachment-text-preview", () => ({
  AttachmentTextPreview: ({ file }: { file: StoryAttachment }) => (
    <pre aria-label={`${file.filename} text preview`}>File contents</pre>
  ),
}));

const file = (id: string, mimeType: string): StoryAttachment => ({
  id,
  filename: id,
  mimeType,
  size: 1024,
  url: `/attachments/${id}`,
  uploadedBy: "author",
  createdAt: "2026-10-16T12:00:00Z",
});
const IMAGE = file("screenshot.png", "image/png");
const PDF = file("review.pdf", "application/pdf");
const WORD = file(
  "plan.docx",
  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
);

describe("comment attachment preview behavior", () => {
  it("marks a video thumbnail decoratively and opens the real video player", () => {
    const video = file("walkthrough.mp4", "video/mp4");
    render(<StoryAttachmentPreview file={video} isInChat variant="comment" />);
    expect(screen.getByLabelText("walkthrough.mp4 thumbnail").tagName).toBe(
      "VIDEO",
    );
    const card = screen.getByRole("button", {
      name: "Preview walkthrough.mp4",
    });
    expect(
      card.parentElement?.querySelector('[aria-hidden="true"] svg'),
    ).toBeInTheDocument();
    fireEvent.click(card);
    expect(screen.getByRole("dialog").querySelector("video")).toHaveAttribute(
      "src",
      video.url,
    );
  });

  it("retains the real image preview and provides a download fallback when navigating to Office files", () => {
    const download = jest.fn();
    render(
      <StoryAttachmentPreview
        file={IMAGE}
        files={[IMAGE, WORD, PDF]}
        isInChat
        onDownload={download}
        unoptimizedImages
        variant="comment"
      />,
    );
    expect(screen.getByRole("img", { name: IMAGE.filename })).toHaveAttribute(
      "src",
      IMAGE.url,
    );
    fireEvent.click(
      screen.getByRole("button", { name: `Preview ${IMAGE.filename}` }),
    );
    const dialog = screen.getByRole("dialog");
    expect(
      within(dialog).getByRole("img", { name: IMAGE.filename }),
    ).toHaveAttribute("src", IMAGE.url);
    expect(
      within(dialog).getByRole("button", { name: /^Download$/ }),
    ).toBeInTheDocument();
    fireEvent.click(
      within(dialog).getByRole("button", { name: "View next attachment" }),
    );
    expect(
      within(dialog).getByText("Download this file to open it in its app."),
    ).toBeInTheDocument();
    fireEvent.click(
      within(dialog).getByRole("button", { name: "Download file" }),
    );
    expect(download).toHaveBeenCalledWith(WORD);
    fireEvent.click(
      within(dialog).getByRole("button", { name: "View next attachment" }),
    );
    expect(within(dialog).getByLabelText("file preview")).toHaveAttribute(
      "data",
      PDF.url,
    );
  });

  it("downloads an Office card directly and exposes full filenames without changing the gallery", () => {
    const download = jest.fn();
    render(
      <StoryAttachmentPreview
        file={WORD}
        isInChat
        onDownload={download}
        variant="comment"
      />,
    );
    const card = screen.getByRole("button", {
      name: `Download ${WORD.filename}`,
    });
    expect(card).toHaveAttribute(
      "title",
      `${WORD.filename} · Word document · 1.00 KB`,
    );
    expect(card).toHaveTextContent(WORD.filename);
    fireEvent.click(card);
    expect(download).toHaveBeenCalledWith(WORD);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("opens TXT and CSV in the local text preview", () => {
    const csv = file("results.csv", "text/csv");
    render(<StoryAttachmentPreview file={csv} isInChat variant="comment" />);
    fireEvent.click(
      screen.getByRole("button", { name: `Preview ${csv.filename}` }),
    );
    expect(
      screen.getByLabelText("results.csv text preview"),
    ).toBeInTheDocument();
  });

  it("removes an attachment without opening its preview", () => {
    const remove = jest.fn();
    render(
      <StoryAttachmentPreview
        file={IMAGE}
        isInChat
        onDelete={remove}
        variant="comment"
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: `Remove ${IMAGE.filename}` }),
    );
    expect(remove).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });
});
