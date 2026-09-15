import type { DateFieldProps } from "./date-field.types";
import { useRef, useState } from "react";
import { Alert, Platform, Pressable, View } from "react-native";
import DateTimePicker, {
  DateTimePickerAndroid,
} from "@react-native-community/datetimepicker";
import { CalendarIcon } from "@/components/icons/calendar";
import { format } from "date-fns";
import { calendarDate, dateValue } from "./date-field-utils";
import { useTheme } from "@/hooks";
import { colors, themeColors } from "@/constants/colors";
import { Badge } from "./badge";
import { Text } from "./Text";
import { Button } from "./Button";
import { BottomSheetModal } from "./bottom-sheet-modal";

export function DateField({
  value,
  label,
  onChange,
  disabled = false,
}: DateFieldProps) {
  const { resolvedTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selection, setSelection] = useState(() => dateValue(value));
  const [saving, setSaving] = useState(false);
  const pending = useRef(false);
  const iconColor = themeColors[resolvedTheme].textMuted;

  const commit = (next: Date | null) => {
    if (pending.current) return;
    pending.current = true;
    setSaving(true);
    void Promise.resolve()
      .then(() => onChange(next))
      .then(() => setIsOpen(false))
      .catch((cause: unknown) => {
        Alert.alert(
          `Could not update ${label.toLowerCase()}`,
          cause instanceof Error ? cause.message : "Please try again.",
        );
      })
      .finally(() => {
        pending.current = false;
        setSaving(false);
      });
  };

  const open = () => {
    if (disabled || pending.current) return;
    const current = dateValue(value);
    if (Platform.OS === "android") {
      DateTimePickerAndroid.open({
        value: current,
        mode: "date",
        positiveButton: { label: "Apply" },
        negativeButton: { label: "Cancel" },
        neutralButton: value ? { label: "Remove date" } : undefined,
        onValueChange: (_event, selected) => commit(calendarDate(selected)),
        onNeutralButtonPress: () => commit(null),
      });
      return;
    }
    setSelection(current);
    setIsOpen(true);
  };

  return (
    <>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`${label}: ${value ? format(dateValue(value), "PPP") : "not set"}`}
        accessibilityState={{ disabled: disabled || saving }}
        disabled={disabled || saving}
        onPress={open}
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
          <CalendarIcon size={16} color={iconColor} />
          <Text fontSize="sm" numberOfLines={1} style={{ flexShrink: 1 }}>
            {value
              ? format(dateValue(value), "MMM d")
              : `Add ${label.toLowerCase()}`}
          </Text>
        </Badge>
      </Pressable>
      {Platform.OS !== "android" && (
        <BottomSheetModal
          isOpen={isOpen}
          onClose={() => {
            if (!pending.current) setIsOpen(false);
          }}
          spacing={16}
        >
          <Text fontSize="lg" fontWeight="semibold">
            {label}
          </Text>
          <DateTimePicker
            value={selection}
            mode="date"
            display="inline"
            themeVariant={resolvedTheme}
            accentColor={colors.primary}
            disabled={saving}
            onValueChange={(_event, selected) => setSelection(selected)}
          />
          <View style={{ flexDirection: "row", gap: 12 }}>
            <Button
              fullWidth={false}
              className="flex-1"
              color="tertiary"
              disabled={saving}
              onPress={() => setIsOpen(false)}
            >
              Cancel
            </Button>
            <Button
              fullWidth={false}
              className="flex-1"
              disabled={saving}
              onPress={() => commit(calendarDate(selection))}
            >
              {saving ? "Applying…" : "Apply"}
            </Button>
          </View>
          {value ? (
            <Button
              color="tertiary"
              disabled={saving}
              onPress={() => commit(null)}
            >
              Remove date
            </Button>
          ) : null}
        </BottomSheetModal>
      )}
    </>
  );
}
