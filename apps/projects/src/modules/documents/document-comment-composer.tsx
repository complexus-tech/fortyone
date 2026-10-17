"use client";

import { useEffect, useRef, useState } from "react";
import { useEditor, useEditorState } from "@tiptap/react";
import { BubbleMenu } from "@tiptap/react/menus";
import StarterKit from "@tiptap/starter-kit";
import Placeholder from "@tiptap/extension-placeholder";
import { Markdown } from "@tiptap/markdown";
import { Button, TextEditor } from "ui";
import {
  DocumentCommentMention,
  DocumentMentionSuggestions,
} from "./document-mentions";
import { useDocumentMentions } from "./use-document-mentions";
import { DocumentCommentFormatting } from "./document-comment-formatting";
import { useDocumentCommentMutations } from "./hooks";
import { renderDocumentComment } from "./document-comment-markdown";
import {
  useDocumentCommentDraft,
  useDocumentCommentScope,
} from "./use-document-comment-draft";
import type { DocumentCommentSelection, DocumentCommentThread } from "./types";

type ComposerProps = {
  documentId: string;
  draftTarget: string;
  canComment: boolean;
  quote?: string;
  invalidAnchor?: boolean;
  pending?: boolean;
  replyToCommentId?: string;
  label: "Comment" | "Reply";
  onCancel: () => void;
  onSubmit: (
    body: string,
    creationId: string,
    replyToCommentId?: string,
  ) => Promise<unknown>;
  onSubmitted?: (result: unknown) => void;
};

function Composer({
  documentId,
  draftTarget,
  canComment,
  quote,
  invalidAnchor,
  pending = false,
  replyToCommentId,
  label,
  onCancel,
  onSubmit,
  onSubmitted,
}: ComposerProps) {
  const draft = useDocumentCommentDraft(documentId, draftTarget);
  const replyTarget = draft.replyToCommentId ?? replyToCommentId;
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const submitted = useRef(false);
  const mounted = useRef(false);
  const hydrated = useRef(false);
  const focused = useRef(false);
  const form = useRef<HTMLFormElement>(null);
  const locked = !draft.ready || !canComment || isSubmitting || pending;
  const editor = useEditor({
    extensions: [
      DocumentCommentMention,
      DocumentMentionSuggestions,
      StarterKit.configure({
        underline: false,
        heading: false,
        horizontalRule: false,
        link: { openOnClick: false },
      }),
      Placeholder.configure({
        placeholder: label === "Reply" ? "Write a reply…" : "Write a comment…",
      }),
      Markdown.configure({ markedOptions: { gfm: true } }),
    ],
    content: "",
    editable: false,
    immediatelyRender: false,
    editorProps: {
      attributes: {
        role: "textbox",
        "aria-label": label,
        "aria-multiline": "true",
        class: "min-h-24 px-3 py-3 outline-none",
      },
      handleKeyDown: (_view, event) => {
        if (
          (event.metaKey || event.ctrlKey) &&
          event.key === "Enter" &&
          !event.isComposing
        ) {
          event.preventDefault();
          form.current?.requestSubmit();
          return true;
        }
        return false;
      },
    },
    onUpdate: ({ editor: current }) => {
      if (hydrated.current) draft.save(current.getMarkdown(), replyTarget);
    },
  });
  useDocumentMentions(editor, documentId, "comment");
  const content = useEditorState({
    editor,
    selector: ({ editor: current }) => ({
      empty: current?.isEmpty ?? true,
      body: current?.getMarkdown() ?? "",
    }),
  });
  const tooLong = Array.from(content?.body ?? "").length > 10_000;

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  useEffect(() => {
    if (!editor || !draft.ready) return;
    if (hydrated.current && editor.getMarkdown() === draft.body) return;
    editor.commands.setContent(renderDocumentComment(draft.body), {
      emitUpdate: false,
    });
    hydrated.current = true;
  }, [draft.body, draft.ready, editor]);
  useEffect(() => {
    editor?.setEditable(!locked, false);
    if (editor && !locked && !focused.current) {
      focused.current = true;
      editor.commands.focus("end");
    }
  }, [editor, locked]);

  const submit = () => {
    if (
      !editor ||
      locked ||
      submitted.current ||
      editor.isEmpty ||
      invalidAnchor ||
      tooLong
    )
      return;
    const body = editor.getMarkdown();
    if (!body.trim()) return;
    submitted.current = true;
    setIsSubmitting(true);
    setSubmitError(null);
    editor.setEditable(false, false);
    const creationId = draft.save(body, replyTarget) ?? crypto.randomUUID();
    void Promise.resolve()
      .then(() => onSubmit(body, creationId, replyTarget))
      .then((result) => {
        draft.clear(body, creationId);
        if (mounted.current) {
          editor.commands.clearContent(false);
          onSubmitted?.(result);
        }
      })
      .catch(() => {
        if (mounted.current)
          setSubmitError(
            "Your comment could not be sent. Your draft is still here — try again.",
          );
      })
      .finally(() => {
        submitted.current = false;
        if (mounted.current) setIsSubmitting(false);
      });
  };

  return (
    <form
      aria-busy={isSubmitting}
      className="border-border bg-surface overflow-hidden rounded-xl border shadow-sm"
      onSubmit={(event) => {
        event.preventDefault();
        submit();
      }}
      ref={form}
    >
      {quote ? (
        <div className="border-border bg-surface-muted/40 border-b px-3 py-2.5">
          <p className="text-text-muted border-primary/60 line-clamp-2 border-l-2 pl-2.5 text-base leading-6">
            {quote}
          </p>
        </div>
      ) : null}
      <TextEditor
        className="document-comment-content prose-base prose-a:text-primary text-base leading-6 [&_.tiptap]:max-h-60 [&_.tiptap]:overflow-y-auto [&_.tiptap_p]:my-1 [&_.tiptap_p:first-child]:mt-0 [&_.tiptap_p:last-child]:mb-0"
        editor={editor}
        hideBubbleMenu
      />
      {editor ? (
        <BubbleMenu
          editor={editor}
          shouldShow={({ editor: current, state }) =>
            current.isEditable && !state.selection.empty
          }
        >
          <div className="border-border bg-surface rounded-lg border p-1 shadow-lg">
            <DocumentCommentFormatting disabled={locked} editor={editor} />
          </div>
        </BubbleMenu>
      ) : null}
      {invalidAnchor ? (
        <p className="text-warning px-3 pb-3 text-base leading-6" role="alert">
          The selected text has changed. Select a passage to start a new
          comment.
        </p>
      ) : null}
      {tooLong ? (
        <p className="text-danger px-3 pb-3 text-base" role="alert">
          Keep your comment under 10,000 characters.
        </p>
      ) : null}
      {submitError || draft.error ? (
        <p className="text-danger px-3 pb-3 text-base leading-6" role="alert">
          {submitError ?? draft.error}
        </p>
      ) : null}
      <div className="border-border/60 flex flex-wrap items-center justify-between gap-2 border-t px-2 py-2">
        {editor ? (
          <DocumentCommentFormatting disabled={locked} editor={editor} />
        ) : (
          <span />
        )}
        <div className="ml-auto flex items-center gap-1">
          <Button
            className="text-base"
            color="tertiary"
            disabled={isSubmitting}
            onClick={onCancel}
            size="sm"
            type="button"
            variant="naked"
          >
            Cancel
          </Button>
          <Button
            className="px-3 text-base"
            disabled={locked || content?.empty || invalidAnchor || tooLong}
            size="sm"
            title="Send (⌘/Ctrl+Enter)"
            type="submit"
          >
            {isSubmitting ? "Sending…" : "Send"}
          </Button>
        </div>
      </div>
    </form>
  );
}

export function DocumentCommentComposer({
  documentId,
  draft,
  draftKey,
  canComment = true,
  onCancel,
  onCreated,
}: {
  documentId: string;
  draft: DocumentCommentSelection;
  draftKey?: string;
  canComment?: boolean;
  onCancel: () => void;
  onCreated?: (thread: DocumentCommentThread) => void;
}) {
  const { create, isCreatingDraft } = useDocumentCommentMutations(documentId);
  const scope = useDocumentCommentScope();
  const target = `selection:${draftKey ?? draft.draftKey ?? JSON.stringify([draft.from, draft.to, draft.text])}`;
  return (
    <Composer
      canComment={canComment}
      documentId={documentId}
      draftTarget={target}
      invalidAnchor={
        draft.isDetached || draft.from < 1 || draft.to <= draft.from
      }
      key={JSON.stringify([
        scope.accountId,
        scope.workspaceId,
        documentId,
        target,
      ])}
      label="Comment"
      onCancel={onCancel}
      onSubmit={(body, creationId) =>
        create.mutateAsync({
          creationId,
          draftTarget: target,
          body,
          quote: draft.text,
          anchorStart: draft.from,
          anchorEnd: draft.to,
        })
      }
      onSubmitted={(thread) => onCreated?.(thread as DocumentCommentThread)}
      pending={create.isPending || isCreatingDraft(target)}
      quote={draft.text}
    />
  );
}

export function DocumentCommentReplyComposer({
  documentId,
  threadId,
  replyToCommentId,
  canComment,
  pending,
  onCancel,
  onCreated,
}: {
  documentId: string;
  threadId: string;
  replyToCommentId?: string;
  canComment: boolean;
  pending: boolean;
  onCancel: () => void;
  onCreated: () => void;
}) {
  const { reply } = useDocumentCommentMutations(documentId);
  const scope = useDocumentCommentScope();
  return (
    <Composer
      canComment={canComment}
      documentId={documentId}
      draftTarget={`reply:${threadId}`}
      key={JSON.stringify([
        scope.accountId,
        scope.workspaceId,
        documentId,
        threadId,
      ])}
      label="Reply"
      onCancel={onCancel}
      onSubmit={(body, creationId, targetCommentId) =>
        reply.mutateAsync({
          body,
          threadId,
          replyToCommentId: targetCommentId,
          creationId,
        })
      }
      onSubmitted={onCreated}
      pending={pending || reply.isPending}
      replyToCommentId={replyToCommentId}
    />
  );
}
