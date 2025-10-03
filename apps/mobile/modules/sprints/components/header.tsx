import React from "react";
import { Pressable } from "react-native";
import { Text, Row, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";

export const Header = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const { name } = teams.find((team) => team.id === teamId)!;
  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {name} /{" "}
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
