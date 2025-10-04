import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useTerminology } from "@/hooks/use-terminology";

export const Header = () => {
  const { objectiveId } = useGlobalSearchParams<{
    objectiveId: string;
  }>();
  const { data: objectives = [] } = useObjectives();
  const objective = objectives.find((obj) => obj.id === objectiveId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {objective?.name} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("storyTerm", { variant: "plural", capitalize: true })}
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
