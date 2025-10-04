import React from "react";
import { View } from "react-native";
import { StatCard } from "./stat-card";
import { OverviewSkeleton } from "./overview-skeleton";
import type { SFSymbol } from "expo-symbols";
import { colors } from "@/constants";
import { Col, Text } from "@/components/ui";
import { useOverviewStats } from "@/modules/home/hooks/use-overview-stats";
import { useTerminology } from "@/hooks";

export const Overview = () => {
  const { getTermDisplay } = useTerminology();
  const { data: summary, isPending } = useOverviewStats();
  const storyTerm = getTermDisplay("storyTerm", {
    variant: "plural",
    capitalize: true,
  });
  if (isPending) {
    return <OverviewSkeleton />;
  }
  const overviewItems = [
    {
      count: summary?.closed,
      label: `${storyTerm} closed`,
      icon: "checkmark.circle.fill",
      iconColor: colors.success,
    },
    {
      count: summary?.overdue,
      label: `${storyTerm} overdue`,
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
    <Col asContainer>
      <Text color="muted" className="mb-4">
        Here&apos;s what&apos;s happening with your{" "}
        {getTermDisplay("storyTerm", { variant: "plural" })}.
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
    </Col>
  );
};
