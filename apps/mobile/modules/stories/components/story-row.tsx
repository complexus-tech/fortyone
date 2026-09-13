import type { Member } from "@/types";
import type { Status } from "@/types/statuses";
import type { DisplayColumn } from "@/types/stories-view-options";
import type { Team } from "@/modules/teams/types";
import type { Story } from "../types";
import { memo } from "react";
import { Pressable } from "react-native";
import { useRouter } from "expo-router";
import { Avatar, Badge, Row, Text } from "@/components/ui";
import { Dot, PriorityIcon } from "@/components/icons";
import { hexToRgba } from "@/lib/utils/colors";

type StoryRowProps = {
  story: Story;
  visibleColumns: DisplayColumn[];
  status?: Status;
  assignee?: Member;
  team?: Team;
};

export const StoryRow = memo(function StoryRow({
  story,
  visibleColumns,
  status,
  assignee,
  team,
}: StoryRowProps) {
  const router = useRouter();
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={[story.title, status?.name, assignee?.fullName]
        .filter(Boolean)
        .join(", ")}
      accessibilityHint="Open story"
      className="active:bg-gray-50 p-4 dark:active:bg-dark-300 min-h-[50px]"
      onPress={() => router.push(`/story/${story.id}`)}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" className="flex-1 gap-1.5">
          {visibleColumns.includes("Priority") ? (
            <PriorityIcon priority={story.priority} size={18} />
          ) : null}
          {visibleColumns.includes("ID") ? (
            <Text color="muted" numberOfLines={1} fontWeight="medium">
              {team?.code}-{story.sequenceId}
            </Text>
          ) : null}
          <Text className="flex-1" numberOfLines={2} fontWeight="medium">
            {story.title}
          </Text>
        </Row>
        <Row align="center" className="gap-1.5">
          {visibleColumns.includes("Status") ? (
            <Badge
              style={{
                backgroundColor: hexToRgba(status?.color || "#6B665C", 0.1),
                borderColor: hexToRgba(status?.color || "#6B665C", 0.2),
                borderWidth: 1,
              }}
              rounded="xl"
              className="pr-2 max-w-28"
            >
              <Row align="center" gap={1}>
                <Dot color={status?.color} size={10} />
                <Text numberOfLines={1} className="text-[15px] shrink">
                  {status?.name || "No status"}
                </Text>
              </Row>
            </Badge>
          ) : null}
          {visibleColumns.includes("Assignee") ? (
            <Avatar
              size="sm"
              name={assignee?.fullName || assignee?.username}
              src={assignee?.avatarUrl}
            />
          ) : null}
        </Row>
      </Row>
    </Pressable>
  );
});
