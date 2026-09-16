import { useMemo, useState } from "react";
import {
  FlatList,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  TextInput,
  View,
  useColorScheme,
  useWindowDimensions,
} from "react-native";
import { useSafeAreaInsets } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import type { StoryPriority } from "@/modules/stories/types";
import type { StatusCategory } from "@/types/statuses";
import { Avatar, IconButton, Text } from "@/components/ui";
import { PriorityIcon, StatusIcon } from "@/components/icons";
import { themeColors } from "@/constants/colors";

export type MetadataOption = {
  id: string;
  label: string;
  description?: string;
  color?: string | null;
  icon?:
    | { kind: "status"; category: StatusCategory; color: string }
    | { kind: "priority"; priority: StoryPriority }
    | { kind: "assignee"; name: string; src?: string | null };
};

function MetadataOptionIcon({ option }: { option: MetadataOption }) {
  switch (option.icon?.kind) {
    case "status":
      return (
        <StatusIcon
          size={20}
          category={option.icon.category}
          color={option.icon.color}
        />
      );
    case "priority":
      return <PriorityIcon size={20} priority={option.icon.priority} />;
    case "assignee":
      return <Avatar size="sm" name={option.icon.name} src={option.icon.src} />;
    default:
      return option.color ? (
        <View
          style={{
            width: 10,
            height: 10,
            borderRadius: 3,
            backgroundColor: option.color,
          }}
        />
      ) : null;
  }
}

type MetadataSheetProps = {
  isOpen: boolean;
  title: string;
  options: MetadataOption[];
  selectedIds: string[];
  multiple?: boolean;
  emptyText?: string;
  onClose: () => void;
  onSelect: (option: MetadataOption) => void;
};

export const MetadataSheet = ({
  isOpen,
  title,
  options,
  selectedIds,
  multiple,
  emptyText = "Nothing available",
  onClose,
  onSelect,
}: MetadataSheetProps) => {
  const dark = useColorScheme() === "dark";
  const { height } = useWindowDimensions();
  const { bottom } = useSafeAreaInsets();
  const [query, setQuery] = useState("");
  const foreground = themeColors[dark ? "dark" : "light"].foreground;
  const muted = themeColors[dark ? "dark" : "light"].textMuted;
  const filteredOptions = useMemo(() => {
    const search = query.trim().toLowerCase();
    return search
      ? options.filter((option) =>
          `${option.label} ${option.description ?? ""}`
            .toLowerCase()
            .includes(search),
        )
      : options;
  }, [options, query]);

  return (
    <Modal
      visible={isOpen}
      transparent
      animationType="slide"
      onRequestClose={onClose}
    >
      <KeyboardAvoidingView
        behavior={Platform.OS === "ios" ? "padding" : "height"}
        style={{ flex: 1, justifyContent: "flex-end" }}
      >
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={`Close ${title.toLowerCase()} picker`}
          onPress={onClose}
          style={{
            position: "absolute",
            inset: 0,
            backgroundColor: "rgba(0,0,0,0.35)",
          }}
        />
        <View
          accessibilityViewIsModal
          style={{
            flexShrink: 1,
            maxHeight: "85%",
            marginHorizontal: 12,
            marginTop: 12,
            marginBottom: 12,
            borderRadius: 24,
            paddingHorizontal: 16,
            paddingTop: 8,
            paddingBottom: Math.max(12, bottom),
            backgroundColor: themeColors[dark ? "dark" : "light"].surface,
          }}
        >
          <View
            style={{
              flexDirection: "row",
              alignItems: "center",
              gap: 8,
              paddingBottom: 12,
            }}
          >
            <Text
              accessibilityRole="header"
              fontSize="lg"
              fontWeight="bold"
              style={{ flex: 1 }}
            >
              {title}
            </Text>
            <IconButton
              icon="close"
              label={`Close ${title.toLowerCase()} picker`}
              onPress={onClose}
            />
          </View>
          <View
            style={{
              flexDirection: "row",
              alignItems: "center",
              gap: 8,
              borderRadius: 22,
              paddingHorizontal: 12,
              marginBottom: 8,
              backgroundColor:
                themeColors[dark ? "dark" : "light"].surfaceMuted,
            }}
          >
            <Ionicons name="search" size={19} color={muted} />
            <TextInput
              accessibilityLabel={`Search ${title.toLowerCase()}`}
              placeholder="Search…"
              placeholderTextColor={muted}
              value={query}
              onChangeText={setQuery}
              autoCorrect={false}
              returnKeyType="search"
              style={{
                flex: 1,
                minHeight: 44,
                paddingVertical: 10,
                fontSize: 16,
                color: foreground,
              }}
            />
          </View>
          <FlatList
            data={filteredOptions}
            keyExtractor={(option) => option.id}
            extraData={selectedIds}
            initialNumToRender={12}
            maxToRenderPerBatch={12}
            windowSize={5}
            style={{
              maxHeight: Math.min(360, height * 0.5),
              flexGrow: 0,
              flexShrink: 1,
            }}
            keyboardShouldPersistTaps="handled"
            keyboardDismissMode="on-drag"
            ListEmptyComponent={
              <Text color="muted" style={{ paddingVertical: 24 }}>
                {options.length ? "No matches found" : emptyText}
              </Text>
            }
            renderItem={({ item: option }) => {
              const selected = selectedIds.includes(option.id);
              return (
                <Pressable
                  accessibilityRole={multiple ? "checkbox" : "radio"}
                  accessibilityLabel={
                    option.description
                      ? `${option.label}, ${option.description}`
                      : option.label
                  }
                  accessibilityState={{ checked: selected }}
                  onPress={() => onSelect(option)}
                  style={({ pressed }) => ({
                    flexDirection: "row",
                    alignItems: "center",
                    gap: 10,
                    minHeight: 52,
                    paddingVertical: 12,
                    opacity: pressed ? 0.6 : 1,
                  })}
                >
                  <MetadataOptionIcon option={option} />
                  <View style={{ flex: 1, minWidth: 0 }}>
                    <Text numberOfLines={1}>{option.label}</Text>
                    {option.description ? (
                      <Text color="muted" fontSize="xs">
                        {option.description}
                      </Text>
                    ) : null}
                  </View>
                  {selected ? (
                    <Ionicons name="checkmark" size={20} color={foreground} />
                  ) : null}
                </Pressable>
              );
            }}
          />
        </View>
      </KeyboardAvoidingView>
    </Modal>
  );
};
