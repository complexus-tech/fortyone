import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useMembers } from "@/modules/members/hooks/use-members";
import React from "react";
import { Row, Badge, Text, Avatar } from "@/components/ui";
import { hexToRgba } from "@/lib/utils/colors";
import { Dot, PriorityIcon } from "@/components/icons";
import { Story } from "@/modules/stories/types";
import { format } from "date-fns";
import { SymbolView } from "expo-symbols";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";

export const Properties = ({ story }: { story: Story }) => {
  const { colorScheme } = useColorScheme();
  const { data: statuses = [] } = useTeamStatuses(story.teamId);
  const { data: members = [] } = useMembers();
  const status = statuses.find((s) => s.id === story.statusId);
  const assignee = members.find((m) => m.id === story.assigneeId);
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <Row wrap gap={2} asContainer className="my-4">
      <Badge
        style={{
          backgroundColor: hexToRgba(status?.color, 0.1),
          borderColor: hexToRgba(status?.color, 0.2),
        }}
      >
        <Dot color={status?.color} size={12} />
        <Text>{status?.name || "No Status"}</Text>
      </Badge>
      <Badge color="tertiary">
        <PriorityIcon priority={story.priority || "No Priority"} />
        <Text>{story.priority || "No Priority"}</Text>
      </Badge>
      <Badge color="tertiary" className="pl-1.5">
        <Avatar
          size="xs"
          name={assignee?.fullName || assignee?.username}
          src={assignee?.avatarUrl}
        />
        <Text>{assignee?.username || "No Assignee"}</Text>
      </Badge>
      <Badge color="tertiary">
        <SymbolView name="calendar" size={16} tintColor={iconColor} />
        <Text color={story.startDate ? undefined : "muted"}>
          {story.startDate
            ? format(new Date(story.startDate), "MMM d")
            : "Add start date"}
        </Text>
      </Badge>
    </Row>
  );
};
