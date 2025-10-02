import React from "react";
import { Wrapper, Col, Text } from "@/components/ui";
import { SymbolView, SFSymbol } from "expo-symbols";

type StatCardProps = {
  count?: number;
  label: string;
  icon: SFSymbol;
  iconColor?: string;
};

export const StatCard = ({ count, label, icon, iconColor }: StatCardProps) => {
  return (
    <Wrapper>
      <Col gap={2}>
        <Text fontSize="2xl" fontWeight="semibold">
          {count || 0}
        </Text>
        <Text color="muted" fontSize="md" className="opacity-80">
          {label}
        </Text>
      </Col>
      <SymbolView
        name={icon}
        size={20}
        tintColor={iconColor}
        style={{ position: "absolute", top: 14, right: 16 }}
      />
    </Wrapper>
  );
};
