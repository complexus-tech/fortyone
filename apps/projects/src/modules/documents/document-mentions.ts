import type { Editor } from "@tiptap/core";
import { Extension, mergeAttributes } from "@tiptap/core";
import { DocumentMention } from "@fortyone/document-editor";
import { Suggestion } from "@tiptap/suggestion";
import type { DocumentMentionItem } from "./document-mention-list";
import {
  documentMentionPluginKey,
  renderDocumentMentionSuggestion,
} from "./document-mention-suggestion";

export type DocumentMentionSearch = (
  query: string,
) => Promise<DocumentMentionItem[]>;

type MentionStorage = { search: DocumentMentionSearch };

export const DocumentMentionSuggestions = Extension.create<
  Record<string, never>,
  MentionStorage
>({
  name: "documentMentionSuggestions",
  addStorage: () => ({ search: async () => [] }),
  addProseMirrorPlugins() {
    return [
      Suggestion<DocumentMentionItem, DocumentMentionItem>({
        editor: this.editor,
        pluginKey: documentMentionPluginKey,
        char: "@",
        allowSpaces: true,
        items: ({ query }) => this.storage.search(query),
        allow: ({ state, range }) => {
          const position = state.doc.resolve(range.from);
          return (
            Boolean(
              position.parent.type.contentMatch.matchType(
                state.schema.nodes.mention,
              ),
            ) &&
            !position.parent.type.spec.code &&
            !position.marks().some((mark) => mark.type.name === "code")
          );
        },
        command: ({ editor, range, props }) => {
          if (props.kind === "message") return;
          const isReference =
            props.kind === "story" || props.kind === "objective";
          if (
            editor.state.doc.resolve(range.to).nodeAfter?.text?.startsWith(" ")
          )
            range.to += 1;
          editor
            .chain()
            .focus()
            .insertContentAt(range, [
              {
                type: isReference ? "documentReference" : "mention",
                attrs: isReference
                  ? {
                      id: props.id,
                      label: props.label,
                      entityType: props.kind,
                      displayId: props.displayId,
                    }
                  : {
                      id: props.id,
                      label: props.label,
                      mentionKind: props.kind === "agent" ? "agent" : "user",
                      agentKey: props.kind === "agent" ? "maya" : null,
                    },
              },
              { type: "text", text: " " },
            ])
            .run();
        },
        render: renderDocumentMentionSuggestion,
      }),
    ];
  },
});

export function setDocumentMentionSearch(
  editor: Editor,
  search: DocumentMentionSearch,
) {
  const storage = (
    editor.storage as unknown as { documentMentionSuggestions?: MentionStorage }
  ).documentMentionSuggestions;
  if (storage) storage.search = search;
}

const escapeMarkdownLabel = (value: string) =>
  value.replace(/[\\[\]_*`]/g, "\\$&");

// Comment identity stays in Markdown; the body editor uses the shared human-only node.
export const DocumentCommentMention = DocumentMention.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      mentionKind: {
        default: "user",
        parseHTML: (element) =>
          element.getAttribute("data-mention-kind") === "agent"
            ? "agent"
            : "user",
      },
      agentKey: {
        default: null,
        parseHTML: (element) => element.getAttribute("data-agent-key"),
      },
    };
  },
  parseHTML() {
    return [
      { tag: 'span[data-type="mention"]', priority: 1001 },
      { tag: 'a[data-type="mention"]', priority: 1001 },
    ];
  },
  renderHTML({ node, HTMLAttributes }) {
    const agent =
      node.attrs.mentionKind === "agent" && node.attrs.agentKey === "maya";
    return [
      "span",
      mergeAttributes(HTMLAttributes, {
        "data-type": "mention",
        "data-id": node.attrs.id,
        "data-label": node.attrs.label,
        "data-mention-kind": agent ? "agent" : "user",
        ...(agent ? { "data-agent-key": "maya" } : {}),
        class: "mention",
      }),
      `@${node.attrs.label || node.attrs.id}`,
    ];
  },
  renderMarkdown(node) {
    const attrs = node.attrs ?? {};
    const agent = attrs.mentionKind === "agent" && attrs.agentKey === "maya";
    return `[@${escapeMarkdownLabel(String(attrs.label ?? attrs.id ?? ""))}](${agent ? "agent:maya" : `mention:${attrs.id}`})`;
  },
});
