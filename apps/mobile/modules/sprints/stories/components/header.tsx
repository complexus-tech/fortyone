import React from "react";
import { Pressable } from "react-native";
import { Row, Text, Back } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";
import { useSprints } from "@/modules/sprints/hooks";
import { useTerminology } from "@/hooks/use-terminology";

export const Header = () => {
  const { sprintId } = useGlobalSearchParams<{ sprintId: string }>();
  const { data: sprints = [] } = useSprints();
  const sprint = sprints.find((sprint) => sprint.id === sprintId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {sprint?.name} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("storyTerm", {
            variant: "plural",
            capitalize: true,
          })}
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
