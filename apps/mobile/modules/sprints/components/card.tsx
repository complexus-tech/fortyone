import React from "react";
import { Sprint } from "../types";
import { Badge, Col, Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";

export const Card = ({ sprint }: { sprint: Sprint }) => {
  return (
    <Row
      align="center"
      justify="between"
      className="px-4 py-3 border-t-[0.5px] border-gray-100"
    >
      <Row align="center" gap={3}>
        <SymbolView
          name="play.circle"
          size={20}
          weight="bold"
          tintColor={colors.gray.DEFAULT}
        />
        <Col>
          <Text numberOfLines={1}>{sprint.name}</Text>
          <Row align="center" gap={1}>
            <SymbolView
              name="calendar"
              size={20}
              tintColor={colors.gray.DEFAULT}
            />
            <Text fontSize="sm" numberOfLines={1}>
              {sprint.name}
            </Text>
          </Row>
        </Col>
      </Row>
      <Badge color="tertiary" rounded="md">
        <Text fontSize="sm">Upcoming</Text>
      </Badge>
    </Row>
  );
};
