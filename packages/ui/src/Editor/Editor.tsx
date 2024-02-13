"use client";

import { useState } from "react";
import { EditorContent, BubbleMenu, Editor } from "@tiptap/react";
import StarterKit from "@tiptap/starter-kit";
import UnderlineExt from "@tiptap/extension-underline";
import TaskItem from "@tiptap/extension-task-item";
import TaskList from "@tiptap/extension-task-list";
import LinkExt from "@tiptap/extension-link";
import Placeholder from "@tiptap/extension-placeholder";
import { BubbleMenu as CustomBubbleMenu } from "./BubbleMenu";

export const TextEditor = ({
  editor,
  placeholder = "",
}: {
  editor: Editor | null;
  placeholder?: string;
}) => {
  const extensions = [
    StarterKit,
    UnderlineExt,
    TaskItem,
    TaskList,
    LinkExt,
    Placeholder.configure({ placeholder }),
  ];
  extensions.forEach((ext) => {
    editor?.extensionManager.extensions.push(ext);
  });
  const [isLinkOpen, setIsLinkOpen] = useState(false);

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
        className="prose prose-lg max-w-full prose-slate leading-7 prose-a:text-primary dark:prose-invert prose-headings:font-medium prose-pre:text-dark-200 prose-pre:bg-gray-50 dark:prose-pre:bg-dark-200/80 prose-pre:text-[1.1rem]"
        editor={editor}
        placeholder="Issue description"
      />
    </>
  );
};
