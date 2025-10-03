import React from "react";
import { Objective } from "../types";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";

export const Card = ({ objective }: { objective: Objective }) => {
  return (
    <Row
      align="center"
      justify="between"
      gap={3}
      className="px-4 py-3.5 border-t-[0.5px] border-gray-100"
    >
      <Text numberOfLines={1} fontWeight="semibold">
        {objective.name} Lorem ipsum dolor sit.
      </Text>
      <Row align="center" gap={2}>
        <Row className="bg-gray-100 rounded-lg p-1.5">
          <SymbolView
            name="target"
            size={20}
            weight="bold"
            tintColor={colors.gray.DEFAULT}
          />
        </Row>
        {/* <Text numberOfLines={1} fontWeight="semibold">
          {objective.name} Lorem ipsum dolor sit.
        </Text> */}
      </Row>
    </Row>
  );
};
