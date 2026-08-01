"use client";

import { useEffect, useRef } from "react";
import Link from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import Underline from "@tiptap/extension-underline";
import { useEditor } from "@tiptap/react";
import { cn } from "lib";
import { TextEditor } from "ui";
import { createRichTextStarterKit } from "@/lib/tiptap/starter-kit";

type StrategyDescriptionEditorProps = {
  ariaLabel: string;
  className?: string;
  content: string;
  contentClassName?: string;
  editable: boolean;
  onBlur?: () => void;
  onChange: (description: string | null) => void;
  placeholder?: string;
};

export const StrategyDescriptionEditor = ({
  ariaLabel,
  className,
  content,
  contentClassName,
  editable,
  onBlur,
  onChange,
  placeholder = "Add a description...",
}: StrategyDescriptionEditorProps) => {
  const onBlurRef = useRef(onBlur);
  const onChangeRef = useRef(onChange);

  useEffect(() => {
    onBlurRef.current = onBlur;
    onChangeRef.current = onChange;
  }, [onBlur, onChange]);

  const editor = useEditor({
    content,
    editable,
    editorProps: {
      attributes: {
        "aria-label": ariaLabel,
        class: cn("min-h-24", contentClassName),
      },
    },
    extensions: [
      createRichTextStarterKit(),
      Underline,
      Link.configure({
        autolink: true,
      }),
      Placeholder.configure({ placeholder }),
    ],
    immediatelyRender: false,
    onBlur: () => {
      onBlurRef.current?.();
    },
    onUpdate: ({ editor: currentEditor }) => {
      onChangeRef.current(
        currentEditor.isEmpty ? null : currentEditor.getHTML(),
      );
    },
  });

  useEffect(() => {
    editor?.setEditable(editable);
  }, [editable, editor]);

  return (
    <TextEditor
      className={cn("text-foreground", className)}
      editor={editor}
      hideBubbleMenu={!editable}
    />
  );
};
