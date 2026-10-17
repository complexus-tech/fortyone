import type { Editor } from "@tiptap/core";
import type { SuggestionProps } from "@tiptap/suggestion";
import { exitSuggestion, SuggestionPluginKey } from "@tiptap/suggestion";
import type { Instance } from "tippy.js";
import Mention from "@tiptap/extension-mention";
import { mergeAttributes, ReactRenderer } from "@tiptap/react";
import tippy from "tippy.js";
import { toast } from "sonner";
import { MentionList } from "./mention-list";
import type { MentionItem, MentionListRef } from "./mention-list";

const isAborted = (signal: AbortSignal) => signal.aborted;

type MentionStorage = {
  people?: MentionItem[];
  search?: (query: string, signal: AbortSignal) => Promise<MentionItem[]>;
  pending?: AbortController;
};
export const setChatMentionSearch = (
  editor: Editor,
  search: MentionStorage["search"],
) => {
  const current = storage(editor);
  if (current) current.search = search;
};
const storage = (editor: Editor) =>
  (editor.storage as unknown as { mention?: MentionStorage }).mention;
export const setChatMentionPeople = (editor: Editor, people: MentionItem[]) => {
  const current = storage(editor);
  if (current) current.people = people;
};

export const ChatMention = Mention.extend({
  addAttributes() {
    return {
      ...this.parent?.(),
      mentionKind: {
        default: "user",
        renderHTML: (attrs) => ({ "data-mention-kind": attrs.mentionKind }),
      },
      agentKey: {
        default: null,
        renderHTML: (attrs) =>
          attrs.agentKey === "maya" ? { "data-agent-key": "maya" } : {},
      },
    };
  },
}).configure({
  HTMLAttributes: {
    class: "bg-primary/10 text-primary rounded px-1 py-0.5 font-medium",
  },
  renderHTML({ node, options }) {
    return [
      "span",
      mergeAttributes(options.HTMLAttributes),
      `@${String(node.attrs.label ?? "Member")}`,
    ];
  },
  suggestion: {
    pluginKey: SuggestionPluginKey,
    items: async ({ editor, query }) => {
      const current = storage(editor);
      if (!current?.search)
        return (current?.people ?? [])
          .filter((person) =>
            person.label.toLowerCase().includes(query.toLowerCase()),
          )
          .slice(0, 8);
      current.pending?.abort();
      const request = new AbortController();
      current.pending = request;
      await new Promise<void>((resolve) => {
        const timer = setTimeout(resolve, 180);
        request.signal.addEventListener(
          "abort",
          () => {
            clearTimeout(timer);
            resolve();
          },
          { once: true },
        );
      });
      if (request.signal.aborted) return [];
      try {
        return await current.search(query, request.signal);
      } catch (error) {
        if (!isAborted(request.signal))
          toast.error("Could not find conversation members", {
            description:
              error instanceof Error ? error.message : "Please try again.",
          });
        return [];
      }
    },
    render: () => {
      let renderer: ReactRenderer<MentionListRef> | undefined;
      let popup: Instance | undefined;
      return {
        onStart: (props: SuggestionProps<MentionItem>) => {
          renderer = new ReactRenderer(MentionList, {
            editor: props.editor,
            props,
          });
          if (!props.clientRect) return;
          popup = tippy(document.body, {
            appendTo: () => document.body,
            content: renderer.element,
            getReferenceClientRect: () => props.clientRect?.() ?? new DOMRect(),
            interactive: true,
            placement: "bottom-start",
            showOnCreate: true,
            trigger: "manual",
          });
        },
        onUpdate: (props: SuggestionProps<MentionItem>) => {
          renderer?.updateProps(props);
          popup?.setProps({
            getReferenceClientRect: () => props.clientRect?.() ?? new DOMRect(),
          });
        },
        onKeyDown: ({ event, view }) => {
          if (event.key === "Escape") {
            exitSuggestion(view, SuggestionPluginKey);
            popup?.hide();
            return true;
          }
          return renderer?.ref?.onKeyDown(event) ?? false;
        },
        onExit: () => {
          popup?.destroy();
          renderer?.destroy();
        },
      };
    },
  },
});
