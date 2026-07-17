"use client";

import { useState } from "react";
import { EditorContent, EditorContentProps } from "@tiptap/react";
import { BubbleMenu } from "@tiptap/react/menus";

import { BubbleMenu as CustomBubbleMenu } from "./bubble-menu";
import { cn } from "lib";

type TextEditorProps = EditorContentProps & {
  asTitle?: boolean;
  hideBubbleMenu?: boolean;
};

export const TextEditor = ({
  editor,
  className = "",
  asTitle = false,
  hideBubbleMenu = false,
  innerRef,
  ...rest
}: TextEditorProps) => {
  const [isLinkOpen, setIsLinkOpen] = useState(false);

  return (
    <>
      {editor && !asTitle && !hideBubbleMenu && (
        <BubbleMenu
          editor={editor}
          options={{
            onHide: () => {
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
        className={cn(
          {
            "rich-text-editor prose prose-lg max-w-full prose-stone prose-a:text-primary dark:prose-invert prose-headings:font-medium prose-pre:text-foreground prose-pre:bg-surface-muted prose-pre:text-[1.1rem] prose-strong:font-medium prose-h1:text-3xl prose-h2:text-2xl prose-h3:text-xl prose-h4:text-lg prose-h5:text-lg prose-h6:text-lg":
              !asTitle,
            "mb-4": asTitle,
          },
          className,
        )}
        editor={editor}
        {...rest}
      />
    </>
  );
};
