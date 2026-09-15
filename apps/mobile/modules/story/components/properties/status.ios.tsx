import type { Status } from "@/types/statuses";
import type { SFSymbol } from "expo-symbols";
import { PropertyChip } from "./property-chip";
import React, { useState, useRef } from "react";
import { BottomSheetModal, Text as UIText } from "@/components/ui";
import { StatusIcon } from "@/components/icons";
import type { Story } from "@/modules/stories/types";
import { Pressable } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { Button, HStack, Image, Spacer, Text } from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  contentShape,
  disabled as disabledModifier,
  font,
  frame,
  shapes,
} from "@expo/ui/swift-ui/modifiers";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { hexToRgba } from "@/lib/utils/colors";
import { truncateText } from "@/lib/utils";

const STATUS_SYMBOLS: Record<Status["category"], SFSymbol> = {
  backlog: "circle.dashed",
  unstarted: "circle",
  started: "circle.righthalf.filled",
  paused: "pause.circle",
  completed: "checkmark.circle.fill",
  cancelled: "xmark.circle",
};

const Item = ({
  status,
  onPress,
  isSelected,
  disabled = false,
}: {
  status: Status;
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
        accessibilityLabel(`${status.name}${isSelected ? ", selected" : ""}`),
      ]}
    >
      <HStack
        spacing={10}
        modifiers={[frame({ minHeight: 44 }), contentShape(shapes.rectangle())]}
      >
        <HStack modifiers={[frame({ width: 18, height: 18 })]}>
          <Image
            systemName={STATUS_SYMBOLS[status.category]}
            color={status.color}
            size={18}
          />
        </HStack>
        <Text modifiers={[font({ textStyle: "body" })]}>{status.name}</Text>
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

export const StatusBadge = ({
  story,
  onStatusChange,
  disabled = false,
}: {
  story: Story;
  onStatusChange: (statusId: string) => Promise<void>;
  disabled?: boolean;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pending = useRef(false);
  const select = (value: string) => {
    if (pending.current) return;
    pending.current = true;
    setSaving(true);
    setError(null);
    void onStatusChange(value)
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
  const {
    data: statuses = [],
    isPending,
    error: loadError,
    refetch,
  } = useTeamStatuses(story.teamId);
  const currentStatus = statuses.find((s) => s.id === story.statusId);

  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`Change status: ${currentStatus?.name || "No status"}`}
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
        <PropertyChip
          style={{
            backgroundColor: hexToRgba(currentStatus?.color, 0.1),
            borderColor: hexToRgba(currentStatus?.color, 0.2),
          }}
        >
          <StatusIcon
            category={currentStatus?.category}
            color={currentStatus?.color}
            size={16}
          />
          <UIText fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {truncateText(currentStatus?.name || "No Status", 16)}
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
        {isPending && <Text>Loading statuses…</Text>}
        {loadError && (
          <>
            <Text>{loadError.message}</Text>
            <Button
              onPress={() => {
                void refetch();
              }}
              modifiers={[buttonStyle("plain")]}
            >
              <Text>Try again</Text>
            </Button>
          </>
        )}
        {!isPending && !loadError && statuses.length === 0 && (
          <Text>No statuses available</Text>
        )}
        {statuses.map((status) => (
          <Item
            key={status.id}
            status={status}
            disabled={saving}
            onPress={() => select(status.id)}
            isSelected={story.statusId === status.id}
          />
        ))}
      </BottomSheetModal>
    </>
  );
};
