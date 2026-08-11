/* global describe, expect, it, jest -- Jest globals are provided by the projects test runner. */

import type { Editor } from "@tiptap/core";
import type { ComponentProps, ReactNode } from "react";
import { act, render, waitFor } from "@testing-library/react";
import type { Member } from "@/types";
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
type MentionItems = (input: { editor: Editor; query: string }) => MentionItem[];
type MentionExtensionOptions = {
  suggestion: { items: MentionItems };
};

jest.mock("ui", () => ({
  Button: ({ children, onClick }: ComponentProps<"button">) => (
    <button onClick={onClick} type="button">
      {children}
    </button>
  ),
  Flex: ({ children }: { children?: ReactNode }) => <div>{children}</div>,
  TextEditor: ({ editor }: { editor: Editor | null }) => {
    mockEditor = editor;
    return <div data-testid="comment-editor" />;
  },
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
jest.mock("@/lib/hooks/update-comment-mutation", () => ({
  useUpdateCommentMutation: () => ({ mutate: jest.fn() }),
}));
jest.mock("@/modules/story/hooks/comment-mutation", () => ({
  useCommentStoryMutation: () => ({ mutate: jest.fn() }),
}));

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
