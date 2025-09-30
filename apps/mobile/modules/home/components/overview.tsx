import React from "react";
import { View } from "react-native";
import { StatCard } from "./stat-card";

type OverviewProps = {
  stats?: {
    closed: number;
    overdue: number;
    inProgress: number;
    created: number;
    assigned: number;
  };
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

  const overviewItems = [
    {
      count: data.closed,
      label: "Stories closed",
      icon: "checkmark-circle-outline" as const,
    },
    {
      count: data.overdue,
      label: "Stories overdue",
      icon: "alert-circle-outline" as const,
    },
    {
      count: data.inProgress,
      label: "In progress",
      icon: "time-outline" as const,
    },
    {
      count: data.assigned,
      label: "Assigned to you",
      icon: "person-outline" as const,
    },
  ];

  return (
    <View className="mb-4 flex-row flex-wrap gap-3">
      {overviewItems.map((item) => (
        <View key={item.label} className="w-[48%]">
          <StatCard {...item} />
        </View>
      ))}
    </View>
  );
};
