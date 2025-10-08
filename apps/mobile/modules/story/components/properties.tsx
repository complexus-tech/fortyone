import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useMembers } from "@/modules/members/hooks/use-members";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeamSprints } from "@/modules/sprints/hooks";
import React from "react";
import { Row, Badge, Text, Avatar, Col } from "@/components/ui";
import { hexToRgba } from "@/lib/utils/colors";
import { truncateText } from "@/lib/utils/text";
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
  const { data: objectives = [] } = useTeamObjectives(story.teamId);
  const { data: sprints = [] } = useTeamSprints(story.teamId);
  const status = statuses.find((s) => s.id === story.statusId);
  const assignee = members.find((m) => m.id === story.assigneeId);
  const objective = objectives.find((o) => o.id === story.objectiveId);
  const sprint = sprints.find((s) => s.id === story.sprintId);
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  return (
    <Row asContainer>
      <Col className="my-4 dark:bg-dark-300/80 border bg-gray-50/60 border-gray-100/70 dark:border-dark-200 p-4 rounded-3xl">
        <Text className="mb-2.5">Properties</Text>
        <Row wrap gap={2}>
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
          <Badge color="tertiary">
            <SymbolView name="calendar" size={16} tintColor={iconColor} />
            <Text color={story.endDate ? undefined : "muted"}>
              {story.endDate
                ? format(new Date(story.endDate), "MMM d")
                : "Add deadline"}
            </Text>
          </Badge>
          <Badge color="tertiary">
            <SymbolView name="target" size={16} tintColor={iconColor} />
            <Text color={story.objectiveId ? undefined : "muted"}>
              {truncateText(objective?.name || "Add objective", 12)}
            </Text>
          </Badge>
          <Badge color="tertiary">
            <SymbolView name="play.circle" size={16} tintColor={iconColor} />
            <Text color={story.sprintId ? undefined : "muted"}>
              {truncateText(sprint?.name || "Add sprint", 10)}
            </Text>
          </Badge>
          <Badge color="tertiary">
            <SymbolView name="tag.fill" size={16} tintColor={iconColor} />
            <Text color={story.labels?.length ? undefined : "muted"}>
              {story.labels?.length
                ? `${story.labels.length} labels`
                : "Add labels"}
            </Text>
          </Badge>
        </Row>
      </Col>
    </Row>
  );
};
