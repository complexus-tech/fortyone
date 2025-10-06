import React from "react";
import { Pressable } from "react-native";
import { useRouter } from "expo-router";
import { Dot, PriorityIcon } from "@/components/icons";
import { Row } from "./row";
import { Text } from "./text";
import { Avatar } from "./avatar";
import { Story as StoryType } from "@/modules/stories/types";
import { useTeamStatuses } from "@/modules/statuses";
import { useMembers } from "@/modules/members";
import { Badge } from "@/components/ui";
import { hexToRgba } from "@/lib/utils/colors";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { DisplayColumn } from "@/types/stories-view-options";

export const Story = ({
  story: { id, title, statusId, priority, assigneeId, teamId, sequenceId },
  visibleColumns = ["Status", "Assignee", "Priority"],
}: {
  story: StoryType;
  visibleColumns?: DisplayColumn[];
}) => {
  const router = useRouter();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: members = [] } = useMembers();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId);
  const status = statuses.find((status) => status.id === statusId);
  const assignee = members.find((member) => member.id === assigneeId);
  // display the first 10 characters of the status name and if it's longer than 10 characters, add an ellipsis
  const statusName =
    status?.name && status?.name.length > 8
      ? status?.name.slice(0, 8) + "..."
      : status?.name;

  const handlePress = () => {
    router.push(`/story/${id}`);
  };

  return (
    <Pressable
      className="active:bg-gray-50 py-3 px-4 dark:active:bg-dark-300"
      onPress={handlePress}
    >
      <Row justify="between" align="center" gap={3}>
        <Row align="center" className="flex-1 gap-1.5">
          {visibleColumns.includes("Priority") && (
            <PriorityIcon priority={priority} size={18} />
          )}
          {visibleColumns.includes("ID") && (
            <Text color="muted" numberOfLines={1} fontWeight="medium">
              {team?.code}-{sequenceId}
            </Text>
          )}
          <Text className="flex-1" numberOfLines={1} fontWeight="medium">
            {title}
          </Text>
        </Row>
        <Row align="center" className="gap-1.5">
          {visibleColumns.includes("Status") && (
            <Badge
              style={{
                backgroundColor: hexToRgba(status?.color || "#6B665C", 0.1),
                borderColor: hexToRgba(status?.color || "#6B665C", 0.2),
                borderWidth: 1,
              }}
              className="px-2"
            >
              <Row align="center" gap={1}>
                <Dot color={status?.color} size={10} />
                <Text>{statusName || "No Status"}</Text>
              </Row>
            </Badge>
          )}
          {visibleColumns.includes("Assignee") && (
            <Avatar
              size="sm"
              name={assignee?.fullName || assignee?.username}
              src={assignee?.avatarUrl}
            />
          )}
        </Row>
      </Row>
    </Pressable>
  );
};
