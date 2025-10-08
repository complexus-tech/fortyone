import React from "react";
import { Avatar, Row, Text } from "@/components/ui";
import { View } from "react-native";
import { formatDistanceToNow } from "date-fns";
import { useMembers } from "@/modules/members/hooks/use-members";
import { useTeamStatuses } from "@/modules/statuses/hooks/use-statuses";
import { useTeamObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTeamSprints } from "@/modules/sprints/hooks";
import type { StoryActivity } from "@/modules/stories/types";
import { useTerminology } from "@/hooks";
import { truncateText } from "@/lib/utils";

export const ActivityItem = ({
  userId,
  field,
  currentValue,
  type,
  createdAt,
  teamId,
  isTimeShown = false,
}: StoryActivity & { teamId: string; isTimeShown?: boolean }) => {
  const { getTermDisplay } = useTerminology();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useTeamStatuses(teamId);
  const { data: objectives = [] } = useTeamObjectives(teamId);
  const { data: sprints = [] } = useTeamSprints(teamId);

  const member = members.find((m) => m.id === userId);

  if (field === "completed_at") {
    return null;
  }

  const getFieldLabel = (field: string) => {
    const fieldMap: Record<string, string> = {
      title: "Title",
      description: "Description",
      status_id: "Status",
      priority: "Priority",
      assignee_id: "Assignee",
      start_date: "Start date",
      end_date: "Deadline",
      sprint_id: "Sprint",
      objective_id: "Objective",
    };
    return fieldMap[field] || field;
  };

  const getFieldValue = (field: string, value: string) => {
    if (!value || value.includes("nil")) {
      return `No ${getFieldLabel(field)}`;
    }
    switch (field) {
      case "status_id":
        return statuses.find((s) => s.id === value)?.name || value;
      case "assignee_id":
        return members.find((m) => m.id === value)?.username || value;
      case "objective_id":
        return objectives.find((o) => o.id === value)?.name || value;
      case "sprint_id":
        return sprints.find((s) => s.id === value)?.name || value;
      case "start_date":
      case "end_date":
        return new Date(value.split(" ")[0]).toLocaleDateString();
      default:
        return value;
    }
  };

  const formatActivityText = () => {
    if (type === "create") {
      return (
        <Row align="center">
          <Avatar
            name={member?.fullName || member?.username}
            src={member?.avatarUrl}
            size="xs"
            className="mr-2"
          />
          <Text fontSize="sm" className="opacity-90" fontWeight="medium">
            {member?.username || "Unknown"}
          </Text>
          <Text fontSize="sm" color="muted" numberOfLines={1}>
            {` created the ${getTermDisplay("storyTerm")}`}
          </Text>
        </Row>
      );
    }
    const fieldLabel = getFieldLabel(field);
    const fieldValue = getFieldValue(field, currentValue);
    return (
      <Row align="center">
        <Avatar
          name={member?.fullName || member?.username}
          src={member?.avatarUrl}
          size="xs"
          className="mr-2"
        />
        <Text fontSize="sm" className="opacity-90" fontWeight="medium">
          {member?.username || "Unknown"}
        </Text>
        <Text fontSize="sm" color="muted" numberOfLines={1}>
          {truncateText(` changed ${fieldLabel} to ${fieldValue}`, 36)}
        </Text>
      </Row>
    );
  };

  return (
    <View className="my-2.5 pl-0.5">
      {formatActivityText()}
      {isTimeShown && (
        <Text
          numberOfLines={1}
          color="muted"
          fontSize="sm"
          className="mt-1 pl-7"
        >
          {formatDistanceToNow(new Date(createdAt), { addSuffix: true })}
        </Text>
      )}
    </View>
  );
};
