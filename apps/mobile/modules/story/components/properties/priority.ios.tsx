import { PropertyChip } from "./property-chip";
import React, { useState, useRef } from "react";
import { BottomSheetModal, Text as UIText } from "@/components/ui";
import { PriorityIcon } from "@/components/icons";
import { StoryPriority } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import {
  Button,
  HStack,
  Image,
  RNHostView,
  Spacer,
  Text,
} from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  disabled as disabledModifier,
  font,
  foregroundStyle,
  frame,
} from "@expo/ui/swift-ui/modifiers";

const Item = ({
  priority,
  onPress,
  isSelected,
  disabled = false,
}: {
  priority: StoryPriority;
  onPress: () => void;
  isSelected: boolean;
  disabled?: boolean;
}) => {
  const { resolvedTheme } = useTheme();
  return (
    <Button
      onPress={onPress}
      modifiers={[
        buttonStyle("plain"),
        disabledModifier(disabled),
        accessibilityLabel(`${priority}${isSelected ? ", selected" : ""}`),
      ]}
    >
      <HStack spacing={6} modifiers={[frame({ minHeight: 44 })]}>
        <RNHostView matchContents>
          <PriorityIcon size={20} priority={priority} />
        </RNHostView>
        <Text
          modifiers={[
            font({ textStyle: "body" }),
            foregroundStyle(themeColors[resolvedTheme].foreground),
          ]}
        >
          {priority}
        </Text>
        <Spacer />
        {isSelected && (
          <Image
            systemName="checkmark.circle.fill"
            size={17}
            color={themeColors[resolvedTheme].foreground}
          />
        )}
      </HStack>
    </Button>
  );
};

export const PriorityBadge = ({
  priority,
  onPriorityChange,
  disabled = false,
}: {
  priority: StoryPriority;
  onPriorityChange: (priority: StoryPriority) => Promise<void>;
  disabled?: boolean;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pending = useRef(false);
  const select = (value: StoryPriority) => {
    if (pending.current) return;
    pending.current = true;
    setSaving(true);
    setError(null);
    void onPriorityChange(value)
      .then(() => setIsOpen(false))
      .catch((cause: unknown) =>
        setError(
          cause instanceof Error
            ? cause.message
            : "Could not update this property. Try again.",
        ),
      )
      .finally(() => {
        pending.current = false;
        setSaving(false);
      });
  };
  const priorities: StoryPriority[] = [
    "No Priority",
    "Low",
    "Medium",
    "High",
    "Urgent",
  ];

  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`Change priority: ${priority}`}
        disabled={disabled}
        accessibilityState={{ disabled }}
        onPress={() => {
          setError(null);
          setIsOpen(true);
        }}
        style={({ pressed }) => ({
          minHeight: 44,
          maxWidth: "100%",
          minWidth: 0,
          flexShrink: 1,
          justifyContent: "center",
          opacity: pressed ? 0.6 : 1,
        })}
      >
        <PropertyChip color="tertiary">
          <PriorityIcon priority={priority || "No Priority"} />
          <UIText fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {priority || "No Priority"}
          </UIText>
        </PropertyChip>
      </Pressable>
      <BottomSheetModal
        nativeContent
        spacing={24}
        isOpen={isOpen}
        onClose={() => {
          if (!pending.current) setIsOpen(false);
        }}
        padding={{
          leading: 24,
          trailing: 24,
          top: 36,
          bottom: 5,
        }}
      >
        {error && <Text>{error}</Text>}
        {priorities.map((p) => (
          <Item
            key={p}
            priority={p}
            disabled={saving}
            onPress={() => select(p)}
            isSelected={priority === p}
          />
        ))}
      </BottomSheetModal>
    </>
  );
};
