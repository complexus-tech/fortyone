/* global beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
/* eslint-disable @next/next/no-img-element -- Native image mock exposes load events without Next.js layout and decoding behavior in JSDOM. */
import { fireEvent, render, screen, within } from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { AudioPlaybackProvider } from "./audio-playback-context";
import type { ChatAttachment, ChatMessage } from "./types";
import { MessageAttachments } from "./message-attachments";

const mockOpenMessage = jest.fn();
jest.mock("@/components/ui/attachment-preview/cached-thumbnail", () => ({
  CachedAttachmentThumbnail: ({
    url,
    fileName,
    imageClassName,
    onImageLoad,
  }: {
    url: string;
    fileName: string;
    imageClassName: string;
    onImageLoad?: (width: number, height: number) => void;
  }) => (
    <img
      alt={fileName}
      src={url}
      className={imageClassName}
      loading="lazy"
      decoding="async"
      onLoad={(event) =>
        onImageLoad?.(
          event.currentTarget.naturalWidth,
          event.currentTarget.naturalHeight,
        )
      }
    />
  ),
}));
jest.mock(
  "./attachment-preview-context",
  () => ({
    useChatAttachmentPreview: () => ({ openMessage: mockOpenMessage }),
  }),
  { virtual: true },
);
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
jest.mock("@/components/ui/attachment-preview/pdf-thumbnail", () => ({
  PdfThumbnail: ({ url, fileName }: { url: string; fileName: string }) => (
    <span
      aria-label={`First page of ${fileName}`}
      data-preview-url={url}
      role="img"
    />
  ),
}));

const attachment = (
  id: string,
  mediaType = "image/png",
  fileName = `${id}.png`,
): ChatAttachment => ({ id, mediaType, fileName, sizeBytes: 1024, url: "" });
const messageWith = (attachments: ChatAttachment[]): ChatMessage => ({
  id: "message",
  conversationId: "product",
  threadRootId: null,
  creationId: "creation",
  sequence: 1,
  author: {
    id: "author",
    fullName: "Joseph",
    avatarUrl: null,
    isSystem: false,
  },
  text: "Files for review",
  richContent: null,
  createdAt: "2026-10-17T08:00:00Z",
  editedAt: null,
  deletedAt: null,
  version: 1,
  replyCount: 0,
  lastReplyAt: null,
  replyAuthors: [],
  reactions: [],
  attachments,
  attachmentIds: attachments.map(({ id }) => id),
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
});

beforeEach(() => {
  jest.clearAllMocks();
});

describe("message attachment layouts", () => {
  it("loads only four images for an eight-image gallery and opens the clicked image", () => {
    const message = messageWith(
      Array.from({ length: 8 }, (_, index) => attachment(`image-${index + 1}`)),
    );
    render(<MessageAttachments message={message} />);

    const images = screen.getAllByRole("img");
    expect(images).toHaveLength(4);
    expect(screen.getByText("+4")).toBeInTheDocument();
    images.forEach((image) => {
      expect(image).toHaveAttribute("loading", "lazy");
      expect(image).toHaveAttribute("decoding", "async");
    });
    const fourth = screen.getByRole("button", {
      name: "Preview image-4.png and 4 more images",
    });
    fireEvent.click(fourth);
    expect(mockOpenMessage).toHaveBeenCalledWith(message, "image-4", fourth);
    expect(
      screen.queryByRole("img", { name: "image-5.png" }),
    ).not.toBeInTheDocument();
  });

  it("fits a single portrait frame to its natural aspect ratio after loading", () => {
    render(
      <MessageAttachments message={messageWith([attachment("portrait")])} />,
    );
    const image = screen.getByRole("img", { name: "portrait.png" });
    expect(image).toHaveClass("object-contain");
    expect(image).toHaveAttribute(
      "src",
      "/api/chat/second/conversations/product/attachments/portrait/thumbnail",
    );
    Object.defineProperties(image, {
      naturalWidth: { value: 600 },
      naturalHeight: { value: 1200 },
    });
    fireEvent.load(image);
    const frame = screen.getByRole("button", { name: "Preview portrait.png" });
    expect(frame).toHaveStyle({
      width: "180px",
      "max-width": "100%",
    });
    expect(frame.style.aspectRatio).toBe("600 / 1200");
    expect(frame).toHaveClass("max-h-[360px]", "max-w-[480px]");
  });

  it.each([1, 2])(
    "shows first-page thumbnails for %i PDFs through authenticated URLs with one preview control per file",
    (count) => {
      const files = Array.from({ length: count }, (_, index) =>
        attachment(
          `pdf-${index + 1}`,
          "application/pdf",
          `Brief ${index + 1}.pdf`,
        ),
      );
      const message = messageWith(files);
      render(<MessageAttachments message={message} />);
      expect(
        screen.getAllByRole("img", { name: /First page of/ }),
      ).toHaveLength(count);
      expect(screen.getAllByRole("button")).toHaveLength(count);
      expect(mockOpenMessage).not.toHaveBeenCalled();

      files.forEach((file) => {
        const button = screen.getByRole("button", {
          name: `Preview ${file.fileName}`,
        });
        const thumbnail = within(button).getByRole("img", {
          name: `First page of ${file.fileName}`,
        });
        expect(thumbnail).toHaveAttribute(
          "data-preview-url",
          `/api/chat/second/conversations/product/attachments/${file.id}/preview`,
        );
        expect(within(button).getByText("PDF file · 1 KB")).toBeInTheDocument();
        expect(button.querySelector("button")).toBeNull();
        fireEvent.click(thumbnail);
        expect(mockOpenMessage).toHaveBeenLastCalledWith(
          message,
          file.id,
          button,
        );
      });
    },
  );

  it("keeps three PDFs compact without mounting first-page renderers", () => {
    const message = messageWith(
      Array.from({ length: 3 }, (_, index) =>
        attachment(
          `pdf-${index + 1}`,
          "application/pdf",
          `Brief ${index + 1}.pdf`,
        ),
      ),
    );
    render(<MessageAttachments message={message} />);
    expect(
      screen.queryByRole("img", { name: /First page of/ }),
    ).not.toBeInTheDocument();
    const buttons = screen.getAllByRole("button");
    expect(buttons).toHaveLength(3);
    buttons.forEach((button, index) => {
      fireEvent.click(button);
      expect(mockOpenMessage).toHaveBeenLastCalledWith(
        message,
        message.attachments[index].id,
        button,
      );
    });
  });

  it("keeps multiple document cards compact and selects each exact attachment", () => {
    const message = messageWith([
      attachment("pdf", "application/pdf", "Brief.pdf"),
      attachment("text", "text/plain", "Notes.txt"),
    ]);
    render(<MessageAttachments message={message} />);
    expect(
      screen.queryByRole("img", { name: /First page of/ }),
    ).not.toBeInTheDocument();
    const documents = screen.getAllByRole("button");
    expect(documents).toHaveLength(2);
    documents.forEach((button, index) => {
      fireEvent.click(button);
      expect(mockOpenMessage).toHaveBeenLastCalledWith(
        message,
        message.attachments[index].id,
        button,
      );
    });
  });

  it("offers inline audio playback without creating or preloading media before play", () => {
    const AudioConstructor = jest.spyOn(window, "Audio");
    const { container } = render(
      <QueryClientProvider client={new QueryClient()}>
        <AudioPlaybackProvider>
          <MessageAttachments
            message={messageWith([
              attachment(
                "voice",
                "audio/webm;codecs=opus",
                "Voice message.webm",
              ),
            ])}
          />
        </AudioPlaybackProvider>
      </QueryClientProvider>,
    );
    expect(
      screen.getByRole("button", { name: "Play Voice message.webm" }),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Transcribe" }),
    ).toBeInTheDocument();
    expect(container.querySelector("audio")).toBeNull();
    expect(AudioConstructor).not.toHaveBeenCalled();
    expect(mockOpenMessage).not.toHaveBeenCalled();
    AudioConstructor.mockRestore();
  });

  it("uses a video placeholder that opens the shared player on demand", () => {
    const message = messageWith([
      attachment("video", "video/mp4", "Walkthrough.mp4"),
    ]);
    const { container } = render(<MessageAttachments message={message} />);
    expect(container.querySelector("video")).toBeNull();
    const button = screen.getByRole("button", { name: "Play Walkthrough.mp4" });
    fireEvent.click(button);
    expect(mockOpenMessage).toHaveBeenCalledWith(message, "video", button);
  });

  it("keeps an image accessible through its independent full-preview button", () => {
    const message = messageWith([attachment("image")]);
    render(<MessageAttachments message={message} />);
    // The thumbnail component owns loading/error rendering. Its surrounding
    // preview control must remain available independently of thumbnail state.
    fireEvent.error(screen.getByRole("img"));
    const button = screen.getByRole("button", { name: "Preview image.png" });
    fireEvent.click(button);
    expect(mockOpenMessage).toHaveBeenCalledWith(message, "image", button);
  });

  it("does not mount attachments for deleted messages", () => {
    const message = {
      ...messageWith([attachment("image")]),
      deletedAt: "2026-10-17T09:00:00Z",
    };
    const { container } = render(<MessageAttachments message={message} />);
    expect(container).toBeEmptyDOMElement();
  });
});
