import React from "react";
import { Sprint } from "../types";
import { Badge, Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { format } from "date-fns";
import { Pressable } from "react-native";

type SprintStatus = "completed" | "in progress" | "upcoming";

const statusColors = {
  completed: "tertiary",
  "in progress": "success",
  upcoming: "tertiary",
} as const;

export const Card = ({ sprint }: { sprint: Sprint }) => {
  const startDateObj = new Date(sprint.startDate);
  const endDateObj = new Date(sprint.endDate);
  const now = new Date();

  // Calculate sprint status
  let sprintStatus: SprintStatus = "completed";
  if (startDateObj <= now && endDateObj >= now) {
    sprintStatus = "in progress";
  } else if (startDateObj > now) {
    sprintStatus = "upcoming";
  }

  return (
    <Pressable
      style={({ pressed }) => [pressed && { backgroundColor: colors.gray[50] }]}
    >
      <Row
        align="center"
        justify="between"
        className="px-4 py-3 border-t-[0.5px] border-gray-100"
        gap={3}
      >
        <Row align="center" gap={3} className="w-8/12">
          <Row className="bg-gray-100 rounded-lg p-1.5">
            <SymbolView
              name="play.circle"
              size={20}
              weight="bold"
              tintColor={colors.gray.DEFAULT}
            />
          </Row>
          <Col>
            <Text numberOfLines={1} fontWeight="semibold">
              {sprint.name}
            </Text>
            <Row align="center" gap={1}>
              <SymbolView
                name="calendar"
                size={19}
                tintColor={colors.gray.DEFAULT}
              />
              <Text fontSize="sm" numberOfLines={1}>
                {format(startDateObj, "MMM d")} → {format(endDateObj, "MMM d")}
              </Text>
            </Row>
          </Col>
        </Row>
        <Badge color={statusColors[sprintStatus]} rounded="md">
          <Text
            fontSize="sm"
            className="capitalize"
            color={sprintStatus === "in progress" ? "white" : undefined}
          >
            {sprintStatus}
          </Text>
        </Badge>
      </Row>
    </Pressable>
  );
};
