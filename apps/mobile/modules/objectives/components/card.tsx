import React from "react";
import { Objective } from "../types";
import { Badge, Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { format } from "date-fns";
import { Pressable } from "react-native";
import { useColorScheme } from "nativewind";
import { router } from "expo-router";

type ObjectiveHealth = "On Track" | "At Risk" | "Off Track" | null;

const getHealthColor = (health: ObjectiveHealth) => {
  switch (health) {
    case "At Risk":
      return "warning";
    case "Off Track":
      return "danger";
    default:
      return "tertiary";
  }
};

export const Card = ({ objective }: { objective: Objective }) => {
  const { colorScheme } = useColorScheme();
  const startDateObj = new Date(objective.startDate);
  const endDateObj = new Date(objective.endDate);
  const now = new Date();

  // Calculate objective status based on dates
  let objectiveStatus: ObjectiveHealth = null;
  if (startDateObj <= now && endDateObj >= now) {
    objectiveStatus = objective.health || "On Track";
  } else if (startDateObj > now) {
    objectiveStatus = "On Track";
  } else {
    objectiveStatus = objective.health || "Off Track";
  }

  const textColor = ["At Risk", "Off Track"].includes(objectiveStatus)
    ? "white"
    : undefined;

  const handlePress = () => {
    router.push(`/team/${objective.teamId}/objectives/${objective.id}`);
  };

  return (
    <Pressable
      className="active:bg-gray-50 dark:active:bg-dark-300"
      onPress={handlePress}
    >
      <Row
        align="center"
        justify="between"
        className="p-4 border-b border-gray-50 dark:border-dark"
      >
        <Row align="center" gap={3} className="w-8/12">
          <Row className="bg-gray-100 dark:bg-dark-200 rounded-lg p-1.5">
            <SymbolView
              name="target"
              size={20}
              weight="bold"
              tintColor={
                colorScheme === "light" ? colors.gray.DEFAULT : colors.gray[300]
              }
            />
          </Row>
          <Col gap={1}>
            <Text numberOfLines={1} fontWeight="semibold">
              {objective.name}
            </Text>
            <Text fontSize="sm" numberOfLines={1}>
              {format(startDateObj, "MMM d")} → {format(endDateObj, "MMM d")}
            </Text>
          </Col>
        </Row>
        <Badge color={getHealthColor(objectiveStatus)} rounded="md">
          <Text fontSize="sm" className="capitalize" color={textColor}>
            {objectiveStatus || "No Status"}
          </Text>
        </Badge>
      </Row>
    </Pressable>
  );
};
