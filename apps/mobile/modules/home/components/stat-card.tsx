import type { ComponentProps } from "react";
import type { SFSymbol } from "expo-symbols";
import { StyleSheet, View } from "react-native";
import { SymbolView } from "expo-symbols";
import { Ionicons } from "@expo/vector-icons";
import { Text } from "@/components/ui";
import { useTheme } from "@/hooks";
import { colors, themeColors } from "@/constants/colors";

type StatCardProps = {
  count?: number;
  label: string;
  icon: ComponentProps<typeof Ionicons>["name"];
  systemImage: SFSymbol;
  attention?: boolean;
};

export const StatCard = ({
  count = 0,
  label,
  icon,
  systemImage,
  attention,
}: StatCardProps) => {
  const { resolvedTheme } = useTheme();
  const dark = resolvedTheme === "dark";
  const iconColor =
    attention && count > 0
      ? colors.danger
      : themeColors[dark ? "dark" : "light"].textMuted;

  return (
    <View
      accessible
      accessibilityLabel={`${count} ${label.toLowerCase()}`}
      style={{
        flex: 1,
        minWidth: 0,
        padding: 16,
        gap: 8,
        minHeight: 96,
        borderRadius: 12,
        borderWidth: StyleSheet.hairlineWidth,
        borderColor: themeColors[dark ? "dark" : "light"].border,
        backgroundColor: themeColors[dark ? "dark" : "light"].surfaceMuted,
      }}
    >
      <View
        style={{
          flexDirection: "row",
          alignItems: "center",
          justifyContent: "space-between",
          gap: 8,
        }}
      >
        <Text
          numberOfLines={1}
          ellipsizeMode="tail"
          style={{
            flexShrink: 1,
            fontSize: 26,
            lineHeight: 32,
            fontWeight: "500",
            fontVariant: ["tabular-nums"],
          }}
        >
          {count}
        </Text>
        <SymbolView
          name={systemImage}
          size={18}
          tintColor={iconColor}
          fallback={<Ionicons name={icon} size={18} color={iconColor} />}
        />
      </View>
      <Text
        color="muted"
        numberOfLines={1}
        ellipsizeMode="tail"
        style={{ fontSize: 15, lineHeight: 20, fontWeight: "400" }}
      >
        {label}
      </Text>
    </View>
  );
};
