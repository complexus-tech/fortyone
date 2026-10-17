"use client";

import type { ReactNode } from "react";
import type { Editor } from "@tiptap/react";
import { useMemo } from "react";
import { Dialog, Popover } from "ui";
import { DocumentCommentComposer } from "./document-comment-composer";
import type { DocumentCommentSelection, DocumentCommentThread } from "./types";

export const DocumentCommentsOverlay = ({
  children,
  isDesktop,
  onClose,
}: {
  children: ReactNode;
  isDesktop: boolean;
  onClose: () => void;
}) => {
  if (!isDesktop) {
    return (
      <Dialog
        onOpenChange={(open) => {
          if (!open) onClose();
        }}
        open
      >
        <Dialog.Content
          aria-describedby={undefined}
          className="mt-auto mb-3 flex h-[85dvh] max-h-[85dvh] flex-col"
          hideClose
          overlayClassName="items-end"
        >
          <Dialog.Title className="sr-only">Document comments</Dialog.Title>
          {children}
        </Dialog.Content>
      </Dialog>
    );
  }

  return (
    <div className="border-border bg-surface-elevated shadow-shadow absolute top-20 right-4 bottom-4 z-30 w-[calc(25rem_+_20px)] max-w-[calc(100%_-_2rem)] overflow-hidden rounded-2xl border-[0.5px] shadow-xl">
      {children}
    </div>
  );
};

export const DocumentCommentSelectionPopover = ({
  canComment,
  documentId,
  draft,
  editor,
  isDesktop,
  onClose,
  onCreated,
}: {
  canComment: boolean;
  documentId: string;
  draft: DocumentCommentSelection;
  editor: Editor;
  isDesktop: boolean;
  onClose: () => void;
  onCreated: (thread: DocumentCommentThread) => void;
}) => {
  // A virtual anchor follows the selected passage when its scroll container moves.
  const anchor = useMemo(
    () => ({
      current: {
        getBoundingClientRect: () => {
          if (editor.isDestroyed) return new DOMRect();
          const maximum = editor.state.doc.content.size;
          const start = editor.view.coordsAtPos(Math.min(draft.from, maximum));
          const end = editor.view.coordsAtPos(Math.min(draft.to, maximum));
          const left = Math.min(start.left, end.left);
          const top = Math.min(start.top, end.top);
          return new DOMRect(
            left,
            top,
            Math.max(start.right, end.right) - left,
            Math.max(start.bottom, end.bottom) - top,
          );
        },
        contextElement: editor.view.dom,
      },
    }),
    [draft.from, draft.to, editor],
  );

  const composer = (
    <DocumentCommentComposer
      canComment={canComment}
      documentId={documentId}
      draft={draft}
      onCancel={onClose}
      onCreated={onCreated}
    />
  );

  if (!isDesktop) {
    return (
      <Dialog
        onOpenChange={(open) => {
          if (!open) onClose();
        }}
        open
      >
        <Dialog.Content aria-describedby={undefined} className="p-3" hideClose>
          <Dialog.Title className="sr-only">
            Comment on selected text
          </Dialog.Title>
          {composer}
        </Dialog.Content>
      </Dialog>
    );
  }

  return (
    <Popover
      onOpenChange={(open) => {
        if (!open) onClose();
      }}
      open
    >
      <Popover.Anchor virtualRef={anchor} />
      <Popover.Content
        align="start"
        aria-label="Comment on selected text"
        className="m-0 w-[26rem] max-w-[calc(100vw_-_2rem)] p-0"
        collisionPadding={16}
        onCloseAutoFocus={(event) => {
          event.preventDefault();
        }}
        onInteractOutside={(event) => {
          if (
            event.target instanceof Element &&
            event.target.closest("[data-document-mention-picker]")
          )
            event.preventDefault();
        }}
        side="bottom"
        sideOffset={12}
        updatePositionStrategy="always"
      >
        {composer}
      </Popover.Content>
    </Popover>
  );
};
