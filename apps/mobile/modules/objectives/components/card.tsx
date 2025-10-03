import React from "react";
import { Objective } from "../types";
import { Badge, Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { format } from "date-fns";
import { Pressable } from "react-native";

type ObjectiveHealth = "On Track" | "At Risk" | "Off Track" | null;

const healthColors = {
  "On Track": "success",
  "At Risk": "warning",
  "Off Track": "destructive",
  null: "tertiary",
} as const;

export const Card = ({ objective }: { objective: Objective }) => {
  const startDateObj = new Date(objective.startDate);
  const endDateObj = new Date(objective.endDate);

  return (
    <Pressable
      style={({ pressed }) => [pressed && { backgroundColor: colors.gray[50] }]}
    >
      <Row
        align="center"
        justify="between"
        className="px-4 py-3 border-t-[0.5px] border-gray-100"
      >
        <Row align="center" gap={3}>
          <Row className="bg-gray-100 rounded-lg p-1.5">
            <SymbolView
              name="target"
              size={20}
              weight="bold"
              tintColor={colors.gray.DEFAULT}
            />
          </Row>
          <Col>
            <Text numberOfLines={1} fontWeight="semibold">
              {objective.name}
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
        <Badge color={healthColors[objective.health]} rounded="md">
          <Text
            fontSize="sm"
            className="capitalize"
            color={objective.health === "On Track" ? "white" : undefined}
          >
            {objective.health || "No Status"}
          </Text>
        </Badge>
      </Row>
    </Pressable>
  );
};
