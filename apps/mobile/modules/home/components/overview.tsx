import { WebIcon } from "@/components/icons/web-icon";
import { StatusIcon } from "@/components/icons/status";
import { useTheme } from "@/hooks/theme";
import { themeColors } from "@/constants/colors";
import { Pressable, View, useWindowDimensions } from "react-native";
import { useRouter } from "expo-router";
import { Text } from "@/components/ui";
import { ScreenHeader } from "@/components/ui/screen-header";
import { QueryState } from "@/components/ui/query-state";
import { useOverviewStats } from "@/modules/home/hooks/use-overview-stats";
import { StatCard } from "./stat-card";
import { OverviewSkeleton } from "./overview-skeleton";

export const Overview = () => {
  const router = useRouter();
  const { resolvedTheme } = useTheme();
  const iconColor = themeColors[resolvedTheme].foreground;
  const { width, fontScale } = useWindowDimensions();
  const stacked = fontScale >= 1.4 || width < 340;
  const { data: summary, isPending, error, refetch } = useOverviewStats();

  return (
    <View>
      <ScreenHeader
        title="Your work"
        trailing={
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="View My Work"
            onPress={() => router.push("/my-work")}
            style={{ minHeight: 44, justifyContent: "center", paddingLeft: 12 }}
          >
            <Text
              color="muted"
              numberOfLines={1}
              style={{ fontSize: 15, lineHeight: 20 }}
            >
              View all
            </Text>
          </Pressable>
        }
      />
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
                <StatusIcon category="started" size={18} color={iconColor} />
              }
            />
          </View>
          <View style={{ flexDirection: stacked ? "column" : "row", gap: 10 }}>
            <StatCard
              count={summary?.overdue}
              label="Overdue"
              icon={<WebIcon name="clock" size={18} color={iconColor} />}
            />
            <StatCard
              count={summary?.closed}
              label="Closed"
              icon={
                <StatusIcon category="completed" size={18} color={iconColor} />
              }
            />
          </View>
        </View>
      )}
    </View>
  );
};
