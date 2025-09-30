import React, { useState } from "react";
import { Pressable } from "react-native";
import { Text } from "./text";
import { Row } from "./row";
import { colors } from "@/constants";

type Tab = {
  value: string;
  label: string;
};

type TabsProps = {
  tabs: Tab[];
  defaultValue?: string;
  onValueChange?: (value: string) => void;
};

export const Tabs = ({ tabs, defaultValue, onValueChange }: TabsProps) => {
  const [activeTab, setActiveTab] = useState(defaultValue || tabs[0]?.value);

  const handleTabPress = (value: string) => {
    setActiveTab(value);
    onValueChange?.(value);
  };

  return (
    <Row gap={2} asContainer className="mb-1">
      {tabs.map((tab) => {
        const isActive = activeTab === tab.value;
        return (
          <Pressable
            key={tab.value}
            onPress={() => handleTabPress(tab.value)}
            style={({ pressed }) => ({
              flex: 1,
              alignItems: "center",
              paddingVertical: 5,
              paddingHorizontal: 8,
              borderRadius: 24,
              borderWidth: 1,
              borderColor: colors.gray[100],
              backgroundColor: isActive
                ? colors.gray[50]
                : pressed
                  ? colors.gray[50]
                  : "transparent",
            })}
          >
            <Text
              fontWeight={isActive ? "semibold" : "medium"}
              color={isActive ? "black" : "muted"}
            >
              {tab.label}
            </Text>
          </Pressable>
        );
      })}
    </Row>
  );
};
