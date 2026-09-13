import { useState } from "react";
import { ActivityIndicator, View } from "react-native";
import { Button, Text } from "@/components/ui";
import { RichTextEditor } from "@/components/rich-text/editor";
import {
  EMPTY_RICH_TEXT,
  isRichTextValue,
} from "@/components/rich-text/content";
import { useDraft } from "@/components/rich-text/use-draft";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { useCreateCommentMutation } from "../hooks/use-create-comment-mutation";

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
  const [saved, setSaved] = useState(false);
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
      title="Add comment"
      saveLabel="Send"
      initialHtml={draft.value.html}
      placeholder="Share an update or mention someone…"
      onDraft={draft.persist}
      onClose={onClose}
      onDiscard={async () => {
        await draft.clear();
        onClose();
      }}
      onSave={async (value) => {
        if (!value.text.trim())
          throw new Error("Write a comment before sending.");
        // A failed local draft cleanup must not send the comment a second time.
        if (!saved) {
          await mutation.mutateAsync({
            storyId,
            comment: value.html,
            mentions: value.mentions,
          });
          setSaved(true);
        }
        await draft.clear();
        onClose();
      }}
    />
  );
};

export const CommentComposer = ({ storyId }: { storyId: string }) => {
  const [editing, setEditing] = useState(false);
  return (
    <View className="my-3">
      <Button color="tertiary" onPress={() => setEditing(true)}>
        Add a comment
      </Button>
      {editing && (
        <ComposeComment storyId={storyId} onClose={() => setEditing(false)} />
      )}
    </View>
  );
};
