import React from "react";
import { Pressable } from "react-native";
import { SafeContainer, Text, Row, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

const SprintsHeader = () => {
  return (
    <Row className="pb-2" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        Product /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          Sprints
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

export default function SprintsScreen() {
  return (
    <SafeContainer isFull>
      <SprintsHeader />

      <Text color="muted" className="mt-4">
        This is the Sprints tab for the team page.
      </Text>
    </SafeContainer>
  );
}
