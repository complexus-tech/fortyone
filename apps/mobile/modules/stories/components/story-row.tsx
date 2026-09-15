import type { Member } from "@/types";
import type { Status } from "@/types/statuses";
import type { DisplayColumn } from "@/types/stories-view-options";
import type { Team } from "@/modules/teams/types";
import type { Story } from "../types";
import { memo, useRef } from "react";
import { View, useColorScheme } from "react-native";
import { useRouter } from "expo-router";
import { Avatar } from "@/components/ui/Avatar";
import { Text } from "@/components/ui/Text";
import { PriorityIcon, StatusIcon } from "@/components/icons";
import { colors, themeColors } from "@/constants/colors";
import { SwipeableRow } from "@/components/ui/swipeable-row";
import { useUpdateStoryMutation } from "@/modules/story/hooks/use-update-story-mutation";
import { useAuthStore } from "@/store/auth";
import { canCompleteStory } from "../utils/quick-actions";

type StoryRowProps = {
  story: Story;
  visibleColumns: DisplayColumn[];
  status?: Status;
  completionStatus?: Status;
  assignee?: Member;
  team?: Team;
  statusInGroupHeader?: boolean;
};

export const StoryRow = memo(function StoryRow({
  story,
  visibleColumns,
  status,
  completionStatus,
  assignee,
  team,
  statusInGroupHeader = false,
}: StoryRowProps) {
  const router = useRouter();
  const dark = useColorScheme() === "dark";
  const updateMutation = useUpdateStoryMutation();
  const completing = useRef(false);
  const sessionEpoch = useAuthStore((state) => state.sessionEpoch);
  const canComplete = canCompleteStory(story, status, completionStatus);
  const completeStory = () => {
    if (
      !canComplete ||
      !completionStatus ||
      completing.current ||
      useAuthStore.getState().sessionEpoch !== sessionEpoch
    )
      return;
    completing.current = true;
    updateMutation.mutate(
      { storyId: story.id, payload: { statusId: completionStatus.id } },
      {
        onSettled: () => {
          completing.current = false;
        },
      },
    );
  };
  const showStatus = visibleColumns.includes("Status");
  const showPriority = visibleColumns.includes("Priority");
  const showAssignee = visibleColumns.includes("Assignee");
  const metadata = [
    visibleColumns.includes("ID")
      ? `${team?.code ? `${team.code}-` : "#"}${story.sequenceId}`
      : null,
    showStatus && !statusInGroupHeader ? status?.name || "No status" : null,
  ]
    .filter(Boolean)
    .join(" · ");

  return (
    <SwipeableRow
      action={
        canComplete
          ? {
              label: "Complete",
              accessibilityLabel: `Mark as ${completionStatus?.name ?? "completed"}`,
              icon: "checkmark-circle-outline",
              backgroundColor: colors.success,
              foregroundColor: colors.black,
              onPress: completeStory,
              disabled: updateMutation.isPending,
            }
          : undefined
      }
      accessibilityRole="button"
      accessibilityLabel={[
        story.title,
        showStatus ? status?.name : null,
        showPriority ? story.priority : null,
        showAssignee
          ? assignee?.fullName || assignee?.username || "Unassigned"
          : null,
      ]
        .filter(Boolean)
        .join(", ")}
      accessibilityHint="Open task"
      className="active:bg-gray-50 dark:active:bg-dark-200"
      style={{
        flexDirection: "row",
        alignItems: "flex-start",
        gap: 10,
        paddingHorizontal: 20,
        paddingVertical: 11,
        minHeight: 52,
      }}
      onPress={() => router.push(`/story/${story.id}`)}
    >
      {showPriority || showStatus ? (
        <View
          pointerEvents="none"
          style={{
            flexDirection: "row",
            alignItems: "center",
            gap: 8,
            minHeight: 22,
          }}
        >
          {showStatus ? (
            <StatusIcon
              category={status?.category}
              size={20}
              color={
                status?.color || themeColors[dark ? "dark" : "light"].textMuted
              }
            />
          ) : null}
          {showPriority ? (
            <PriorityIcon priority={story.priority} size={18} />
          ) : null}
        </View>
      ) : null}
      <View style={{ flex: 1, minWidth: 0, gap: 3 }}>
        <Text
          numberOfLines={1}
          ellipsizeMode="tail"
          style={{ fontSize: 16, lineHeight: 22, fontWeight: "500" }}
        >
          {story.title}
        </Text>
        {metadata ? (
          <Text
            color="muted"
            numberOfLines={1}
            style={{ fontSize: 13, lineHeight: 18, fontWeight: "400" }}
          >
            {metadata}
          </Text>
        ) : null}
      </View>
      {showAssignee ? (
        <View
          style={{
            flexDirection: "row",
            alignItems: "center",
            gap: 10,
            minHeight: 24,
          }}
        >
          <Avatar
            name={assignee?.fullName || assignee?.username}
            src={assignee?.avatarUrl}
            style={{ width: 24, height: 24 }}
          />
        </View>
      ) : null}
    </SwipeableRow>
  );
});
