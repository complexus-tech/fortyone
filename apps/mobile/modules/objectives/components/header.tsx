import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";

export const Header = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();

  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        Product /{" "}
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
