import type { RichTextValue } from "@/components/rich-text/content";
import { useState } from "react";
import { Pressable, View } from "react-native";
import { Col, Text, Button } from "@/components/ui";
import { QueryState } from "@/components/ui/query-state";
import { RichTextEditor } from "@/components/rich-text/editor";
import { RichTextViewer } from "@/components/rich-text/viewer";
import { isRichTextValue } from "@/components/rich-text/content";
import { useDraft } from "@/components/rich-text/use-draft";

type DescriptionDraft = { baseline: string; value: RichTextValue };
const isDraft = (value: unknown): value is DescriptionDraft => {
  if (!value || typeof value !== "object") return false;
  const draft = value as DescriptionDraft;
  return typeof draft.baseline === "string" && isRichTextValue(draft.value);
};
type Props = {
  html: string;
  draftKey: string;
  onSave?: (html: string, baseline: string) => Promise<void>;
};
function EditDescription({
  html,
  draftKey,
  onSave,
  onClose,
}: Props & { onSave: NonNullable<Props["onSave"]>; onClose: () => void }) {
  const draft = useDraft(
    draftKey,
    {
      baseline: html,
      value: { html, text: "", mentions: [] } as RichTextValue,
    },
    isDraft,
  );
  if (!draft.ready)
    return (
      <View>
        <QueryState
          loading={!draft.error}
          title="Restoring description draft"
          message={draft.error ?? undefined}
        />
        {draft.error ? (
          <Button onPress={() => void draft.reset()}>
            Discard unreadable draft
          </Button>
        ) : null}
        <Button onPress={onClose}>Close</Button>
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
      onDraft={(value) => draft.persist({ ...draft.valueRef.current, value })}
      onSave={async (value) => {
        await onSave(value.html, draft.value.baseline);
        await draft.clear();
        onClose();
      }}
    />
  );
}
export function DetailDescription(props: Props) {
  const [editing, setEditing] = useState(false);
  return (
    <Col asContainer align="stretch" className="mb-[12px]">
      {props.html ? <RichTextViewer html={props.html} /> : null}
      {props.onSave ? (
        <Pressable
          accessibilityRole="button"
          onPress={() => setEditing(true)}
          className="min-h-[44px] justify-center py-[4px]"
        >
          <Text color="muted">
            {props.html ? "Edit description" : "Add description…"}
          </Text>
        </Pressable>
      ) : null}
      {editing && props.onSave ? (
        <EditDescription
          {...props}
          onSave={props.onSave}
          onClose={() => setEditing(false)}
        />
      ) : null}
    </Col>
  );
}
