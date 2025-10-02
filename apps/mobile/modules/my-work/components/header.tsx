import React from "react";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { Pressable } from "react-native";
import { colors } from "@/constants";

export const Header = () => {
  return (
    <Row className="pb-2" asContainer justify="between" align="center">
      <Text fontSize="2xl" fontWeight="semibold">
        My Work
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [
          pressed && { backgroundColor: colors.gray[50] },
        ]}
      >
        <SymbolView name="ellipsis" tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};
