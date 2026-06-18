import React, { useRef, useState } from "react";
import {
  Pressable,
  TextInput,
  TextInputProps,
  View,
  Text as RNText,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";

type Selection = NonNullable<TextInputProps["selection"]>;

type EditorAction = {
  icon: keyof typeof Ionicons.glyphMap;
  label: string;
  prefix: string;
  suffix?: string;
  placeholder: string;
};

const EDITOR_ACTIONS: EditorAction[] = [
  {
    icon: "text",
    label: "Bold",
    prefix: "**",
    suffix: "**",
    placeholder: "important detail",
  },
  {
    icon: "sparkles",
    label: "Italic",
    prefix: "_",
    suffix: "_",
    placeholder: "note",
  },
  {
    icon: "list",
    label: "Bullet",
    prefix: "- ",
    placeholder: "next step",
  },
  {
    icon: "code-slash",
    label: "Code",
    prefix: "`",
    suffix: "`",
    placeholder: "value",
  },
];

type DescriptionEditorProps = {
  value: string;
  onChangeText: (value: string) => void;
};

export const DescriptionEditor = ({
  value,
  onChangeText,
}: DescriptionEditorProps) => {
  const inputRef = useRef<TextInput>(null);
  const { resolvedTheme } = useTheme();
  const [selection, setSelection] = useState<Selection>({
    start: value.length,
    end: value.length,
  });

  const borderColor =
    resolvedTheme === "light" ? colors.gray[100] : colors.dark[100];
  const foregroundColor = resolvedTheme === "light" ? colors.black : colors.white;
  const mutedColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const applyAction = (action: EditorAction) => {
    const selected = value.slice(selection.start, selection.end);
    const text = selected || action.placeholder;
    const suffix = action.suffix ?? "";
    const nextValue =
      value.slice(0, selection.start) +
      action.prefix +
      text +
      suffix +
      value.slice(selection.end);
    const cursor = selection.start + action.prefix.length + text.length + suffix.length;

    onChangeText(nextValue);
    requestAnimationFrame(() => {
      setSelection({ start: cursor, end: cursor });
      inputRef.current?.focus();
    });
  };

  return (
    <View
      style={{
        borderWidth: 1,
        borderColor,
        borderRadius: 18,
        borderCurve: "continuous",
        overflow: "hidden",
      }}
    >
      <View
        style={{
          flexDirection: "row",
          alignItems: "center",
          gap: 6,
          paddingHorizontal: 10,
          paddingVertical: 8,
          borderBottomWidth: 1,
          borderBottomColor: borderColor,
        }}
      >
        {EDITOR_ACTIONS.map((action) => (
          <Pressable
            key={action.label}
            accessibilityLabel={action.label}
            onPress={() => applyAction(action)}
            style={{
              width: 34,
              height: 32,
              alignItems: "center",
              justifyContent: "center",
              borderRadius: 10,
              backgroundColor:
                resolvedTheme === "light" ? colors.gray[50] : colors.dark[200],
            }}
          >
            <Ionicons name={action.icon} size={16} color={foregroundColor} />
          </Pressable>
        ))}
        <RNText
          style={{
            color: mutedColor,
            fontSize: 13,
            fontWeight: "600",
            marginLeft: "auto",
          }}
        >
          Markdown
        </RNText>
      </View>
      <TextInput
        ref={inputRef}
        value={value}
        onChangeText={onChangeText}
        onSelectionChange={(event) => setSelection(event.nativeEvent.selection)}
        selection={selection}
        multiline
        textAlignVertical="top"
        placeholder="Add details, acceptance criteria, links, or quick notes..."
        placeholderTextColor={mutedColor}
        style={{
          minHeight: 170,
          padding: 16,
          color: foregroundColor,
          fontSize: 16,
          lineHeight: 23,
          fontWeight: "500",
        }}
      />
    </View>
  );
};
