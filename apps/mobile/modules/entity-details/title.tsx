import type { ReactNode } from "react";
import { useRef, useState } from "react";
import { themeColors } from "@/constants/colors";
import {
  ActivityIndicator,
  KeyboardAvoidingView,
  Modal,
  Platform,
  Pressable,
  TextInput,
  View,
} from "react-native";
import { SafeAreaProvider, SafeAreaView } from "react-native-safe-area-context";
import { Text, Col, Row, IconButton } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { useDraft } from "@/components/rich-text/use-draft";
import { DiscardDraftButton } from "@/components/rich-text/draft-recovery";
import { useTheme } from "@/hooks";

type TitleDraft = { title: string; baseline: string };
const isTitleDraft = (value: unknown): value is TitleDraft =>
  Boolean(
    value &&
      typeof value === "object" &&
      typeof (value as TitleDraft).title === "string" &&
      typeof (value as TitleDraft).baseline === "string",
  );

const EditTitle = ({
  draftKey,
  title,
  onSave,
  onClose,
}: {
  draftKey: string;
  title: string;
  onSave: (title: string, baseline: string) => Promise<void>;
  onClose: () => void;
}) => {
  const { resolvedTheme } = useTheme();
  const draft = useDraft(draftKey, { title, baseline: title }, isTitleDraft);
  const [saving, setSaving] = useState(false);
  const pending = useRef(false);
  const [error, setError] = useState<string | null>(null);
  const save = async () => {
    if (pending.current || !draft.ready || !draft.value.title.trim()) return;
    pending.current = true;
    setSaving(true);
    setError(null);
    try {
      await onSave(draft.value.title.trim(), draft.value.baseline);
      await draft.clear();
      onClose();
    } catch (cause) {
      setError(
        cause instanceof Error
          ? cause.message
          : "Your title was not saved. Try again.",
      );
    }
    pending.current = false;
    setSaving(false);
  };
  return (
    <Modal
      visible
      presentationStyle="pageSheet"
      animationType="slide"
      onRequestClose={() => {
        if (!saving) onClose();
      }}
    >
      <SafeAreaProvider>
        <SafeAreaView
          style={{
            flex: 1,
            backgroundColor: themeColors[resolvedTheme].background,
          }}
        >
          <KeyboardAvoidingView
            behavior={Platform.OS === "ios" ? "padding" : undefined}
            style={{ flex: 1, padding: 20 }}
          >
            <Row
              justify="between"
              align="center"
              className="mb-[16px] gap-[12px]"
            >
              <IconButton
                icon="close"
                label="Close title editor"
                disabled={saving}
                onPress={onClose}
                style={{
                  borderRadius: 22,
                  backgroundColor: themeColors[resolvedTheme].surfaceMuted,
                }}
              />
              <Text fontWeight="bold" className="flex-1 text-center">
                Edit title
              </Text>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={saving ? "Saving title" : "Save title"}
                accessibilityState={{
                  disabled: !draft.ready || !draft.value.title.trim() || saving,
                  busy: saving,
                }}
                disabled={!draft.ready || !draft.value.title.trim() || saving}
                onPress={() => void save()}
                className="size-[44px] items-center justify-center rounded-full bg-gray-50 dark:bg-dark-100 active:opacity-60"
              >
                {saving ? (
                  <ActivityIndicator accessibilityLabel="Saving title" />
                ) : (
                  <Ionicons
                    name="arrow-up"
                    size={24}
                    color={
                      !draft.ready || !draft.value.title.trim()
                        ? themeColors[resolvedTheme].textDisabled
                        : themeColors[resolvedTheme].foreground
                    }
                  />
                )}
              </Pressable>
            </Row>
            {!draft.ready && (
              <ActivityIndicator accessibilityLabel="Restoring title draft" />
            )}
            {draft.ready && (
              <TextInput
                accessibilityLabel="Title"
                autoFocus
                multiline
                value={draft.value.title}
                editable={!saving}
                onChangeText={(title) =>
                  draft.update({ ...draft.valueRef.current, title })
                }
                style={{
                  fontSize: 28,
                  lineHeight: 35,
                  fontWeight: "700",
                  minHeight: 120,
                  color: themeColors[resolvedTheme].foreground,
                }}
              />
            )}
            {(error || draft.error) && (
              <Text color="danger" accessibilityRole="alert">
                {error || draft.error}
              </Text>
            )}
            <DiscardDraftButton
              onDiscard={async () => {
                await draft.clear();
                onClose();
              }}
            />
            <View style={{ marginTop: 16 }}>
              <Text color="muted" fontSize="sm">
                Changes are kept as a draft on this device.
              </Text>
            </View>
          </KeyboardAvoidingView>
        </SafeAreaView>
      </SafeAreaProvider>
    </Modal>
  );
};

export function DetailTitle({
  children,
  title,
  draftKey,
  onSave,
}: {
  children?: ReactNode;
  title: string;
  draftKey: string;
  onSave?: (title: string, baseline: string) => Promise<void>;
}) {
  const [editing, setEditing] = useState(false);
  return (
    <Col asContainer className="pt-[4px]" align="stretch">
      {children}
      <Pressable
        accessibilityRole={onSave ? "button" : "header"}
        accessibilityLabel={onSave ? `Edit title: ${title}` : title}
        disabled={!onSave}
        onPress={() => setEditing(true)}
      >
        <Text
          fontSize="2xl"
          fontWeight="bold"
          style={{ fontSize: 28, lineHeight: 35, letterSpacing: -0.5 }}
        >
          {title}
        </Text>
      </Pressable>
      {editing && onSave ? (
        <EditTitle
          draftKey={draftKey}
          title={title}
          onSave={onSave}
          onClose={() => setEditing(false)}
        />
      ) : null}
    </Col>
  );
}
