/* global afterEach, beforeEach, describe, expect, it, jest -- Provided by the projects test runner. */
import type { ReactNode } from "react";
import { useEffect, useRef } from "react";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ChatAttachment, ChatMessage, ChatResource } from "./types";
import {
  ChatAttachmentPreviewProvider,
  useChatAttachmentPreview,
} from "./attachment-preview-context";
import { getMessage } from "./api";
import { chatKeys } from "./constants";

type MockDialogProps = {
  attachments: ChatAttachment[];
  selectedId: string;
  loading: boolean;
  error?: string;
  footer: ReactNode;
  onClose: () => void;
  onCloseAutoFocus: (event: Event) => void;
};

const MockPreviewDialog = ({
  attachments,
  selectedId,
  loading,
  error,
  footer,
  onClose,
  onCloseAutoFocus,
}: MockDialogProps) => {
  const restoreFocus = useRef(onCloseAutoFocus);
  const closeButton = useRef<HTMLButtonElement>(null);
  restoreFocus.current = onCloseAutoFocus;
  useEffect(() => {
    closeButton.current?.focus();
    return () => {
      restoreFocus.current(new Event("close", { cancelable: true }));
    };
  }, []);
  return (
    <div
      aria-label="Attachment viewer"
      data-selected-id={selectedId}
      role="dialog"
    >
      {loading ? <p role="status">Loading attachments</p> : null}
      {error ? <p role="alert">{error}</p> : null}
      {attachments.map((attachment) => (
        <a href={attachment.url} key={attachment.id}>
          {attachment.fileName}
        </a>
      ))}
      <button onClick={onClose} ref={closeButton} type="button">
        Close preview
      </button>
      {footer}
    </div>
  );
};

jest.mock(
  "@/components/ui/attachment-preview",
  () => ({
    AttachmentPreviewDialog: (props: MockDialogProps) => (
      <MockPreviewDialog {...props} />
    ),
  }),
  { virtual: true },
);
jest.mock("@/hooks/use-workspace-path", () => ({
  useWorkspacePath: () => ({ workspaceSlug: "second" }),
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
jest.mock("./api", () => ({ getMessage: jest.fn() }));

const attachments: ChatAttachment[] = [
  {
    id: "image",
    fileName: "Photo.png",
    mediaType: "image/png",
    sizeBytes: 1000,
    url: "https://untrusted.example/photo.png",
  },
  {
    id: "pdf",
    fileName: "Brief.pdf",
    mediaType: "application/pdf",
    sizeBytes: 2000,
    url: "",
  },
  {
    id: "voice",
    fileName: "Voice.webm",
    mediaType: "audio/webm",
    sizeBytes: 3000,
    url: "",
  },
];
const message: ChatMessage = {
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
  text: "Please review",
  richContent: null,
  createdAt: "2026-10-17T08:00:00Z",
  editedAt: null,
  deletedAt: null,
  version: 1,
  replyCount: 0,
  unreadReplyCount: 0,
  lastReplyAt: null,
  replyAuthors: [],
  reactions: [],
  attachments,
  attachmentIds: attachments.map(({ id }) => id),
  links: [],
  taskRefs: [],
  mentionedUserIds: [],
  agentMentions: [],
};
const resource: ChatResource = {
  id: "pdf",
  attachmentId: "pdf",
  conversationId: "product",
  messageId: "message",
  kind: "file",
  fileName: "Brief.pdf",
  title: "Brief.pdf",
  mediaType: "application/pdf",
  sizeBytes: 2000,
  url: "",
  description: "",
  createdAt: message.createdAt,
  author: message.author,
};
const mockedGetMessage = jest.mocked(getMessage);
const onViewMessage = jest.fn();
let client: QueryClient;

const Trigger = ({ fromResource }: { fromResource: boolean }) => {
  const { openMessage, openResource } = useChatAttachmentPreview();
  return (
    <button
      onClick={(event) => {
        if (fromResource) openResource(resource, event.currentTarget);
        else openMessage(message, "image", event.currentTarget);
      }}
      type="button"
    >
      Open attachment
    </button>
  );
};
const Fixture = ({
  visible = true,
  fromResource = false,
}: {
  visible?: boolean;
  fromResource?: boolean;
}) => (
  <QueryClientProvider client={client}>
    <ChatAttachmentPreviewProvider
      conversationName="Product"
      onViewMessage={onViewMessage}
    >
      <section aria-label="Product conversation">
        <h2 tabIndex={-1}>Product</h2>
        {visible ? <Trigger fromResource={fromResource} /> : null}
      </section>
    </ChatAttachmentPreviewProvider>
  </QueryClientProvider>
);

beforeEach(() => {
  jest.clearAllMocks();
  client = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  });
  mockedGetMessage.mockResolvedValue(message);
});
afterEach(() => {
  cleanup();
  client.clear();
});

describe("chat attachment preview sessions", () => {
  it("resolves a Files & links item to its message's gallery with authorized URLs", async () => {
    render(<Fixture fromResource />);
    fireEvent.click(screen.getByRole("button", { name: "Open attachment" }));

    await waitFor(() => {
      expect(screen.getAllByRole("link")).toHaveLength(2);
    });
    expect(mockedGetMessage).toHaveBeenCalledWith(
      { workspaceSlug: "second" },
      "message",
      expect.any(AbortSignal),
    );
    expect(screen.getByRole("dialog")).toHaveAttribute(
      "data-selected-id",
      "pdf",
    );
    expect(screen.getByRole("link", { name: "Photo.png" })).toHaveAttribute(
      "href",
      "https://api.example.test/workspaces/second/chat/conversations/product/attachments/image",
    );
    expect(screen.getByRole("link", { name: "Brief.pdf" })).toHaveAttribute(
      "href",
      "/api/chat/second/conversations/product/attachments/pdf/preview",
    );
    expect(
      screen.queryByRole("link", { name: "Voice.webm" }),
    ).not.toBeInTheDocument();
  });

  it("keeps the viewer open after its virtualized trigger unmounts and restores focus to the conversation", async () => {
    const { rerender } = render(<Fixture />);
    fireEvent.click(screen.getByRole("button", { name: "Open attachment" }));
    await waitFor(() => {
      expect(mockedGetMessage).toHaveBeenCalledTimes(1);
    });

    rerender(<Fixture visible={false} />);

    expect(
      screen.queryByRole("button", { name: "Open attachment" }),
    ).not.toBeInTheDocument();
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByRole("link", { name: "Photo.png" })).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Close preview" }));
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Product" })).toHaveFocus();
  });

  it("restores focus to a trigger that is still mounted", async () => {
    render(<Fixture />);
    const trigger = screen.getByRole("button", { name: "Open attachment" });
    fireEvent.click(trigger);
    await waitFor(() => {
      expect(mockedGetMessage).toHaveBeenCalledTimes(1);
    });
    fireEvent.click(screen.getByRole("button", { name: "Close preview" }));
    expect(trigger).toHaveFocus();
  });

  it.each(["deleted", "revoked", "different conversation"] as const)(
    "removes previously visible attachment content when the source becomes %s",
    async (condition) => {
      render(<Fixture />);
      fireEvent.click(screen.getByRole("button", { name: "Open attachment" }));
      await waitFor(() => {
        expect(screen.getAllByRole("link")).toHaveLength(2);
      });
      await waitFor(() => {
        expect(client.isFetching()).toBe(0);
      });
      if (condition === "revoked")
        mockedGetMessage.mockRejectedValue(new Error("Forbidden"));
      else
        mockedGetMessage.mockResolvedValue({
          ...message,
          ...(condition === "deleted"
            ? { deletedAt: "2026-10-17T09:00:00Z" }
            : { conversationId: "private" }),
        });

      await act(async () => {
        await client.invalidateQueries({
          queryKey: [...chatKeys.all("second"), "message", "message"],
        });
      });

      await waitFor(() => {
        expect(screen.getByRole("alert")).toBeInTheDocument();
      });
      expect(screen.queryAllByRole("link")).toHaveLength(0);
      expect(screen.getByRole("dialog")).toBeInTheDocument();
    },
  );

  it("returns to the source conversation without stealing navigation focus", async () => {
    onViewMessage.mockImplementation(() => {
      screen.getByRole("heading", { name: "Product" }).focus();
    });
    render(<Fixture fromResource />);
    fireEvent.click(screen.getByRole("button", { name: "Open attachment" }));
    await waitFor(() => {
      expect(screen.getAllByRole("link")).toHaveLength(2);
    });
    fireEvent.click(
      screen.getByRole("button", { name: "View in conversation" }),
    );

    expect(onViewMessage).toHaveBeenCalledWith("message");
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
    expect(screen.getByRole("heading", { name: "Product" })).toHaveFocus();
  });
});
