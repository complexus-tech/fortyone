import type { Editor } from "@tiptap/core";
import { useCallback, useEffect, useRef } from "react";
import { useIsMutating } from "@tanstack/react-query";
import { getRichTextContentType } from "@/lib/tiptap/markdown";
import { useDesktopOutbox } from "@/shell/desktop/desktop-outbox-coordinator";

export const useStoryEditorSaving = (storyId: string) => {
  const pending = useIsMutating({
    predicate: (mutation) =>
      (mutation.state.variables as { storyId?: string } | undefined)
        ?.storyId === storyId,
  });
  const outbox = useDesktopOutbox();
  return (
    pending > 0 ||
    Boolean(
      outbox?.snapshot.operations.some(
        (operation) =>
          operation.resourceId === storyId &&
          (operation.kind === "story.update" ||
            operation.kind === "story.create"),
      ),
    )
  );
};

/** Reconcile remote changes after blur without replacing an unsaved local edit. */
export const useStoryEditorSync = ({
  editor,
  storyId,
  text,
  html,
  saving,
}: {
  editor: Editor | null;
  storyId: string;
  text: string;
  html?: string;
  saving: boolean;
}) => {
  const local = useRef({ storyId, editor, dirty: false });
  const markEdited = useCallback(() => {
    local.current = { storyId, editor, dirty: true };
  }, [editor, storyId]);

  useEffect(() => {
    if (!editor || editor.isDestroyed) return;
    const changedStory = local.current.storyId !== storyId;
    if (changedStory || local.current.editor !== editor) {
      local.current = { storyId, editor, dirty: false };
    }
    const apply = () =>
      editor.commands.setContent(html || text, {
        emitUpdate: false,
        ...(html !== undefined
          ? { contentType: getRichTextContentType(text, html) }
          : {}),
      });

    // A reused editor belongs to its new story immediately, even while focused.
    // MainDetails flushes the old story's captured draft before this effect runs.
    if (changedStory) apply();

    const sync = () => {
      if (editor.isDestroyed || saving) return;
      const matches = html
        ? editor.getHTML() === html
        : editor.getText() === text;
      if (matches) {
        local.current.dirty = false;
        return;
      }
      if (editor.isFocused || local.current.dirty) return;
      apply();
    };

    sync();
    editor.on("update", markEdited);
    editor.on("blur", sync);
    return () => {
      editor.off("update", markEdited);
      editor.off("blur", sync);
    };
  }, [editor, html, markEdited, saving, storyId, text]);

  return markEdited;
};
