import type { ReactNode } from "react";
import { View } from "react-native";
import { Text } from "@/components/ui";
import { GlassSurface } from "@/components/ui/glass-surface";

type StatCardProps = {
  count?: number;
  label: string;
  icon: ReactNode;
};

export const StatCard = ({ count = 0, label, icon }: StatCardProps) => {
  return (
    <GlassSurface
      cornerRadius={16}
      accessible
      accessibilityLabel={`${count} ${label.toLowerCase()}`}
      style={{
        flex: 1,
        minWidth: 0,
        paddingHorizontal: 16,
        paddingVertical: 12,
        gap: 4,
        minHeight: 80,
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
            fontWeight: "600",
            fontVariant: ["tabular-nums"],
          }}
        >
          {count}
        </Text>
        {icon}
      </View>
      <Text
        color="muted"
        numberOfLines={1}
        ellipsizeMode="tail"
        style={{ fontSize: 15, lineHeight: 20, fontWeight: "600" }}
      >
        {label}
      </Text>
    </GlassSurface>
  );
};
