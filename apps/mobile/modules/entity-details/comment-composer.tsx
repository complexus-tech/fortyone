import type { RichTextValue } from "@/components/rich-text/content";
import { useState } from "react";
import { View } from "react-native";
import { randomUUID } from "expo-crypto";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { RichTextEditor } from "@/components/rich-text/editor";
import {
  EMPTY_RICH_TEXT,
  isRichTextValue,
} from "@/components/rich-text/content";
import { useDraft } from "@/components/rich-text/use-draft";
import { QueryState } from "@/components/ui/query-state";
import { Button, Text } from "@/components/ui";
import { CommentComposerTrigger } from "@/modules/story/components/comment-composer";
import { createCommentFinalizer } from "@/modules/story/components/comment-submission";

type CommentDraft = { content: RichTextValue; idempotencyKey: string };
const isDraft = (value: unknown): value is CommentDraft => {
  if (!value || typeof value !== "object") return false;
  const draft = value as CommentDraft;
  return (
    isRichTextValue(draft.content) && typeof draft.idempotencyKey === "string"
  );
};
type Props = {
  draftKey: string;
  label?: string;
  notice?: string;
  onSent?: () => void;
  onSend: (value: RichTextValue, idempotencyKey: string) => Promise<void>;
};
function Compose({
  draftKey,
  onSend,
  onSent,
  onClose,
}: Props & { onClose: () => void }) {
  const draft = useDraft<CommentDraft>(
    draftKey,
    { content: EMPTY_RICH_TEXT, idempotencyKey: randomUUID() },
    isDraft,
  );
  const [sent, setSent] = useState(false);
  const [finalizer] = useState(() =>
    createCommentFinalizer(() => setSent(true)),
  );
  if (!draft.ready)
    return (
      <View>
        <QueryState
          loading={!draft.error}
          title="Restoring comment draft"
          message={draft.error ?? undefined}
        />
        {draft.error ? (
          <Button color="tertiary" onPress={() => void draft.reset()}>
            Discard unreadable draft
          </Button>
        ) : null}
        <Button color="tertiary" onPress={onClose}>
          Close
        </Button>
      </View>
    );
  return (
    <RichTextEditor
      title={sent ? "Comment sent" : "Add comment"}
      saveLabel={sent ? "Finish" : "Send"}
      readOnly={sent}
      initialHtml={draft.value.content.html}
      onDraft={(content) =>
        finalizer.persistDraft(() =>
          draft.persist({ ...draft.valueRef.current, content }),
        )
      }
      onClose={async () => {
        await finalizer.close(draft.clear);
        onClose();
      }}
      onDiscard={async () => {
        await finalizer.discard(draft.clear);
        onClose();
      }}
      onSave={async (value) => {
        await finalizer.send(async () => {
          if (!value.text.trim())
            throw new Error("Write a comment before sending.");
          await onSend(value, draft.valueRef.current.idempotencyKey);
        }, draft.clear);
        onClose();
        onSent?.();
      }}
    />
  );
}
export function DetailCommentComposer(props: Props) {
  const [editing, setEditing] = useState(false);
  const insets = useSafeAreaInsets();
  return (
    <View
      className="px-[20px] pt-[8px]"
      style={{ paddingBottom: Math.max(insets.bottom, 12) }}
    >
      {props.notice ? (
        <Text color="muted" fontSize="xs" className="mb-2">
          {props.notice}
        </Text>
      ) : null}
      <CommentComposerTrigger
        label={props.label}
        onPress={() => setEditing(true)}
      />
      {editing ? (
        <Compose {...props} onClose={() => setEditing(false)} />
      ) : null}
    </View>
  );
}
