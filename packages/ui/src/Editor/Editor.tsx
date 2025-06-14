"use client";

import { useState } from "react";
import { EditorContent, BubbleMenu, EditorContentProps } from "@tiptap/react";

import { BubbleMenu as CustomBubbleMenu } from "./BubbleMenu";
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
        className={cn(
          {
            "prose prose-lg max-w-full prose-stone leading-7 prose-a:text-primary dark:prose-invert prose-headings:font-medium prose-pre:text-dark-200 prose-pre:bg-gray-50 dark:prose-pre:text-gray-200 dark:prose-pre:bg-dark-200/80 prose-pre:text-[1.1rem] prose-strong:font-medium prose-h1:text-3xl prose-h2:text-2xl prose-h3:text-xl prose-h4:text-lg prose-h5:text-lg prose-h6:text-lg":
              !asTitle,
          },
          className
        )}
        editor={editor}
        {...rest}
      />
    </>
  );
};
