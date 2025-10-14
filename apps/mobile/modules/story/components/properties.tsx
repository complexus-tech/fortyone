import React from "react";
import { Row, Badge, Text, Col } from "@/components/ui";
import { Story, StoryPriority } from "@/modules/stories/types";
import { format } from "date-fns";
import { SymbolView } from "expo-symbols";
import { useColorScheme } from "nativewind";
import { colors } from "@/constants";
import { useFeatures, useSprintsEnabled } from "@/hooks";
import { PriorityBadge } from "./properties/priority";
import { StatusBadge } from "./properties/status";
import { AssigneeBadge } from "./properties/assignee";
import { ObjectiveBadge } from "./properties/objective";
import { SprintBadge } from "./properties/sprint";
import { LabelsBadge } from "./properties/labels";

export const Properties = ({ story }: { story: Story }) => {
  const { colorScheme } = useColorScheme();
  const sprintsEnabled = useSprintsEnabled(story.teamId);
  const { objectiveEnabled } = useFeatures();
  const iconColor =
    colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300];

  const onPriorityChange = (priority: StoryPriority) => {
    console.log("priority", priority);
  };

  const onStatusChange = (statusId: string) => {
    console.log("status", statusId);
  };

  const onAssigneeChange = (assigneeId: string | null) => {
    console.log("assignee", assigneeId);
  };

  const onObjectiveChange = (objectiveId: string | null) => {
    console.log("objective", objectiveId);
  };

  const onSprintChange = (sprintId: string | null) => {
    console.log("sprint", sprintId);
  };

  const onLabelsChange = (labelIds: string[]) => {
    console.log("labels", labelIds);
  };

  return (
    <Row asContainer>
      <Col className="my-4 ">
        <Text className="mb-2.5">Properties</Text>
        <Row wrap gap={2}>
          <StatusBadge story={story} onStatusChange={onStatusChange} />
          <PriorityBadge
            priority={story.priority || "No Priority"}
            onPriorityChange={onPriorityChange}
          />
          <AssigneeBadge story={story} onAssigneeChange={onAssigneeChange} />
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
          {objectiveEnabled && (
            <ObjectiveBadge
              story={story}
              onObjectiveChange={onObjectiveChange}
            />
          )}
          {sprintsEnabled && (
            <SprintBadge story={story} onSprintChange={onSprintChange} />
          )}
          <LabelsBadge story={story} onLabelsChange={onLabelsChange} />
        </Row>
      </Col>
    </Row>
  );
};
