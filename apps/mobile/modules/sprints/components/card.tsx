import React from "react";
import { Sprint } from "../types";
import { Badge, Row, Text } from "@/components/ui";

export const Card = ({ sprint }: { sprint: Sprint }) => {
  return (
    <Row
      align="center"
      justify="between"
      className="p-4 border-t border-gray-100"
    >
      <Row>
        <Text>{sprint.name}</Text>
      </Row>
      <Badge color="tertiary" rounded="md">
        <Text fontSize="sm">Upcoming</Text>
      </Badge>
    </Row>
  );
};
