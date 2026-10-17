import type { Editor } from "@tiptap/core";
import type { Transaction } from "@tiptap/pm/state";
import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "next/navigation";
import {
  updateCommentDraftHighlight,
  updateCommentHighlights,
} from "@/lib/tiptap/comment-highlights";
import type { DocumentCommentSelection, DocumentCommentThread } from "./types";

type CommentInteractionOptions = {
  editor: Editor | null;
  commentThreads: DocumentCommentThread[];
  documentRevision?: number;
  collaborationReady: boolean;
  closeRelatedWork: () => void;
};

export function useDocumentCommentInteractions({
  editor,
  commentThreads,
  documentRevision,
  collaborationReady,
  closeRelatedWork,
}: CommentInteractionOptions) {
  const searchParams = useSearchParams();
  const [isCommentsOpen, setIsCommentsOpen] = useState(false);
  const [commentDraft, setCommentDraft] =
    useState<DocumentCommentSelection | null>(null);
  const [isCommentComposerOpen, setIsCommentComposerOpen] = useState(false);
  const [activeCommentThreadId, setActiveCommentThreadId] = useState<string>();
  const closeComments = useCallback(() => {
    setIsCommentsOpen(false);
  }, []);
  useEffect(() => {
    const threadId = searchParams.get("thread");
    if (!threadId) return;
    setActiveCommentThreadId(threadId);
    setIsCommentsOpen(true);
  }, [searchParams]);
  useEffect(() => {
    if (!editor) return;
    updateCommentHighlights(
      editor,
      commentThreads
        .filter((thread) => !thread.resolvedAt)
        .map((thread) => ({
          from: thread.anchorStart,
          id: thread.id,
          to: thread.anchorEnd,
          quote: thread.quote,
        })),
    );
  }, [commentThreads, editor, documentRevision, collaborationReady]);

  useEffect(() => {
    if (!editor) return;
    updateCommentDraftHighlight(
      editor,
      commentDraft && isCommentComposerOpen && !commentDraft.isDetached
        ? { from: commentDraft.from, id: "draft", to: commentDraft.to }
        : null,
    );
  }, [commentDraft, editor, isCommentComposerOpen]);

  useEffect(() => {
    if (!editor) return;
    const openThread = (event: MouseEvent) => {
      const target = event.target;
      if (!(target instanceof Element)) return;
      const threadId = target.closest<HTMLElement>("[data-comment-thread-id]")
        ?.dataset.commentThreadId;
      if (!threadId || threadId === "draft") return;
      setActiveCommentThreadId(threadId);
      setIsCommentComposerOpen(false);
      setIsCommentsOpen(true);
      closeRelatedWork();
    };
    const element = editor.view.dom;
    element.addEventListener("click", openThread);
    return () => {
      element.removeEventListener("click", openThread);
    };
  }, [editor, closeRelatedWork]);

  useEffect(() => {
    if (!editor) return;
    const mapDraft = ({ transaction }: { transaction: Transaction }) => {
      if (!transaction.docChanged) return;
      setCommentDraft((current) => {
        if (!current) return current;
        const from = transaction.mapping.map(current.from, 1);
        const to = transaction.mapping.map(current.to, -1);
        const isDetached =
          from < 1 ||
          to <= from ||
          to > transaction.doc.content.size ||
          transaction.doc.textBetween(from, to, "\n\n").trim() !== current.text;
        // Retain the quoted context and draft body even if the selected passage was removed.
        return { ...current, from, to, isDetached };
      });
    };
    editor.on("transaction", mapDraft);
    return () => {
      editor.off("transaction", mapDraft);
    };
  }, [editor]);

  const openComments = () => {
    closeRelatedWork();
    setIsCommentComposerOpen(false);
    setActiveCommentThreadId(undefined);
    setIsCommentsOpen(true);
  };
  const startComment = (selection: DocumentCommentSelection) => {
    setCommentDraft({
      ...selection,
      draftKey: `${selection.from}:${selection.to}:${selection.text}`,
    });
    closeRelatedWork();
    setIsCommentsOpen(false);
    setIsCommentComposerOpen(true);
  };
  return {
    isCommentsOpen,
    setIsCommentsOpen,
    commentDraft,
    setCommentDraft,
    isCommentComposerOpen,
    setIsCommentComposerOpen,
    activeCommentThreadId,
    setActiveCommentThreadId,
    closeComments,
    openComments,
    startComment,
  };
}
