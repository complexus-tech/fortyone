import type { TabsControlProps } from "./tabs-control.types";
import { Pressable, View } from "react-native";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks/theme";
import { Text } from "./Text";

export function TabsControl({
  options,
  value,
  onValueChange,
  labelSize = 15,
  accessibilityLabel,
}: TabsControlProps) {
  const { resolvedTheme } = useTheme();
  const theme = themeColors[resolvedTheme];

  return (
    <View
      accessibilityRole="tablist"
      accessibilityLabel={accessibilityLabel}
      style={{ flexDirection: "row", gap: 4 }}
    >
      {options.map((option) => {
        const selected = value === option.value;

        return (
          <Pressable
            key={option.value}
            accessibilityRole="tab"
            accessibilityState={{ selected }}
            onPress={() => {
              if (!selected) onValueChange(option.value);
            }}
            className="min-h-[32px] items-center justify-center rounded-full px-[14px] py-[4px]"
            style={({ pressed }) => ({
              backgroundColor: selected
                ? resolvedTheme === "dark"
                  ? theme.surfaceElevated
                  : theme.surfaceMuted
                : "transparent",
              opacity: pressed ? 0.65 : 1,
            })}
          >
            <Text
              fontWeight="semibold"
              style={{
                fontSize: labelSize,
                lineHeight: labelSize === 14 ? 19 : 20,
                color: selected ? theme.foreground : theme.textMuted,
              }}
            >
              {option.label}
            </Text>
          </Pressable>
        );
      })}
    </View>
  );
}
