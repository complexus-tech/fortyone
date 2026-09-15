import type { ReactNode } from "react";
import { useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Keyboard,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  TextInput,
  View,
  useColorScheme,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Ionicons } from "@expo/vector-icons";
import { IconButton, Text } from "@/components/ui";
import { GlassIconButton } from "@/components/ui/glass-icon-button";
import { themeColors } from "@/constants/colors";
import { filterPropertyOptions, performPropertyChange } from "./picker-utils";

export type PropertyOption = {
  id: string;
  label: string;
  description?: string;
  icon?: ReactNode;
};
type Props = {
  title: string;
  trigger: ReactNode;
  options: PropertyOption[];
  selectedIds: string[];
  onSelect: (id: string) => Promise<void>;
  clearLabel?: string;
  onClear?: () => Promise<void>;
  onCreate?: (name: string) => Promise<void>;
  multiple?: boolean;
  searchable?: boolean;
  disabled?: boolean;
  loading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
};

/** The native modal owns its keyboard, scroll viewport and dismissal lifecycle. */
export const PropertyBottomSheet = ({
  title,
  trigger,
  options,
  selectedIds,
  onSelect,
  clearLabel,
  onClear,
  onCreate,
  multiple = false,
  searchable = true,
  disabled = false,
  loading = false,
  error,
  onRetry,
}: Props) => {
  const dark = useColorScheme() === "dark";
  const [open, setOpen] = useState(false);
  const [query, setQuery] = useState("");
  const [saving, setSaving] = useState(false);
  const [actionError, setActionError] = useState<string | null>(null);
  const pending = useRef(false);
  const muted = themeColors[dark ? "dark" : "light"].textMuted;
  const filtered = filterPropertyOptions(options, query);
  const close = () => {
    if (!pending.current) {
      Keyboard.dismiss();
      setOpen(false);
    }
  };
  const select = (action: () => Promise<void>, keepOpen = multiple) => {
    if (pending.current) return;
    pending.current = true;
    setSaving(true);
    setActionError(null);
    void performPropertyChange(action, {
      onSuccess: () => {
        if (!keepOpen) {
          Keyboard.dismiss();
          setOpen(false);
        }
      },
      onError: setActionError,
      onSettled: () => {
        pending.current = false;
        setSaving(false);
      },
    });
  };
  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`Change ${title.toLowerCase()}`}
        accessibilityState={{ disabled }}
        disabled={disabled}
        onPress={() => {
          setQuery("");
          setActionError(null);
          setOpen(true);
        }}
        style={({ pressed }) => ({
          minHeight: 44,
          maxWidth: "100%",
          minWidth: 0,
          flexShrink: 1,
          justifyContent: "center",
          opacity: disabled ? 0.5 : pressed ? 0.6 : 1,
        })}
      >
        {trigger}
      </Pressable>
      <Modal
        visible={open}
        animationType="slide"
        presentationStyle="pageSheet"
        allowSwipeDismissal={!saving}
        onRequestClose={close}
        onDismiss={() => {
          setOpen(false);
          setQuery("");
          setActionError(null);
        }}
      >
        <SafeAreaProvider>
          <SafeAreaView
            edges={["top", "bottom", "left", "right"]}
            style={{
              flex: 1,
              backgroundColor: themeColors[dark ? "dark" : "light"].surface,
            }}
          >
            <KeyboardAvoidingView
              behavior={Platform.OS === "ios" ? "padding" : undefined}
              style={{ flex: 1 }}
            >
              <View
                style={{
                  flexDirection: "row",
                  alignItems: "center",
                  paddingHorizontal: 20,
                  paddingVertical: 8,
                }}
              >
                <View style={{ width: 44 }}>
                  {saving && (
                    <ActivityIndicator accessibilityLabel="Saving property" />
                  )}
                </View>
                <Text
                  accessibilityRole="header"
                  fontWeight="semibold"
                  style={{ flex: 1, textAlign: "center" }}
                >
                  {title}
                </Text>
                <GlassIconButton
                  icon="close"
                  systemImage="xmark"
                  label={`Close ${title.toLowerCase()} picker`}
                  onPress={close}
                  disabled={saving}
                />
              </View>
              {searchable && (
                <View
                  style={{
                    flexDirection: "row",
                    alignItems: "center",
                    marginHorizontal: 20,
                    paddingLeft: 12,
                    borderRadius: 12,
                    backgroundColor:
                      themeColors[dark ? "dark" : "light"].surfaceMuted,
                  }}
                >
                  <Ionicons name="search" size={20} color={muted} />
                  <TextInput
                    accessibilityLabel={`Search ${title.toLowerCase()}`}
                    placeholder={`Search ${title.toLowerCase()}…`}
                    placeholderTextColor={muted}
                    value={query}
                    onChangeText={setQuery}
                    autoCorrect={false}
                    autoCapitalize="none"
                    returnKeyType="search"
                    style={{
                      flex: 1,
                      minHeight: 48,
                      paddingHorizontal: 10,
                      fontSize: 16,
                      color: themeColors[dark ? "dark" : "light"].foreground,
                    }}
                  />
                  {query.length > 0 && (
                    <IconButton
                      icon="close-circle"
                      label="Clear search"
                      onPress={() => setQuery("")}
                    />
                  )}
                </View>
              )}
              {(actionError || error) && (
                <View style={{ paddingHorizontal: 20, paddingVertical: 12 }}>
                  <Text color="danger" accessibilityRole="alert">
                    {actionError || error?.message}
                  </Text>
                  {error && onRetry && (
                    <Text
                      onPress={onRetry}
                      style={{ minHeight: 44, paddingTop: 12 }}
                    >
                      Try again
                    </Text>
                  )}
                </View>
              )}
              <FlatList
                style={{ flex: 1 }}
                data={filtered}
                keyExtractor={(item) => item.id}
                keyboardShouldPersistTaps="handled"
                keyboardDismissMode="on-drag"
                initialNumToRender={12}
                maxToRenderPerBatch={12}
                windowSize={5}
                contentContainerStyle={{ paddingTop: 8, paddingBottom: 20 }}
                extraData={[selectedIds, saving]}
                renderItem={({ item }) => (
                  <Pressable
                    accessibilityRole={multiple ? "checkbox" : "radio"}
                    accessibilityLabel={item.label}
                    accessibilityState={{
                      checked: selectedIds.includes(item.id),
                      disabled: saving,
                    }}
                    disabled={saving}
                    onPress={() => select(() => onSelect(item.id))}
                    style={({ pressed }) => ({
                      flexDirection: "row",
                      alignItems: "center",
                      minHeight: 52,
                      paddingHorizontal: 20,
                      paddingVertical: 12,
                      gap: 10,
                      opacity: saving ? 0.5 : pressed ? 0.6 : 1,
                    })}
                  >
                    {item.icon}
                    <View style={{ flex: 1, minWidth: 0 }}>
                      <Text numberOfLines={1} ellipsizeMode="tail">
                        {item.label}
                      </Text>
                      {item.description && (
                        <Text
                          color="muted"
                          fontSize="xs"
                          numberOfLines={1}
                          ellipsizeMode="tail"
                        >
                          {item.description}
                        </Text>
                      )}
                    </View>
                    {selectedIds.includes(item.id) && (
                      <Ionicons
                        name="checkmark"
                        size={20}
                        color={themeColors[dark ? "dark" : "light"].foreground}
                      />
                    )}
                  </Pressable>
                )}
                ListHeaderComponent={
                  onClear ? (
                    <Pressable
                      accessibilityRole="radio"
                      accessibilityLabel={clearLabel}
                      accessibilityState={{
                        checked: selectedIds.length === 0,
                        disabled: saving,
                      }}
                      disabled={saving}
                      onPress={() => select(onClear, false)}
                      style={{
                        minHeight: 52,
                        flexDirection: "row",
                        alignItems: "center",
                        paddingHorizontal: 20,
                        gap: 10,
                      }}
                    >
                      <Text color="muted" style={{ flex: 1 }}>
                        {clearLabel}
                      </Text>
                      {selectedIds.length === 0 && (
                        <Ionicons name="checkmark" size={20} color={muted} />
                      )}
                    </Pressable>
                  ) : null
                }
                ListEmptyComponent={
                  loading ? (
                    <ActivityIndicator
                      style={{ padding: 24 }}
                      accessibilityLabel={`Loading ${title.toLowerCase()}`}
                    />
                  ) : !error ? (
                    <Text
                      color="muted"
                      style={{ padding: 20, textAlign: "center" }}
                    >
                      {query.trim()
                        ? "No matching results"
                        : "No options available"}
                    </Text>
                  ) : null
                }
                ListFooterComponent={
                  onCreate &&
                  query.trim() &&
                  filtered.length === 0 &&
                  !loading &&
                  !error ? (
                    <Pressable
                      accessibilityRole="button"
                      disabled={saving}
                      onPress={() => select(() => onCreate(query.trim()), true)}
                      style={{
                        minHeight: 52,
                        paddingHorizontal: 20,
                        flexDirection: "row",
                        alignItems: "center",
                        gap: 10,
                      }}
                    >
                      <Ionicons name="add" size={20} color={muted} />
                      <Text>{`Create “${query.trim()}”`}</Text>
                    </Pressable>
                  ) : null
                }
              />
            </KeyboardAvoidingView>
          </SafeAreaView>
        </SafeAreaProvider>
      </Modal>
    </>
  );
};
