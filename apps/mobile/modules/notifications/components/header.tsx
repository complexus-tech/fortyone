import React from "react";
import { Pressable } from "react-native";
import { Row, Text } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

export const Header = () => {
  return (
    <Row justify="between" align="center" asContainer className="pb-3">
      <Text fontSize="2xl" fontWeight="semibold" color="black">
        Notifications
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
