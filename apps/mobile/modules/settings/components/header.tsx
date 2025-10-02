import React from "react";
import { Pressable } from "react-native";
import { Text, Row, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

export const Header = () => {
  return (
    <Row className="pb-2" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        Settings
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
