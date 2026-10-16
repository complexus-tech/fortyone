import type { Editor } from "@tiptap/core";
import { mergeAttributes, ReactRenderer } from "@tiptap/react";
import Mention from "@tiptap/extension-mention";
import Placeholder from "@tiptap/extension-placeholder";
import tippy from "tippy.js";
import { getCommentEditorBaseExtensions } from "@/lib/tiptap/comment-editor";
import { serializeCommentToGitHubMarkdown } from "@/lib/tiptap/comment-markdown";
import type { MentionItem, MentionListRef } from "./mentions/list";
import { MentionList } from "./mentions/list";
import {
  CommentAttachment,
  CommentAttachmentDrop,
} from "./comment-attachment-extension";

const EMPTY_MARKDOWN = "";
const MENTION_TRIGGER = "@";

const StoryCommentMention = Mention.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      mentionKind: {
        default: "user",
        parseHTML: (element) =>
          element.getAttribute("data-mention-kind") === "agent"
            ? "agent"
            : "user",
        renderHTML: (attributes) =>
          attributes.mentionKind === "agent"
            ? { "data-mention-kind": "agent" }
            : {},
      },
      agentKey: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-agent-key"),
        renderHTML: (attributes) =>
          attributes.mentionKind === "agent" && attributes.agentKey === "maya"
            ? { "data-agent-key": "maya" }
            : {},
      },
    };
  },
  parseHTML() {
    return [
      { tag: 'a[data-type="mention"]', priority: 1001 },
      { tag: 'span[data-type="mention"]', priority: 1001 },
    ];
  },
});

type MentionStorage = {
  users?: MentionItem[];
};

const getMentionStorage = (editor: Editor) =>
  (editor.storage as unknown as { mention?: MentionStorage }).mention;

const getStoryCommentMentionUsers = (editor: Editor) =>
  getMentionStorage(editor)?.users ?? [];

export const setStoryCommentMentionUsers = (
  editor: Editor,
  users: MentionItem[],
) => {
  const storage = getMentionStorage(editor);
  if (storage) storage.users = users;
};

const renderMentionSuggestion = () => {
  let component: ReactRenderer<MentionListRef>;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tippy.js instance type is complex
  let popup: any;

  return {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onStart: (props: any) => {
      component = new ReactRenderer(MentionList, {
        editor: props.editor,
        props,
      });

      if (!props.clientRect) return;

      popup = tippy("body", {
        appendTo: () => document.body,
        content: component.element,
        getReferenceClientRect: props.clientRect,
        interactive: true,
        placement: "bottom-start",
        showOnCreate: true,
        trigger: "manual",
      });
    },

    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onUpdate(props: any) {
      component.updateProps(props as Record<string, unknown>);
      if (!props.clientRect) return;

      popup?.[0]?.setProps({ getReferenceClientRect: props.clientRect });
    },

    // eslint-disable-next-line @typescript-eslint/no-explicit-any -- Tiptap suggestion props type is complex
    onKeyDown(props: any) {
      if (props.event.key === "Escape") {
        popup?.[0]?.hide();
        return true;
      }

      return component.ref?.onKeyDown(props.event as KeyboardEvent) ?? false;
    },

    onExit() {
      popup?.[0]?.destroy();
      component.destroy();
    },
  };
};

export const getStoryCommentEditorExtensions = ({
  enableMentions = false,
  enableAttachments = false,
  onFiles,
  placeholder,
}: {
  enableMentions?: boolean;
  enableAttachments?: boolean;
  onFiles?: (editor: Editor, files: File[]) => void;
  placeholder: string;
}) => {
  const mentionExtensions = enableMentions
    ? [
        StoryCommentMention.configure({
          HTMLAttributes: {
            class: "mention",
          },
          renderHTML({ node, options, suggestion }) {
            const label = node.attrs.label ?? node.attrs.id ?? EMPTY_MARKDOWN;
            const trigger = suggestion?.char ?? MENTION_TRIGGER;

            return [
              node.attrs.mentionKind === "agent" ? "span" : "a",
              mergeAttributes(
                node.attrs.mentionKind === "agent"
                  ? { title: "Maya · AI assistant" }
                  : { href: `/profile/${node.attrs.id}` },
                options.HTMLAttributes,
              ),
              `${trigger}${label}`,
            ];
          },
          suggestion: {
            items: ({ editor, query }: { editor: Editor; query: string }) => {
              const mentionUsers = getStoryCommentMentionUsers(editor);
              if (!query || query.trim() === "")
                return mentionUsers.slice(0, 6);

              return mentionUsers
                .filter(
                  (user) =>
                    user.label.toLowerCase().includes(query.toLowerCase()) ||
                    user.username.toLowerCase().includes(query.toLowerCase()),
                )
                .slice(0, 6);
            },
            render: renderMentionSuggestion,
          },
        }),
      ]
    : [];

  return [
    ...getCommentEditorBaseExtensions(),
    ...(enableAttachments
      ? [CommentAttachment, CommentAttachmentDrop.configure({ onFiles })]
      : []),
    ...mentionExtensions,
    Placeholder.configure({ placeholder }),
  ];
};

// Compatibility export for existing Story consumers; new shared consumers use the lib contract.
export const serializeStoryCommentToGitHubMarkdown =
  serializeCommentToGitHubMarkdown;
