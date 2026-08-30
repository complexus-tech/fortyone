"use client";

import { useEffect } from "react";
import { useEditor } from "@tiptap/react";
import Document from "@tiptap/extension-document";
import Link from "@tiptap/extension-link";
import { TaskItem, TaskList } from "@tiptap/extension-list";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import TextExtension from "@tiptap/extension-text";
import Underline from "@tiptap/extension-underline";
import { useDebouncedCallback } from "@/hooks/debounce";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";
import type {
  IntegrationRequest,
  UpdateIntegrationRequestInput,
} from "../types";

const DEBOUNCE_DELAY = 1000;

export const useRequestEditors = ({
  onUpdate,
  request,
  requestId,
}: {
  onUpdate: (payload: UpdateIntegrationRequestInput) => void;
  request?: IntegrationRequest;
  requestId: string;
}) => {
  const {
    callback: debouncedDescriptionUpdate,
    flush: flushDescriptionUpdate,
  } = useDebouncedCallback(onUpdate, DEBOUNCE_DELAY, {
    flushOnUnmount: true,
  });
  const { callback: debouncedTitleUpdate, flush: flushTitleUpdate } =
    useDebouncedCallback(onUpdate, DEBOUNCE_DELAY, {
      flushOnUnmount: true,
    });

  const descriptionEditor = useEditor({
    extensions: [
      createRichTextStarterKit(),
      Underline,
      TaskList,
      TaskItem.configure({ nested: true }),
      Link.configure({ autolink: true }),
      Placeholder.configure({ placeholder: "Add description..." }),
    ],
    content: request?.description ?? "",
    editable: request?.status === "pending",
    onUpdate: ({ editor }) => {
      debouncedDescriptionUpdate({
        description: editor.getHTML(),
      });
    },
    onBlur: flushDescriptionUpdate,
    immediatelyRender: false,
  });

  const titleEditor = useEditor({
    extensions: [
      Document,
      Paragraph,
      TextExtension,
      Placeholder.configure({ placeholder: "Add title..." }),
    ],
    content: request?.title ?? "",
    editable: request?.status === "pending",
    onUpdate: ({ editor }) => {
      debouncedTitleUpdate({
        title: editor.getText(),
      });
    },
    onBlur: flushTitleUpdate,
    immediatelyRender: false,
  });

  useEffect(() => {
    if (
      titleEditor &&
      !titleEditor.isFocused &&
      request?.title &&
      titleEditor.getText() !== request.title
    ) {
      titleEditor.commands.setContent(request.title, { emitUpdate: false });
    }
    if (
      descriptionEditor &&
      !descriptionEditor.isFocused &&
      request?.description &&
      descriptionEditor.getHTML() !== request.description
    ) {
      descriptionEditor.commands.setContent(request.description, {
        emitUpdate: false,
      });
    }
  }, [descriptionEditor, request?.description, request?.title, titleEditor]);

  useEffect(
    () => () => {
      flushDescriptionUpdate();
      flushTitleUpdate();
    },
    [flushDescriptionUpdate, flushTitleUpdate, requestId],
  );

  return { descriptionEditor, titleEditor };
};
