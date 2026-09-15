import type { StoryActivity } from "@/modules/stories/types";
import type { ActivityDisplayContext } from "./activity-display";
import { View } from "react-native";
import { Text } from "@/components/ui";
import {
  formatActivityTimestamp,
  getActivityDisplay,
} from "./activity-display";

export const ActivityItem = ({
  context,
  isTimeShown = false,
  ...activity
}: StoryActivity & {
  context: ActivityDisplayContext;
  isTimeShown?: boolean;
}) => {
  if (activity.field === "completed_at") return null;

  const { actor, message } = getActivityDisplay(activity, context);
  const timestamp = isTimeShown
    ? formatActivityTimestamp(activity.createdAt)
    : null;

  return (
    <View className="flex-row items-start gap-[10px] py-[10px]">
      <View
        accessible={false}
        className="mt-[8px] h-[8px] w-[8px] rounded-full border border-gray-300 dark:border-gray-400"
      />
      <View className="min-w-0 flex-1 gap-[4px]">
        <Text color="muted" numberOfLines={2} ellipsizeMode="tail">
          <Text fontWeight="medium">{actor.name}</Text>
          {` ${message}`}
        </Text>
        {timestamp ? (
          <Text color="muted" fontSize="xs">
            {timestamp}
          </Text>
        ) : null}
      </View>
    </View>
  );
};
