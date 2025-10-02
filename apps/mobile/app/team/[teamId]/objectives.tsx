import React from "react";
import { Pressable } from "react-native";
import { SafeContainer, Text, Row, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

const ObjectivesHeader = () => {
  return (
    <Row className="pb-2" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        Product/
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          Objectives
        </Text>
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

export default function ObjectivesScreen() {
  return (
    <SafeContainer isFull>
      <ObjectivesHeader />

      <Text color="muted" className="mt-4">
        This is the Objectives tab for the team page.
      </Text>
    </SafeContainer>
  );
}
