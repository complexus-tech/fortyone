"use client";

import { ComponentPropsWithRef, useState } from "react";
import { EditorContent, BubbleMenu } from "@tiptap/react";

import { BubbleMenu as CustomBubbleMenu } from "./BubbleMenu";
import { cn } from "lib";

type TextEditorProps = ComponentPropsWithRef<typeof EditorContent> & {
  asTitle?: boolean;
};

export const TextEditor = ({
  editor,
  className = "",
  asTitle = false,
  ...rest
}: TextEditorProps) => {
  const [isLinkOpen, setIsLinkOpen] = useState(false);

  return (
    <>
      {editor && !asTitle && (
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
            "prose prose-lg max-w-full prose-slate leading-7 prose-a:text-primary dark:prose-invert prose-headings:font-medium prose-pre:text-dark-200 prose-pre:bg-gray-50 dark:prose-pre:bg-dark-200/80 prose-pre:text-[1.1rem]":
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
