"use client";

import type { ComponentProps } from "react";
import { useCallback, useState } from "react";
import { TextSelection } from "@tiptap/pm/state";
import { EditorContent, EditorContentProps } from "@tiptap/react";
import { BubbleMenu } from "@tiptap/react/menus";

import {
  BubbleMenu as CustomBubbleMenu,
  type BubbleMenuCreateAction,
  type BubbleMenuPanel,
} from "./bubble-menu";
import { cn } from "lib";

type TextEditorProps = EditorContentProps & {
  asTitle?: boolean;
  bubbleMenuCreateActions?: readonly BubbleMenuCreateAction[];
  bubbleMenuShouldShow?: ComponentProps<typeof BubbleMenu>["shouldShow"];
  hideBubbleMenu?: boolean;
};

export const TextEditor = ({
  editor,
  className = "",
  asTitle = false,
  bubbleMenuCreateActions,
  bubbleMenuShouldShow,
  hideBubbleMenu = false,
  innerRef,
  ...rest
}: TextEditorProps) => {
  const [activeBubbleMenu, setActiveBubbleMenu] =
    useState<BubbleMenuPanel>(null);
  const shouldShowBubbleMenu = useCallback<
    NonNullable<ComponentProps<typeof BubbleMenu>["shouldShow"]>
  >(
    (props) => {
      const { doc, selection } = props.state;
      const isNonEmptyTextSelection =
        selection instanceof TextSelection &&
        !selection.empty &&
        doc.textBetween(props.from, props.to).length > 0;
      const isMenuFocused = props.element.contains(
        window.document.activeElement,
      );

      if (
        !props.editor.isEditable ||
        !isNonEmptyTextSelection ||
        (!props.view.hasFocus() && !isMenuFocused)
      ) {
        return false;
      }

      return bubbleMenuShouldShow?.(props) ?? true;
    },
    [bubbleMenuShouldShow],
  );

  return (
    <>
      {editor && !asTitle && !hideBubbleMenu && (
        <BubbleMenu
          editor={editor}
          options={{
            onHide: () => {
              setActiveBubbleMenu(null);
            },
          }}
          shouldShow={shouldShowBubbleMenu}
        >
          <CustomBubbleMenu
            activeMenu={activeBubbleMenu}
            createActions={bubbleMenuCreateActions}
            editor={editor}
            setActiveMenu={setActiveBubbleMenu}
          />
        </BubbleMenu>
      )}
      <EditorContent
        className={cn(
          {
            "rich-text-editor prose prose-lg max-w-full prose-stone prose-a:text-primary dark:prose-invert prose-headings:font-medium prose-pre:text-foreground prose-pre:bg-surface-muted prose-pre:text-[1.1rem] prose-strong:font-bold prose-h1:text-3xl prose-h2:text-2xl prose-h3:text-xl prose-h4:text-lg prose-h5:text-lg prose-h6:text-lg":
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
