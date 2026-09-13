import { useRef, useState } from "react";
import { Pressable, View } from "react-native";
import { Text } from "@/components/ui";
import { RichTextEditor } from "@/components/rich-text/editor";
import { RichTextViewer } from "@/components/rich-text/viewer";
import type { RichTextValue } from "@/components/rich-text/content";

type Props = {
  value: RichTextValue;
  onChange: (value: RichTextValue) => Promise<void>;
  disabled?: boolean;
};

export const DescriptionEditor = ({ value, onChange, disabled }: Props) => {
  const [editing, setEditing] = useState(false);
  const originalValue = useRef(value);
  return (
    <View>
      {value.html ? <RichTextViewer html={value.html} /> : null}
      <Pressable
        disabled={disabled}
        accessibilityRole="button"
        accessibilityLabel={value.html ? "Edit description" : "Add description"}
        onPress={() => {
          originalValue.current = value;
          setEditing(true);
        }}
        className="min-h-14 justify-center rounded-xl border border-gray-200 px-4 py-3 dark:border-dark-100"
      >
        <Text color="muted">
          {value.html ? "Edit description" : "Add a description…"}
        </Text>
      </Pressable>
      {editing && (
        <RichTextEditor
          initialHtml={value.html}
          onDraft={onChange}
          saveLabel="Done"
          onSave={async (next) => {
            await onChange(next);
            setEditing(false);
          }}
          onClose={() => setEditing(false)}
          onDiscard={async () => {
            await onChange(originalValue.current);
            setEditing(false);
          }}
        />
      )}
    </View>
  );
};
