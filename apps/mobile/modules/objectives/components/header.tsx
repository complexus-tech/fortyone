import React from "react";
import { Row, Text, Back, ContextMenuButton } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";

export const Header = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row className="mb-3" asContainer align="center" gap={3} justify="between">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {team?.name} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("objectiveTerm", {
            variant: "plural",
            capitalize: true,
          })}
        </Text>
      </Text>
      <ContextMenuButton
        actions={[
          {
            systemImage: "link",
            label: "Copy link",
            onPress: () => {},
          },
        ]}
      />
    </Row>
  );
};
