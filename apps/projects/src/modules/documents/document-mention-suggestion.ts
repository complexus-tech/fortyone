import type { SuggestionOptions, SuggestionProps } from "@tiptap/suggestion";
import { exitSuggestion } from "@tiptap/suggestion";
import { PluginKey } from "@tiptap/pm/state";
import { ReactRenderer } from "@tiptap/react";
import type { Instance } from "tippy.js";
import tippy from "tippy.js";
import type {
  DocumentMentionItem,
  DocumentMentionListRef,
} from "./document-mention-list";
import { DocumentMentionList } from "./document-mention-list";

export const documentMentionPluginKey = new PluginKey(
  "documentMentionSuggestions",
);

export const renderDocumentMentionSuggestion: NonNullable<
  SuggestionOptions<DocumentMentionItem, DocumentMentionItem>["render"]
> = () => {
  let component: ReactRenderer<DocumentMentionListRef> | undefined;
  let popup: Instance | undefined;
  const position = (
    props: SuggestionProps<DocumentMentionItem, DocumentMentionItem>,
  ) => {
    const rect = props.clientRect?.();
    if (rect) popup?.setProps({ getReferenceClientRect: () => rect });
  };
  return {
    onStart: (props) => {
      component = new ReactRenderer(DocumentMentionList, {
        editor: props.editor,
        props,
      });
      const rect = props.clientRect?.();
      if (!rect) return;
      popup = tippy(document.body, {
        appendTo: () => document.body,
        content: component.element,
        getReferenceClientRect: () => rect,
        interactive: true,
        placement: "bottom-start",
        showOnCreate: true,
        trigger: "manual",
        // Above both the document panel and its anchored composer.
        zIndex: 100,
      });
    },
    onUpdate: (props) => {
      component?.updateProps(props);
      position(props);
    },
    onKeyDown: ({ event, view }) => {
      if (event.key === "Escape") {
        exitSuggestion(view, documentMentionPluginKey);
        return true;
      }
      return component?.ref?.onKeyDown(event) ?? false;
    },
    onExit: () => {
      popup?.destroy();
      popup = undefined;
      component?.destroy();
      component = undefined;
    },
  };
};
