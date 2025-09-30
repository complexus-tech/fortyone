import React from "react";
import { View, Pressable } from "react-native";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

interface TeamProps {
  name: string;
  color: string;
  onPress?: () => void;
}

export const Team = ({ name, color, onPress }: TeamProps) => {
  return (
    <Pressable
      style={({ pressed }) => [pressed && { backgroundColor: colors.gray[50] }]}
      onPress={onPress}
    >
      <Row
        align="center"
        justify="between"
        className="py-3.5 pl-0.5 min-h-[44px]"
      >
        <Row align="center">
          <View
            className="size-3 rounded-full mr-2"
            style={{ backgroundColor: color }}
          />
          <Text>{name}</Text>
        </Row>
        <SymbolView
          name="chevron.forward"
          size={12}
          tintColor={colors.gray.DEFAULT}
        />
      </Row>
    </Pressable>
  );
};
