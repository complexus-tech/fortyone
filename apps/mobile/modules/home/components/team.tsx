import type { Team as TeamType } from "@/modules/teams/types";
import { View, Pressable } from "react-native";
import { WebIcon } from "@/components/icons/web-icon";
import { useRouter } from "expo-router";
import { Text } from "@/components/ui";
import { themeColors } from "@/constants/colors";
import { useTheme } from "@/hooks";

export const Team = ({ id, name, color }: TeamType) => {
  const router = useRouter();
  const { resolvedTheme } = useTheme();
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={`Open ${name} team`}
      onPress={() => router.push(`/teams/${id}`)}
      className="active:bg-gray-50 dark:active:bg-dark-200"
      style={{
        flexDirection: "row",
        alignItems: "center",
        minHeight: 56,
        paddingHorizontal: 20,
        paddingVertical: 12,
        gap: 12,
      }}
    >
      <View
        style={{
          width: 10,
          height: 10,
          borderRadius: 3,
          backgroundColor: color,
        }}
      />
      <Text
        numberOfLines={1}
        style={{ flex: 1, fontSize: 17, lineHeight: 23, fontWeight: "500" }}
      >
        {name}
      </Text>
      <WebIcon
        name="chevronRight"
        size={15}
        color={themeColors[resolvedTheme].icon}
      />
    </Pressable>
  );
};
