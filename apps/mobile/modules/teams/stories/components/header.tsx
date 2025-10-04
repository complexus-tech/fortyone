import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";

type HeaderProps = {
  teamName?: string;
};

export const Header = ({ teamName = "Product" }: HeaderProps) => {
  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {teamName} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          Stories
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
