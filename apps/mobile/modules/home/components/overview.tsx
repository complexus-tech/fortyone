import { WebIcon } from "@/components/icons/web-icon";
import { StatusIcon } from "@/components/icons/status";
import { useTheme } from "@/hooks/theme";
import { colors, themeColors } from "@/constants/colors";
import { View, useWindowDimensions } from "react-native";
import { QueryState } from "@/components/ui/query-state";
import { useOverviewStats } from "@/modules/home/hooks/use-overview-stats";
import { StatCard } from "./stat-card";
import { OverviewSkeleton } from "./overview-skeleton";

export const Overview = () => {
  const { resolvedTheme } = useTheme();
  const iconColor = themeColors[resolvedTheme].foreground;
  const { width, fontScale } = useWindowDimensions();
  const stacked = fontScale >= 1.4 || width < 340;
  const { data: summary, isPending, error, refetch } = useOverviewStats();

  return (
    <View>
      {isPending ? (
        <OverviewSkeleton />
      ) : error ? (
        <QueryState
          title="Could not load your overview"
          message={error.message}
          onRetry={() => {
            void refetch();
          }}
        />
      ) : (
        <View style={{ paddingHorizontal: 20, gap: 10 }}>
          <View style={{ flexDirection: stacked ? "column" : "row", gap: 10 }}>
            <StatCard
              count={summary?.assigned}
              label="Assigned to you"
              icon={<WebIcon name="user" size={18} color={iconColor} />}
            />
            <StatCard
              count={summary?.inProgress}
              label="In progress"
              icon={
                <StatusIcon
                  category="started"
                  size={18}
                  color={colors.warning}
                />
              }
            />
          </View>
          <View style={{ flexDirection: stacked ? "column" : "row", gap: 10 }}>
            <StatCard
              count={summary?.overdue}
              label="Overdue"
              icon={
                <WebIcon
                  name="clock"
                  size={18}
                  color={
                    resolvedTheme === "dark"
                      ? colors.dangerTextDark
                      : colors.danger
                  }
                />
              }
            />
            <StatCard
              count={summary?.closed}
              label="Closed"
              icon={
                <StatusIcon
                  category="completed"
                  size={18}
                  color={colors.success}
                />
              }
            />
          </View>
        </View>
      )}
    </View>
  );
};
