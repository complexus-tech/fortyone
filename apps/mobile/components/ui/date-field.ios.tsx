import type { DateFieldProps } from "./date-field.types";
import { useRef, useState } from "react";
import { Pressable, useColorScheme } from "react-native";
import { CalendarIcon } from "@/components/icons/calendar";
import {
  Button,
  DatePicker,
  Group,
  HStack,
  ProgressView,
  Spacer,
  Text,
  VStack,
} from "@expo/ui/swift-ui";
import {
  accessibilityLabel,
  buttonStyle,
  datePickerStyle,
  disabled as disabledModifier,
  font,
  foregroundStyle,
  frame,
  interactiveDismissDisabled,
  tint,
} from "@expo/ui/swift-ui/modifiers";
import { format } from "date-fns";
import { colors, themeColors } from "@/constants/colors";
import { Badge } from "./badge";
import { Text as UIText } from "./Text";
import { BottomSheetModal } from "./bottom-sheet-modal";
import { calendarDate, dateValue } from "./date-field-utils";

export function DateField({
  value,
  label,
  onChange,
  disabled = false,
}: DateFieldProps) {
  const dark = useColorScheme() === "dark";
  const [open, setOpen] = useState(false);
  const [selection, setSelection] = useState(() => dateValue(value));
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const pending = useRef(false);
  const foreground = themeColors[dark ? "dark" : "light"].foreground;
  const muted = themeColors[dark ? "dark" : "light"].textMuted;

  const close = () => {
    if (!pending.current) setOpen(false);
  };
  const commit = (next: Date | null) => {
    if (pending.current) return;
    pending.current = true;
    setSaving(true);
    setError(null);
    void Promise.resolve()
      .then(() => onChange(next))
      .then(() => setOpen(false))
      .catch((cause: unknown) => {
        setError(
          cause instanceof Error
            ? cause.message
            : "Could not update this date. Please try again.",
        );
      })
      .finally(() => {
        pending.current = false;
        setSaving(false);
      });
  };

  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`${label}: ${value ? format(dateValue(value), "PPP") : "not set"}`}
        accessibilityState={{ disabled: disabled || saving }}
        disabled={disabled || saving}
        onPress={() => {
          setSelection(dateValue(value));
          setError(null);
          setOpen(true);
        }}
        style={({ pressed }) => ({
          minHeight: 44,
          maxWidth: "100%",
          minWidth: 0,
          flexShrink: 1,
          justifyContent: "center",
          opacity: disabled || saving ? 0.5 : pressed ? 0.6 : 1,
        })}
      >
        <Badge
          color="tertiary"
          rounded="full"
          className="min-w-0 max-w-full shrink border-0 bg-accent px-[8px] py-[4px] dark:bg-accent-dark"
        >
          <CalendarIcon size={16} color={muted} />
          <UIText fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {value
              ? format(dateValue(value), "MMM d")
              : `Add ${label.toLowerCase()}`}
          </UIText>
        </Badge>
      </Pressable>
      <BottomSheetModal
        nativeContent
        isOpen={open}
        onClose={close}
        spacing={0}
        padding={{ leading: 16, trailing: 16, top: 16, bottom: 16 }}
      >
        <Group modifiers={[interactiveDismissDisabled(saving)]}>
          <VStack spacing={12} alignment="leading">
            <HStack spacing={12} modifiers={[frame({ minHeight: 44 })]}>
              <Text
                modifiers={[
                  font({ textStyle: "headline" }),
                  foregroundStyle(foreground),
                ]}
              >
                {label}
              </Text>
              <Spacer />
              {saving && (
                <ProgressView modifiers={[accessibilityLabel("Saving date")]} />
              )}
              <Button
                role="close"
                onPress={close}
                modifiers={[
                  accessibilityLabel(`Close ${label.toLowerCase()} picker`),
                  buttonStyle("glass"),
                  disabledModifier(saving),
                  frame({ width: 44, height: 44 }),
                ]}
              />
            </HStack>
            <DatePicker
              selection={selection}
              displayedComponents={["date"]}
              onDateChange={setSelection}
              modifiers={[
                datePickerStyle("graphical"),
                disabledModifier(saving),
                accessibilityLabel(`Select ${label.toLowerCase()}`),
                tint(colors.primary),
              ]}
            />
            {error && (
              <Text
                modifiers={[
                  foregroundStyle(dark ? colors.dangerTextDark : colors.danger),
                  accessibilityLabel(error),
                ]}
              >
                {error}
              </Text>
            )}
            <HStack spacing={12}>
              <Button
                label="Cancel"
                role="cancel"
                onPress={close}
                modifiers={[
                  buttonStyle("glass"),
                  disabledModifier(saving),
                  frame({ minHeight: 44 }),
                ]}
              />
              <Spacer />
              <Button
                label={saving ? "Applying…" : "Apply"}
                onPress={() => commit(calendarDate(selection))}
                modifiers={[
                  buttonStyle("glassProminent"),
                  tint(colors.primary),
                  foregroundStyle(colors.primaryForeground),
                  disabledModifier(saving),
                  frame({ minHeight: 44 }),
                ]}
              />
            </HStack>
            {value && (
              <Button
                label="Remove date"
                onPress={() => commit(null)}
                modifiers={[
                  buttonStyle("plain"),
                  foregroundStyle(foreground),
                  disabledModifier(saving),
                  frame({ minHeight: 44 }),
                  accessibilityLabel(`Remove ${label.toLowerCase()}`),
                ]}
              />
            )}
          </VStack>
        </Group>
      </BottomSheetModal>
    </>
  );
}
