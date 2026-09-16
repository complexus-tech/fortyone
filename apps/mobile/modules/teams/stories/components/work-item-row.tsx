import type { ReactNode } from "react";
import { memo } from "react";
import { Pressable, View } from "react-native";
import { Text } from "@/components/ui/Text";

type WorkItemRowProps = {
  title: string;
  subtitle?: string;
  leading: ReactNode;
  trailing?: ReactNode;
  accessibilityLabel?: string;
  accessibilityHint: string;
  onPress: () => void;
};

export const WorkItemRow = memo(function WorkItemRow({
  title,
  subtitle,
  leading,
  trailing,
  accessibilityLabel,
  accessibilityHint,
  onPress,
}: WorkItemRowProps) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={
        accessibilityLabel ?? [title, subtitle].filter(Boolean).join(", ")
      }
      accessibilityHint={accessibilityHint}
      onPress={onPress}
      className="active:bg-gray-50 dark:active:bg-dark-200"
      style={{
        flexDirection: "row",
        alignItems: "flex-start",
        gap: 10,
        paddingHorizontal: 20,
        paddingVertical: 11,
        minHeight: 52,
      }}
    >
      <View
        pointerEvents="none"
        style={{
          flexDirection: "row",
          alignItems: "center",
          gap: 8,
          minHeight: 23,
        }}
      >
        {leading}
      </View>
      <View style={{ flex: 1, minWidth: 0, gap: 3 }}>
        <Text
          numberOfLines={1}
          style={{ fontSize: 17, lineHeight: 23, fontWeight: "500" }}
        >
          {title}
        </Text>
        {subtitle ? (
          <Text
            color="muted"
            numberOfLines={1}
            style={{ fontSize: 13, lineHeight: 18, fontWeight: "600" }}
          >
            {subtitle}
          </Text>
        ) : null}
      </View>
      {trailing}
    </Pressable>
  );
});
