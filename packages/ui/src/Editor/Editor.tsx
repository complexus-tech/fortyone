"use client";

import { EditorContent, useEditor, BubbleMenu, Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import Underline from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import Link from "@tiptap/extension-link";
import { BubbleMenu as CustomBubbleMenu } from "./BubbleMenu";
import { useState } from "react";

export const TextEditor = ({ content = "" }: { content?: string }) => {
  const [isLinkOpen, setIsLinkOpen] = useState(false);
  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      TaskList,
      TaskItem.configure({
        nested: true,
      }),
      Link.configure({
        autolink: true,
      }),
    ],
    content,
    editable: true,
  }) as Editor;
  return (
    <>
      {editor && (
        <BubbleMenu
          editor={editor}
          tippyOptions={{
            duration: 100,
            moveTransition: "transform 0.15s ease-out",
            onHidden: () => {
              setIsLinkOpen(false);
            },
          }}
        >
          <CustomBubbleMenu
            editor={editor}
            isLinkOpen={isLinkOpen}
            setIsLinkOpen={setIsLinkOpen}
          />
        </BubbleMenu>
      )}
      <EditorContent
        className="prose prose-lg prose-slate leading-7 prose-a:text-primary dark:prose-invert prose-headings:font-medium dark:prose-pre:bg-dark-200/80 prose-pre:text-lg"
        editor={editor}
        placeholder="Issue description"
      />
    </>
  );
};
