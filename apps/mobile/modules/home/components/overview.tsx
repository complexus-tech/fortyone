import React from "react";
import { View } from "react-native";
import { StatCard } from "./stat-card";
import type { SFSymbol } from "expo-symbols";
import { colors } from "@/constants";
import { Text } from "@/components/ui";

type OverviewProps = {
  stats?: {
    closed: number;
    overdue: number;
    inProgress: number;
    created: number;
    assigned: number;
  };
};

type ItemsProps = {
  count: number;
  label: string;
  icon: SFSymbol;
  iconColor?: string;
};

export const Overview = ({ stats }: OverviewProps) => {
  const defaultStats = {
    closed: 0,
    overdue: 0,
    inProgress: 0,
    created: 0,
    assigned: 0,
  };

  const data = stats || defaultStats;

  const overviewItems: ItemsProps[] = [
    {
      count: data.closed,
      label: "Stories closed",
      icon: "checkmark.circle.fill",
      iconColor: colors.success,
    },
    {
      count: data.overdue,
      label: "Stories overdue",
      icon: "exclamationmark.circle.fill",
      iconColor: colors.danger,
    },
    {
      count: data.inProgress,
      label: "In progress",
      icon: "clock.fill",
      iconColor: colors.warning,
    },
    {
      count: data.assigned,
      label: "Assigned to you",
      icon: "person.circle.fill",
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
            <StatCard {...item} />
          </View>
        ))}
      </View>
    </>
  );
};
