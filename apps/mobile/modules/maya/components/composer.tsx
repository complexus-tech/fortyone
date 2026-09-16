import {
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TextInput,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";
import { WebIcon } from "@/components/icons/web-icon";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";
import { ComposerSurface } from "./composer-surface";

export function MayaComposer({
  value,
  attachments,
  picking,
  dictationStatus,
  dictationSeconds,
  onChange,
  busy,
  disabled,
  onSend,
  onStop,
  onVoice,
  onAttach,
  onRemoveAttachment,
  onDictation,
  onCancelDictation,
}: {
  value: string;
  attachments: readonly {
    id: string;
    uri: string;
    name: string;
    mediaType: string;
    size: number;
  }[];
  picking: boolean;
  dictationStatus: "idle" | "preparing" | "recording" | "transcribing";
  dictationSeconds: number;
  onChange: (text: string) => void;
  busy: boolean;
  disabled: boolean;
  onSend: () => void;
  onStop: () => void;
  onVoice: () => void;
  onAttach: () => void;
  onRemoveAttachment: (id: string) => void;
  onDictation: () => void;
  onCancelDictation: () => void;
}) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];
  const hasContent = Boolean(value.trim()) || attachments.length > 0;
  const dictating = dictationStatus !== "idle";
  const attachmentsDisabled = busy || disabled || picking || dictating;
  const primaryDisabled = disabled || (!busy && picking);
  const primaryLabel = busy
    ? "Stop response"
    : hasContent
      ? "Send message"
      : "Start voice conversation";
  const primaryIcon = busy ? "stop" : "arrow-up";
  const onPrimaryPress = busy ? onStop : hasContent ? onSend : onVoice;
  const elapsedSeconds = Math.max(0, Math.floor(dictationSeconds));
  const elapsed = `${Math.floor(elapsedSeconds / 60)}:${String(
    elapsedSeconds % 60,
  ).padStart(2, "0")}`;
  return (
    <View style={styles.container}>
      {attachments.length > 0 ? (
        <ScrollView
          horizontal
          showsHorizontalScrollIndicator={false}
          keyboardShouldPersistTaps="handled"
          contentContainerStyle={styles.attachments}
        >
          {attachments.map((attachment) => (
            <View
              key={attachment.id}
              style={[
                styles.attachment,
                {
                  backgroundColor: theme.surfaceMuted,
                  borderColor: theme.border,
                },
              ]}
            >
              <Text
                numberOfLines={1}
                ellipsizeMode="middle"
                style={[styles.attachmentName, { color: theme.foreground }]}
              >
                {attachment.name}
              </Text>
              <Pressable
                disabled={attachmentsDisabled}
                accessibilityRole="button"
                accessibilityLabel={`Remove ${attachment.name}`}
                accessibilityState={{ disabled: attachmentsDisabled }}
                onPress={() => onRemoveAttachment(attachment.id)}
                style={({ pressed }) => [
                  styles.control,
                  { opacity: attachmentsDisabled ? 0.35 : pressed ? 0.55 : 1 },
                ]}
              >
                <WebIcon name="close" size={15} color={theme.textSecondary} />
              </Pressable>
            </View>
          ))}
        </ScrollView>
      ) : null}
      {picking ? (
        <Text
          accessibilityLiveRegion="polite"
          style={[styles.picking, { color: theme.textSecondary }]}
        >
          Adding attachments…
        </Text>
      ) : null}
      <ComposerSurface>
        {dictating ? (
          <View style={styles.composer}>
            <Pressable
              disabled={disabled}
              accessibilityRole="button"
              accessibilityLabel="Cancel dictation"
              accessibilityState={{ disabled }}
              onPress={onCancelDictation}
              style={({ pressed }) => [
                styles.control,
                { opacity: disabled ? 0.35 : pressed ? 0.55 : 1 },
              ]}
            >
              <WebIcon
                key="dictation-cancel"
                name="close"
                size={20}
                color={theme.foreground}
              />
            </Pressable>
            <View style={styles.dictationStatus}>
              <Text
                accessibilityLiveRegion={
                  dictationStatus === "transcribing" ? "polite" : "none"
                }
                style={[styles.dictationLabel, { color: theme.foreground }]}
              >
                {dictationStatus === "recording"
                  ? `Recording ${elapsed}`
                  : dictationStatus === "preparing"
                    ? "Preparing microphone…"
                    : "Transcribing…"}
              </Text>
            </View>
            {dictationStatus === "recording" ? (
              <Pressable
                disabled={disabled}
                accessibilityRole="button"
                accessibilityLabel="Finish dictation and transcribe"
                accessibilityState={{ disabled }}
                onPress={onDictation}
                style={({ pressed }) => [
                  styles.control,
                  { opacity: disabled ? 0.35 : pressed ? 0.65 : 1 },
                ]}
              >
                <View
                  pointerEvents="none"
                  style={[
                    styles.primary,
                    { backgroundColor: theme.backgroundInverse },
                  ]}
                >
                  <WebIcon
                    name="check"
                    size={18}
                    color={theme.foregroundInverse}
                  />
                </View>
              </Pressable>
            ) : null}
          </View>
        ) : (
          <View style={styles.composer}>
            <Pressable
              disabled={attachmentsDisabled}
              accessibilityRole="button"
              accessibilityLabel="Add attachments"
              accessibilityState={{
                disabled: attachmentsDisabled,
                busy: picking,
              }}
              onPress={onAttach}
              style={({ pressed }) => [
                styles.control,
                { opacity: attachmentsDisabled ? 0.35 : pressed ? 0.55 : 1 },
              ]}
            >
              {/* Recreate the SVG instead of recycling dictation's cancel path. */}
              <WebIcon
                key="attachment-plus"
                name="plus"
                size={22}
                color={theme.foreground}
              />
            </Pressable>
            <TextInput
              accessibilityLabel="Message Maya"
              placeholder="Ask Maya"
              placeholderTextColor={theme.textMuted}
              value={value}
              onChangeText={onChange}
              multiline
              maxLength={12000}
              editable={!busy && !disabled}
              style={[styles.input, { color: theme.foreground }]}
              keyboardAppearance={resolvedTheme}
            />
            <Pressable
              disabled={attachmentsDisabled}
              accessibilityState={{ disabled: attachmentsDisabled }}
              accessibilityRole="button"
              accessibilityLabel="Dictate message"
              onPress={onDictation}
              style={({ pressed }) => [
                styles.control,
                { opacity: attachmentsDisabled ? 0.35 : pressed ? 0.55 : 1 },
              ]}
            >
              <Ionicons name="mic-outline" size={22} color={theme.foreground} />
            </Pressable>
            <Pressable
              disabled={primaryDisabled}
              accessibilityState={{ disabled: primaryDisabled }}
              accessibilityRole="button"
              accessibilityLabel={primaryLabel}
              onPress={onPrimaryPress}
              style={({ pressed }) => [
                styles.control,
                {
                  opacity: primaryDisabled ? 0.35 : pressed ? 0.65 : 1,
                },
              ]}
            >
              <View
                pointerEvents="none"
                style={[
                  styles.primary,
                  { backgroundColor: theme.backgroundInverse },
                ]}
              >
                {busy || hasContent ? (
                  <Ionicons
                    name={primaryIcon}
                    size={20}
                    color={theme.foregroundInverse}
                  />
                ) : (
                  <WebIcon
                    name="voice"
                    size={20}
                    color={theme.foregroundInverse}
                  />
                )}
              </View>
            </Pressable>
          </View>
        )}
      </ComposerSurface>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    gap: 8,
  },
  attachments: {
    gap: 8,
  },
  attachment: {
    flexDirection: "row",
    alignItems: "center",
    paddingLeft: 12,
    borderRadius: 12,
    borderWidth: StyleSheet.hairlineWidth,
  },
  attachmentName: {
    maxWidth: 160,
    fontSize: 15,
    lineHeight: 20,
  },
  picking: {
    paddingHorizontal: 12,
    fontSize: 15,
    lineHeight: 20,
  },
  composer: {
    minHeight: 48,
    flexDirection: "row",
    alignItems: "flex-end",
    paddingHorizontal: 2,
    paddingVertical: 2,
    gap: 2,
  },
  dictationStatus: {
    flex: 1,
    minHeight: 44,
    justifyContent: "center",
  },
  dictationLabel: {
    fontSize: 17,
    lineHeight: 24,
    fontVariant: ["tabular-nums"],
  },
  input: {
    flex: 1,
    fontSize: 17,
    lineHeight: 24,
    paddingHorizontal: 0,
    paddingTop: 8,
    paddingBottom: 12,
    minHeight: 44,
    maxHeight: 120,
    includeFontPadding: false,
    textAlignVertical: "top",
  },
  control: {
    width: 44,
    height: 44,
    alignItems: "center",
    justifyContent: "center",
    borderRadius: 22,
  },
  primary: {
    width: 28,
    height: 28,
    borderRadius: 14,
    alignItems: "center",
    justifyContent: "center",
  },
});
