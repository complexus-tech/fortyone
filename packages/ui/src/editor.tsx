"use client";

import type { ComponentProps } from "react";
import { useCallback, useEffect, useState } from "react";
import { TextSelection } from "@tiptap/pm/state";
import { EditorContent, EditorContentProps } from "@tiptap/react";
import { BubbleMenu } from "@tiptap/react/menus";

import {
  BubbleMenu as CustomBubbleMenu,
  type BubbleMenuCommentAction,
  type BubbleMenuCreateAction,
  type BubbleMenuPanel,
} from "./bubble-menu";
import { cn } from "lib";

type TextEditorProps = EditorContentProps & {
  allowFormatting?: boolean;
  asTitle?: boolean;
  bubbleMenuCreateActions?: readonly BubbleMenuCreateAction[];
  bubbleMenuCommentAction?: BubbleMenuCommentAction;
  bubbleMenuShouldShow?: ComponentProps<typeof BubbleMenu>["shouldShow"];
  bubbleMenuOptions?: ComponentProps<typeof BubbleMenu>["options"];
  bubbleMenuAppendTo?: ComponentProps<typeof BubbleMenu>["appendTo"];
  bubbleMenuClassName?: string;
  hideBubbleMenu?: boolean;
};

export const TextEditor = ({
  allowFormatting,
  editor,
  className = "",
  asTitle = false,
  bubbleMenuCreateActions,
  bubbleMenuCommentAction,
  bubbleMenuShouldShow,
  bubbleMenuOptions,
  bubbleMenuAppendTo,
  bubbleMenuClassName,
  hideBubbleMenu = false,
  innerRef,
  ...rest
}: TextEditorProps) => {
  const [activeBubbleMenu, setActiveBubbleMenu] =
    useState<BubbleMenuPanel>(null);
  useEffect(() => {
    if (
      !editor ||
      asTitle ||
      hideBubbleMenu ||
      !bubbleMenuCommentAction ||
      (allowFormatting ?? editor.isEditable)
    ) {
      return;
    }

    const ownerDocument = editor.view.dom.ownerDocument;
    const syncReviewerSelection = () => {
      if (editor.isDestroyed) return;
      const { view, state } = editor;
      const selection = ownerDocument.getSelection();
      let nextSelection = TextSelection.near(state.selection.$from);
      if (
        selection?.anchorNode &&
        selection.focusNode &&
        view.dom.contains(selection.anchorNode) &&
        view.dom.contains(selection.focusNode)
      ) {
        nextSelection = TextSelection.between(
          state.doc.resolve(
            view.posAtDOM(selection.anchorNode, selection.anchorOffset),
          ),
          state.doc.resolve(
            view.posAtDOM(selection.focusNode, selection.focusOffset),
          ),
        );
      }

      if (nextSelection.eq(state.selection)) return;
      // Read-only editors do not reliably observe native text selection unless
      // focused. A selection-only transaction lets the comment menu follow it.
      view.dispatch(state.tr.setSelection(nextSelection));
    };

    ownerDocument.addEventListener("selectionchange", syncReviewerSelection);
    return () => {
      ownerDocument.removeEventListener(
        "selectionchange",
        syncReviewerSelection,
      );
    };
  }, [
    allowFormatting,
    asTitle,
    bubbleMenuCommentAction,
    editor,
    hideBubbleMenu,
  ]);

  const shouldShowBubbleMenu = useCallback<
    NonNullable<ComponentProps<typeof BubbleMenu>["shouldShow"]>
  >(
    (props) => {
      const canFormat = allowFormatting ?? props.editor.isEditable;
      const { doc, selection } = props.state;
      const isNonEmptyTextSelection =
        selection instanceof TextSelection &&
        !selection.empty &&
        doc.textBetween(props.from, props.to).length > 0;
      const isMenuFocused = props.element.contains(
        window.document.activeElement,
      );
      const selectionNode = window.getSelection()?.anchorNode;
      const isReadOnlySelection =
        !canFormat &&
        Boolean(selectionNode && props.view.dom.contains(selectionNode));

      if (
        (!canFormat && !bubbleMenuCommentAction) ||
        !isNonEmptyTextSelection ||
        (!props.view.hasFocus() && !isMenuFocused && !isReadOnlySelection)
      ) {
        return false;
      }

      return bubbleMenuShouldShow?.(props) ?? true;
    },
    [allowFormatting, bubbleMenuCommentAction, bubbleMenuShouldShow],
  );

  return (
    <>
      {editor && !asTitle && !hideBubbleMenu && (
        <BubbleMenu
          appendTo={bubbleMenuAppendTo}
          className={bubbleMenuClassName}
          editor={editor}
          options={{
            ...bubbleMenuOptions,
            onHide: () => {
              setActiveBubbleMenu(null);
              bubbleMenuOptions?.onHide?.();
            },
          }}
          shouldShow={shouldShowBubbleMenu}
        >
          <CustomBubbleMenu
            activeMenu={activeBubbleMenu}
            allowFormatting={allowFormatting}
            commentAction={bubbleMenuCommentAction}
            createActions={bubbleMenuCreateActions}
            editor={editor}
            setActiveMenu={setActiveBubbleMenu}
          />
        </BubbleMenu>
      )}
      <EditorContent
        className={cn(
          {
            "rich-text-editor prose prose-lg max-w-full prose-headings:font-medium prose-pre:text-[1.1rem] prose-strong:font-bold prose-h1:text-3xl prose-h2:text-2xl prose-h3:text-xl prose-h4:text-lg prose-h5:text-lg prose-h6:text-lg":
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
