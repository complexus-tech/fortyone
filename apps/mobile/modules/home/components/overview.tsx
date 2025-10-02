import React from "react";
import { View } from "react-native";
import { StatCard } from "./stat-card";
import type { SFSymbol } from "expo-symbols";
import { colors } from "@/constants";
import { Text } from "@/components/ui";
import { StoriesSummary } from "../types";

export const Overview = ({ summary }: { summary?: StoriesSummary }) => {
  const overviewItems = [
    {
      count: summary?.closed,
      label: "Stories closed",
      icon: "checkmark.circle.fill",
      iconColor: colors.success,
    },
    {
      count: summary?.overdue,
      label: "Stories overdue",
      icon: "exclamationmark.circle.fill",
      iconColor: colors.danger,
    },
    {
      count: summary?.inProgress,
      label: "In progress",
      icon: "clock.fill",
      iconColor: colors.warning,
    },
    {
      count: summary?.assigned,
      label: "Assigned to you",
      icon: "person.crop.circle.dashed",
      iconColor: colors.gray.DEFAULT,
    },
  ];

  return (
    <>
      <Text color="muted" className="mb-4">
        Here&apos;s what&apos;s happening with your stories.
      </Text>
      <View className="mb-6 flex-row flex-wrap gap-3">
        {overviewItems.map((item) => (
          <View key={item.label} className="w-[48.5%]">
            <StatCard
              count={item.count}
              label={item.label}
              icon={item.icon as SFSymbol}
              iconColor={item.iconColor}
            />
          </View>
        ))}
      </View>
    </>
  );
};
