import React from "react";
import { Pressable } from "react-native";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

export const Header = () => {
  return (
    <Row justify="between" align="center" asContainer>
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
      </Text>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [pressed && { backgroundColor: "#F2F2F7" }]}
      >
        <SymbolView name="ellipsis" size={24} tintColor={colors.dark[50]} />
      </Pressable>
    </Row>
  );
};
