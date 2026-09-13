import { useState } from "react";
import { Platform, Pressable, View } from "react-native";
import DateTimePicker, {
  DateTimePickerAndroid,
} from "@react-native-community/datetimepicker";
import { MaterialIcons } from "@expo/vector-icons";
import { format, isValid, parseISO } from "date-fns";
import { useTheme } from "@/hooks";
import { colors } from "@/constants";
import { Badge } from "./badge";
import { Text } from "./text";
import { Button } from "./button";
import { BottomSheetModal } from "./bottom-sheet-modal";

type DateFieldProps = {
  value: string | null | undefined;
  label: string;
  onChange: (value: Date | null) => void;
};

function dateValue(value: DateFieldProps["value"]) {
  const parsed = value ? parseISO(value.slice(0, 10)) : new Date();
  return isValid(parsed) ? parsed : new Date();
}

// Consumers serialize with date-fns formatISO(..., { representation: "date" }).
// Preserve the local calendar day instead of introducing a UTC offset.
function calendarDate(date: Date) {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate());
}

export function DateField({ value, label, onChange }: DateFieldProps) {
  const { resolvedTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [selection, setSelection] = useState(() => dateValue(value));
  const iconColor =
    resolvedTheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const open = () => {
    const current = dateValue(value);
    if (Platform.OS === "android") {
      DateTimePickerAndroid.open({
        value: current,
        mode: "date",
        positiveButton: { label: "Apply" },
        negativeButton: { label: "Cancel" },
        neutralButton: value ? { label: "Remove date" } : undefined,
        onValueChange: (_event, selected) => onChange(calendarDate(selected)),
        onNeutralButtonPress: () => onChange(null),
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
        onPress={open}
      >
        <Badge color="tertiary">
          <MaterialIcons name="calendar-today" size={16} color={iconColor} />
          <Text>
            {value
              ? format(dateValue(value), "MMM d")
              : `Add ${label.toLowerCase()}`}
          </Text>
        </Badge>
      </Pressable>
      {Platform.OS !== "android" && (
        <BottomSheetModal
          isOpen={isOpen}
          onClose={() => setIsOpen(false)}
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
            onValueChange={(_event, selected) => setSelection(selected)}
          />
          <View style={{ flexDirection: "row", gap: 12 }}>
            <Button
              fullWidth={false}
              className="flex-1"
              color="tertiary"
              onPress={() => setIsOpen(false)}
            >
              Cancel
            </Button>
            <Button
              fullWidth={false}
              className="flex-1"
              onPress={() => {
                onChange(calendarDate(selection));
                setIsOpen(false);
              }}
            >
              Apply
            </Button>
          </View>
          {value ? (
            <Button
              color="tertiary"
              onPress={() => {
                onChange(null);
                setIsOpen(false);
              }}
            >
              Remove date
            </Button>
          ) : null}
        </BottomSheetModal>
      )}
    </>
  );
}
