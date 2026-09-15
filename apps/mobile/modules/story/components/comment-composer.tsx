import { useState } from "react";
import {
  ActivityIndicator,
  Pressable,
  View,
  useColorScheme,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { themeColors } from "@/constants/colors";
import { Button, Text } from "@/components/ui";
import { RichTextEditor } from "@/components/rich-text/editor";
import {
  EMPTY_RICH_TEXT,
  isRichTextValue,
} from "@/components/rich-text/content";
import { useDraft } from "@/components/rich-text/use-draft";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { useCreateCommentMutation } from "../hooks/use-create-comment-mutation";
import { createCommentFinalizer } from "./comment-submission";

const ComposeComment = ({
  storyId,
  onClose,
}: {
  storyId: string;
  onClose: () => void;
}) => {
  const draft = useDraft(
    `story:${storyId}:comment`,
    EMPTY_RICH_TEXT,
    isRichTextValue,
  );
  const mutation = useCreateCommentMutation();
  const [sent, setSent] = useState(false);
  const [finalizer] = useState(() =>
    createCommentFinalizer(() => setSent(true)),
  );
  if (!draft.ready)
    return (
      <View>
        <ActivityIndicator accessibilityLabel="Restoring comment draft" />
        {draft.error && (
          <Text color="danger" accessibilityRole="alert">
            {draft.error}
          </Text>
        )}
        {draft.error && <DiscardDraftButton onDiscard={draft.reset} />}
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
      initialHtml={draft.value.html}
      placeholder="Share an update or mention someone…"
      onDraft={(value) => finalizer.persistDraft(() => draft.persist(value))}
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
          await mutation.mutateAsync({
            storyId,
            comment: value.html,
            mentions: value.mentions,
          });
        }, draft.clear);
        onClose();
      }}
    />
  );
};

export const CommentComposer = ({ storyId }: { storyId: string }) => {
  const [editing, setEditing] = useState(false);
  const dark = useColorScheme() === "dark";
  return (
    <View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel="Add a comment"
        onPress={() => setEditing(true)}
        className="min-h-[48px] flex-row items-center gap-[10px] rounded-full bg-gray-50 px-[16px] py-[10px] active:opacity-60 dark:bg-dark-100"
        style={{
          boxShadow: dark
            ? "0 0 0 1px rgba(255, 255, 255, 0.06)"
            : "0 2px 16px rgba(0, 0, 0, 0.06)",
        }}
      >
        <Ionicons
          name="add"
          size={22}
          color={themeColors[dark ? "dark" : "light"].textMuted}
        />
        <Text color="muted">Write a comment…</Text>
      </Pressable>
      {editing && (
        <ComposeComment storyId={storyId} onClose={() => setEditing(false)} />
      )}
    </View>
  );
};
