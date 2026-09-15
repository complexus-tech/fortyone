import { useState } from "react";
import { ActivityIndicator, Pressable, View } from "react-native";
import { Col, Text } from "@/components/ui";
import type { DetailedStory } from "@/modules/stories/types";
import { RichTextEditor } from "@/components/rich-text/editor";
import { RichTextViewer } from "@/components/rich-text/viewer";
import {
  getDescriptionHtml,
  isRichTextValue,
  type RichTextValue,
} from "@/components/rich-text/content";
import { useDraft } from "@/components/rich-text/use-draft";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { useUpdateStoryMutation } from "../hooks/use-update-story-mutation";
import { getStory } from "@/modules/stories/queries/get-story";

type DescriptionDraft = { baseline: string; value: RichTextValue };
const isDescriptionDraft = (value: unknown): value is DescriptionDraft => {
  if (!value || typeof value !== "object") return false;
  const draft = value as Record<string, unknown>;
  return typeof draft.baseline === "string" && isRichTextValue(draft.value);
};

const EditDescription = ({
  story,
  onClose,
}: {
  story: DetailedStory;
  onClose: () => void;
}) => {
  const html = getDescriptionHtml(story.descriptionHTML, story.description);
  const draft = useDraft(
    `story:${story.id}:description`,
    {
      baseline: html,
      value: { html, text: story.description ?? "", mentions: [] },
    },
    isDescriptionDraft,
  );
  const mutation = useUpdateStoryMutation();
  if (!draft.ready)
    return (
      <View>
        <ActivityIndicator accessibilityLabel="Restoring description draft" />
        {draft.error && (
          <Text color="danger" accessibilityRole="alert">
            {draft.error}
          </Text>
        )}
        {draft.error && <DiscardDraftButton onDiscard={draft.reset} />}
        <Text onPress={onClose}>Close</Text>
      </View>
    );
  return (
    <RichTextEditor
      initialHtml={draft.value.value.html}
      onClose={onClose}
      onDiscard={async () => {
        await draft.clear();
        onClose();
      }}
      onDraft={async (value) =>
        draft.persist({ ...draft.valueRef.current, value })
      }
      onSave={async (value) => {
        const latest = await getStory(story.id);
        const latestHtml = getDescriptionHtml(
          latest.descriptionHTML,
          latest.description,
        );
        if (latestHtml !== draft.value.baseline && latestHtml !== value.html) {
          throw new Error(
            "This description changed elsewhere after your draft started. Your draft is preserved. Review the latest version on the web before replacing it.",
          );
        }
        if (latestHtml !== value.html)
          await mutation.mutateAsync({
            storyId: story.id,
            payload: { description: value.text, descriptionHTML: value.html },
          });
        await draft.clear();
        onClose();
      }}
    />
  );
};

export const Description = ({ story }: { story: DetailedStory }) => {
  const [editing, setEditing] = useState(false);
  const html = getDescriptionHtml(story.descriptionHTML, story.description);
  return (
    <Col asContainer align="stretch" className="mb-[12px]">
      {html ? <RichTextViewer html={html} /> : null}
      {!story.deletedAt && (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={html ? "Edit description" : "Add description"}
          onPress={() => setEditing(true)}
          className="min-h-[44px] justify-center py-[4px]"
        >
          <Text color="muted">
            {html ? "Edit description" : "Add description…"}
          </Text>
        </Pressable>
      )}
      {editing && (
        <EditDescription story={story} onClose={() => setEditing(false)} />
      )}
    </Col>
  );
};
