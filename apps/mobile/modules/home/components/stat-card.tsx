import React from "react";
import { Wrapper, Col, Text } from "@/components/ui";
import { Ionicons } from "@expo/vector-icons";
import { colors } from "@/constants";

type StatCardProps = {
  count: number;
  label: string;
  icon?: keyof typeof Ionicons.glyphMap;
};

export const StatCard = ({ count, label, icon }: StatCardProps) => {
  return (
    <Wrapper className="p-4">
      <Col gap={2}>
        <Text fontSize="2xl" fontWeight="semibold">
          {count}
        </Text>
        <Text color="muted" fontSize="sm" className="opacity-80">
          {label}
        </Text>
      </Col>
      {icon && (
        <Ionicons
          name={icon}
          size={20}
          color={colors.gray.DEFAULT}
          style={{ position: "absolute", top: 16, right: 16, opacity: 0.5 }}
        />
      )}
    </Wrapper>
  );
};
