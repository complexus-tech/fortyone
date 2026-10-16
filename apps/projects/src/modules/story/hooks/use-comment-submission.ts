import type { Editor } from "@tiptap/core";
import { useCallback, useMemo, useRef, useState } from "react";
import { toast } from "sonner";
import { useUpdateCommentMutation } from "@/lib/hooks/update-comment-mutation";
import {
  extractAgentMentionsFromHTML,
  extractMentionsFromHTML,
} from "@/lib/utils/mentions";
import type { useDesktopDraft } from "@/shell/desktop/use-desktop-draft";
import type {
  CommentDraft,
  useBrowserCommentDraft,
} from "./use-browser-comment-draft";
import { useCommentStoryMutation } from "./comment-mutation";

export type CommentSubmissionIdentity = { html: string; creationId: string };

export const useCommentSubmissionIdentity = () => {
  const identity = useRef<CommentSubmissionIdentity | null>(null);
  const get = useCallback(() => identity.current, []);
  const set = useCallback((value: CommentSubmissionIdentity | null) => {
    identity.current = value;
  }, []);
  return useMemo(() => ({ get, set }), [get, set]);
};

type Options = {
  editor: Editor | null;
  ready: boolean;
  attachments: { hasPendingUploads: () => boolean };
  mayaEnabled: boolean;
  storyId: string;
  parentId?: string;
  commentId?: string;
  onCancel?: () => void;
  submission: ReturnType<typeof useCommentSubmissionIdentity>;
  activeDraft: { save: (payload: CommentDraft) => Promise<void> };
  draft: Pick<
    ReturnType<typeof useDesktopDraft<CommentDraft>>,
    "supported" | "discard"
  >;
  browserDraft: Pick<ReturnType<typeof useBrowserCommentDraft>, "discard">;
};

export const useCommentSubmission = ({
  editor,
  ready,
  attachments,
  mayaEnabled,
  storyId,
  parentId,
  commentId,
  onCancel,
  submission,
  activeDraft,
  draft,
  browserDraft,
}: Options) => {
  const { mutateAsync: createComment, isPending: isCreating } =
    useCommentStoryMutation();
  const { mutateAsync: updateComment, isPending: isUpdating } =
    useUpdateCommentMutation();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const submissionInFlight = useRef(false);
  const isPending = isSubmitting || isCreating || isUpdating;
  const canEdit = ready && !isPending;
  const handleComment = async () => {
    if (
      !editor ||
      !canEdit ||
      attachments.hasPendingUploads() ||
      submissionInFlight.current
    )
      return;
    if (editor.isEmpty) {
      toast.error("Comment is required", {
        description: "Please enter a comment before submitting",
      });
      return;
    }

    const comment = editor.getHTML();
    const mentions = extractMentionsFromHTML(comment).map((m) => m.id);
    const agentMentions = extractAgentMentionsFromHTML(comment);
    if (!commentId && agentMentions.length && !mayaEnabled) {
      toast.error("Maya is not available for this story right now.", {
        description:
          "Your draft is saved. Try again, or remove the Maya mention to post a regular comment.",
      });
      return;
    }
    submissionInFlight.current = true;
    setIsSubmitting(true);
    editor.setEditable(false, false);

    try {
      if (commentId) {
        await updateComment({
          commentId,
          payload: { content: comment, mentions },
          storyId,
        });
      } else {
        const needsIdentity = Boolean(agentMentions.length || parentId);
        let identity = submission.get();
        if (needsIdentity && identity?.html !== comment) {
          identity = { html: comment, creationId: crypto.randomUUID() };
          submission.set(identity);
        }
        if (needsIdentity && identity) {
          await activeDraft.save({
            version: 1,
            html: comment,
            creationId: identity.creationId,
          });
        }
        await createComment({
          storyId,
          payload: {
            comment,
            parentId: parentId ?? null,
            mentions,
            ...(agentMentions.length ? { agentMentions } : {}),
            ...(needsIdentity && identity
              ? { creationId: identity.creationId }
              : {}),
          },
        });
      }
    } catch {
      // Mutation hooks report the failure and retain the editor for retry.
      submissionInFlight.current = false;
      setIsSubmitting(false);
      if (!editor.isDestroyed) editor.setEditable(true, false);
      return;
    }

    try {
      if (draft.supported) {
        await draft.discard(
          (payload) =>
            typeof payload === "object" &&
            payload !== null &&
            "html" in payload &&
            payload.html === comment,
        );
      } else {
        await browserDraft.discard(comment);
      }
    } catch {
      toast.error("Comment saved, but the local draft could not be removed.");
    }
    // Do not erase a newer draft if this composer was reopened or its content
    // changed while the request was in flight.
    if (!editor.isDestroyed && editor.getHTML() === comment) {
      submission.set(null);
      editor.commands.clearContent(false);
      onCancel?.();
    }
    submissionInFlight.current = false;
    setIsSubmitting(false);
    if (!editor.isDestroyed) editor.setEditable(true, false);
  };

  return { handleComment, isPending, canEdit };
};
