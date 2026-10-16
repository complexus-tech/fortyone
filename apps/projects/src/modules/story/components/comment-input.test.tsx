/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import { randomUUID } from "node:crypto";
import type { Editor } from "@tiptap/core";
import type { ComponentProps, ReactNode } from "react";
import {
  act,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import type { Member } from "@/types";
import { DesktopWorkspaceScope } from "@/shell/desktop/workspace-scope";
import { CommentInput } from "./comment-input";
import type { MentionItem } from "./mentions/list";

const MEMBER: Member = {
  avatarUrl: null,
  createdAt: "2026-08-11T00:00:00.000Z",
  email: "thulani@example.com",
  fullName: "Thulani Museta",
  id: "7d3a62e4-a7cc-4489-99b6-a986195525c5",
  isActive: true,
  isInternal: false,
  isSystem: false,
  role: "member",
  updatedAt: "2026-08-11T00:00:00.000Z",
  username: "thulani",
};

let mockEditor: Editor | null = null;
let mockMembers: Member[] = [];
let mockMaya: Member | undefined;
let mockMayaEnabled = false;
const mockCreateComment = jest.fn();
const mockUpdateComment = jest.fn();
let mockBlockingUploads = false;
type MentionItems = (input: { editor: Editor; query: string }) => MentionItem[];
type MentionExtensionOptions = {
  suggestion: { items: MentionItems };
};

jest.mock("ui", () => ({
  Button: ({ children, onClick, disabled }: ComponentProps<"button">) => (
    <button disabled={disabled} onClick={onClick} type="button">
      {children}
    </button>
  ),
  Flex: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  Box: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  Text: ({ children }: { children?: ReactNode }) => <span>{children}</span>,
  TextEditor: ({ editor }: { editor: Editor | null }) => {
    mockEditor = editor;
    return <div data-testid="comment-editor" />;
  },
}));
jest.mock("./comment-attachment-list", () => ({
  CommentAttachmentList: () => null,
}));
jest.mock("@/lib/api-url", () => ({
  getApiUrl: () => "https://api.example.test",
}));
jest.mock("@/modules/story/hooks/use-comment-attachments", () => ({
  COMMENT_ATTACHMENT_ACCEPT: "image/png,application/pdf",
  useCommentAttachments: () => ({
    addFiles: jest.fn(),
    hasBlockingUploads: mockBlockingUploads,
    hasPendingUploads: () => mockBlockingUploads,
    isUploading: mockBlockingUploads,
    remove: jest.fn(),
    retry: jest.fn(),
    uploads: [],
  }),
}));
jest.mock("lib", () => ({
  cn: (...values: unknown[]) =>
    values.filter((value) => typeof value === "string").join(" "),
}));
jest.mock("sonner", () => ({ toast: { error: jest.fn() } }));
jest.mock("@/lib/auth/client", () => ({
  useSession: () => ({ data: null }),
}));
jest.mock("@/lib/hooks/team-members", () => ({
  useTeamMembers: () => ({ data: mockMembers }),
}));
jest.mock("@/lib/hooks/members", () => ({
  useMayaAssignee: () => ({ data: mockMaya }),
}));
jest.mock("../hooks/comment-maya", () => ({
  useStoryMayaRuns: () => ({ data: { enabled: mockMayaEnabled, runs: [] } }),
}));
jest.mock("@/lib/hooks/update-comment-mutation", () => ({
  useUpdateCommentMutation: () => ({
    mutateAsync: mockUpdateComment,
    isPending: false,
  }),
}));
jest.mock("@/modules/story/hooks/comment-mutation", () => ({
  useCommentStoryMutation: () => ({
    mutateAsync: mockCreateComment,
    isPending: false,
  }),
}));

beforeEach(() => {
  Object.defineProperty(window.crypto, "randomUUID", {
    configurable: true,
    value: randomUUID,
  });
  jest.clearAllMocks();
  mockCreateComment.mockReset();
  mockUpdateComment.mockReset();
  mockEditor = null;
  mockMembers = [];
  mockMaya = undefined;
  mockMayaEnabled = false;
  mockBlockingUploads = false;
  window.localStorage.clear();
});

const createDeferred = () => {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((resolvePromise, rejectPromise) => {
    resolve = resolvePromise;
    reject = rejectPromise;
  });
  return { promise, resolve, reject };
};

const renderComposer = (
  props: Partial<ComponentProps<typeof CommentInput>> = {},
) =>
  render(
    <DesktopWorkspaceScope
      scope={{
        accountId: "account",
        workspaceId: "workspace",
        workspaceSlug: "acme",
      }}
    >
      <CommentInput storyId="story" teamId="team" {...props} />
    </DesktopWorkspaceScope>,
  );

const writeComment = async (html: string) => {
  await waitFor(() => {
    expect(mockEditor?.isEditable).toBe(true);
  });
  act(() => {
    mockEditor!.commands.setContent(html);
  });
};

const getMentionItems = (editor: Editor) => {
  const mentionExtension = editor.extensionManager.extensions.find(
    (extension) => extension.name === "mention",
  );
  const mentionOptions = mentionExtension?.options as
    | MentionExtensionOptions
    | undefined;
  if (!mentionOptions) {
    throw new Error("Mention extension is not configured");
  }

  return mentionOptions.suggestion.items;
};

describe("comment input mentions", () => {
  it("offers canonical Maya separately and preserves an agent request identity after failure and draft recovery", async () => {
    mockMaya = {
      ...MEMBER,
      id: "maya-id",
      isSystem: true,
      username: "maya",
      fullName: "Maya",
    };
    mockMayaEnabled = true;
    mockMembers = [mockMaya, { ...MEMBER, fullName: "Maya Chen" }];
    mockCreateComment.mockRejectedValueOnce(new Error("Lost response"));
    const first = renderComposer();
    await writeComment(
      '<p><span data-type="mention" data-mention-kind="agent" data-agent-key="maya" data-id="maya-id" data-label="Maya">@Maya</span> update the description</p>',
    );
    const draft = mockEditor!.getHTML();
    expect(
      getMentionItems(mockEditor!)({ editor: mockEditor!, query: "maya" }),
    ).toEqual([
      expect.objectContaining({
        id: "maya-id",
        mentionKind: "agent",
        agentKey: "maya",
      }),
      expect.objectContaining({ id: MEMBER.id, label: "Maya Chen" }),
    ]);
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    await waitFor(() => {
      expect(mockCreateComment).toHaveBeenCalledTimes(1);
    });
    await waitFor(() => {
      expect(mockEditor?.isEditable).toBe(true);
    });
    const payload = mockCreateComment.mock.calls[0][0].payload;
    expect(payload).toEqual({
      comment: draft,
      parentId: null,
      mentions: [],
      agentMentions: ["maya"],
      creationId: expect.any(String),
    });
    first.unmount();
    renderComposer();
    await waitFor(() => {
      expect(mockEditor?.getHTML()).toBe(draft);
    });
    await waitFor(() => {
      expect(mockEditor?.isEditable).toBe(true);
    });
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    await waitFor(() => {
      expect(mockCreateComment).toHaveBeenCalledTimes(2);
    });
    expect(mockCreateComment.mock.calls[1][0].payload).toEqual(payload);
  });

  it("restores an ordinary answer without inserting a mention and gives the reply a durable identity", async () => {
    const first = renderComposer({ parentId: "maya-question" });
    await writeComment("<p>Update only the empty-state copy.</p>");
    first.unmount();
    renderComposer({ parentId: "maya-question" });
    await waitFor(() => {
      expect(mockEditor?.getHTML()).toBe(
        "<p>Update only the empty-state copy.</p>",
      );
    });
    fireEvent.click(screen.getByRole("button", { name: "Reply" }));
    await waitFor(() => {
      expect(mockCreateComment).toHaveBeenCalledWith({
        storyId: "story",
        payload: {
          comment: "<p>Update only the empty-state copy.</p>",
          parentId: "maya-question",
          mentions: [],
          creationId: expect.any(String),
        },
      });
    });
  });

  it("updates suggestions after members load without resetting the editor", async () => {
    mockMembers = [];
    const props = {
      storyId: "0a38e06d-cc2a-4f09-b4e5-b63395ff7a6b",
      teamId: "26510fea-84d0-44c3-aa74-40e98bd4aa6b",
    };
    const { rerender } = render(<CommentInput {...props} />);

    await waitFor(() => {
      expect(mockEditor).not.toBeNull();
    });

    const initialEditor = mockEditor!;
    act(() => {
      initialEditor.commands.setContent("<p>Draft survives</p>");
    });
    const draft = initialEditor.getHTML();

    expect(
      getMentionItems(initialEditor)({ editor: initialEditor, query: "" }),
    ).toEqual([]);

    mockMembers = [MEMBER];
    rerender(<CommentInput {...props} />);

    await waitFor(() => {
      expect(
        getMentionItems(initialEditor)({
          editor: initialEditor,
          query: "thul",
        }),
      ).toEqual([
        {
          avatar: MEMBER.avatarUrl,
          id: MEMBER.id,
          label: MEMBER.fullName,
          username: MEMBER.username,
        },
      ]);
    });

    expect(mockEditor).toBe(initialEditor);
    expect(initialEditor.getHTML()).toBe(draft);
  });
});

describe("comment submission and drafts", () => {
  it("blocks sending while an attachment is pending or waiting for retry", async () => {
    mockBlockingUploads = true;
    renderComposer();
    await writeComment("<p>With a file</p>");
    const button = screen.getByRole("button", { name: "Comment" });
    expect(button).toBeDisabled();
    fireEvent.click(button);
    expect(mockCreateComment).not.toHaveBeenCalled();
    expect(mockEditor?.getHTML()).toBe("<p>With a file</p>");
  });

  it("restores and submits an attachment-only draft without losing file metadata", async () => {
    const first = renderComposer();
    await writeComment("");
    act(() => {
      mockEditor!.commands.insertContent({
        type: "commentAttachment",
        attrs: {
          id: "7d3a62e4-a7cc-4489-99b6-a986195525c5",
          filename: "review.pdf",
          mimeType: "application/pdf",
          size: 128,
          url: "https://api.example.test/workspaces/acme/stories/story/attachments/7d3a62e4-a7cc-4489-99b6-a986195525c5",
          createdAt: "2026-10-16T12:00:00Z",
          uploadedBy: "author",
        },
      });
    });
    const html = mockEditor!.getHTML();
    first.unmount();
    renderComposer();
    await waitFor(() => {
      expect(mockEditor?.getHTML()).toBe(html);
      expect(mockEditor?.isEditable).toBe(true);
    });
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    await waitFor(() => {
      expect(mockCreateComment).toHaveBeenCalledWith({
        storyId: "story",
        payload: { comment: html, parentId: null, mentions: [] },
      });
    });
  });

  it("blocks duplicate submissions and editing until the request settles", async () => {
    const request = createDeferred();
    mockCreateComment.mockReturnValue(request.promise);
    const onCancel = jest.fn();
    renderComposer({ parentId: "parent", onCancel });
    await writeComment("<p>A reply</p>");

    const button = screen.getByRole("button", { name: "Reply" });
    act(() => {
      fireEvent.click(button);
      fireEvent.click(button);
    });

    await waitFor(() => {
      expect(mockCreateComment).toHaveBeenCalledTimes(1);
    });
    expect(mockCreateComment).toHaveBeenCalledWith({
      storyId: "story",
      payload: {
        comment: "<p>A reply</p>",
        parentId: "parent",
        mentions: [],
        creationId: expect.any(String),
      },
    });
    expect(button).toBeDisabled();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeDisabled();
    expect(mockEditor?.isEditable).toBe(false);

    await act(async () => {
      request.resolve();
    });
    expect(mockEditor?.isEmpty).toBe(true);
    expect(button).toBeEnabled();
    expect(onCancel).toHaveBeenCalledTimes(1);
    expect(window.localStorage.length).toBe(0);
  });

  it("retains a failed edit and makes it available for retry", async () => {
    const request = createDeferred();
    mockUpdateComment.mockReturnValue(request.promise);
    const onCancel = jest.fn();
    renderComposer({
      commentId: "comment",
      initialComment: "<p>Original</p>",
      onCancel,
    });
    await writeComment("<p>Revised</p>");

    fireEvent.click(screen.getByRole("button", { name: "Update" }));
    expect(mockUpdateComment).toHaveBeenCalledWith({
      commentId: "comment",
      storyId: "story",
      payload: { content: "<p>Revised</p>", mentions: [] },
    });
    await act(async () => {
      request.reject(new Error("Offline"));
    });

    expect(mockEditor?.getHTML()).toBe("<p>Revised</p>");
    expect(mockEditor?.isEditable).toBe(true);
    expect(screen.getByRole("button", { name: "Update" })).toBeEnabled();
    expect(onCancel).not.toHaveBeenCalled();
    expect(window.localStorage.length).toBe(1);
  });

  it("does not clear newer content when a pending request succeeds", async () => {
    const request = createDeferred();
    mockCreateComment.mockReturnValue(request.promise);
    renderComposer();
    await writeComment("<p>Submitted</p>");
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));

    act(() => {
      mockEditor!.commands.setContent("<p>Newer draft</p>");
    });
    await act(async () => {
      request.resolve();
    });

    expect(mockEditor?.getHTML()).toBe("<p>Newer draft</p>");
    expect(mockEditor?.isEditable).toBe(true);
    expect(window.localStorage.getItem(window.localStorage.key(0)!)).toContain(
      "Newer draft",
    );
  });

  it("restores a browser draft after the comment panel unmounts", async () => {
    const first = renderComposer();
    await writeComment("<p>Unsent browser draft</p>");
    first.unmount();

    renderComposer();
    await waitFor(() => {
      expect(mockEditor?.getHTML()).toBe("<p>Unsent browser draft</p>");
      expect(mockEditor?.isEditable).toBe(true);
    });
  });

  it("keeps a reopened composer's newer draft when the old submission finishes", async () => {
    const request = createDeferred();
    mockCreateComment.mockReturnValue(request.promise);
    const first = renderComposer();
    await writeComment("<p>Submitted</p>");
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    first.unmount();

    renderComposer();
    await writeComment("<p>Written after reopening</p>");
    await act(async () => {
      request.resolve();
    });

    expect(mockEditor?.getHTML()).toBe("<p>Written after reopening</p>");
    expect(window.localStorage.getItem(window.localStorage.key(0)!)).toContain(
      "Written after reopening",
    );
  });

  it("removes the submitted browser draft even when the composer unmounts", async () => {
    const request = createDeferred();
    mockCreateComment.mockReturnValue(request.promise);
    const first = renderComposer();
    await writeComment("<p>Submitted</p>");
    fireEvent.click(screen.getByRole("button", { name: "Comment" }));
    first.unmount();
    await act(async () => {
      request.resolve();
    });

    expect(window.localStorage.length).toBe(0);
  });

  it("keeps drafts isolated when a mounted story view switches stories", async () => {
    const scope = {
      accountId: "account",
      workspaceId: "workspace",
      workspaceSlug: "acme",
    };
    const composer = (storyId: string) => (
      <DesktopWorkspaceScope scope={scope}>
        <CommentInput storyId={storyId} teamId="team" />
      </DesktopWorkspaceScope>
    );
    const { rerender } = render(composer("first-story"));
    await writeComment("<p>First story draft</p>");
    const firstEditor = mockEditor;

    rerender(composer("second-story"));
    await waitFor(() => {
      expect(mockEditor).not.toBe(firstEditor);
      expect(mockEditor?.isEditable).toBe(true);
      expect(mockEditor?.isEmpty).toBe(true);
    });
    await writeComment("<p>Second story draft</p>");

    rerender(composer("first-story"));
    await waitFor(() => {
      expect(mockEditor?.getHTML()).toBe("<p>First story draft</p>");
      expect(mockEditor?.isEditable).toBe(true);
    });
  });
});
