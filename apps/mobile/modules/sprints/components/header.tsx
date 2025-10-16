import React from "react";
import { Text, Row, Back, ContextMenuButton } from "@/components/ui";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";
import { truncateText } from "@/lib/utils";

export const Header = () => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId)!;
  const { getTermDisplay } = useTerminology();

  return (
    <Row asContainer align="center" gap={3} justify="between">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {truncateText(team?.name ?? "", 12)} /{" "}
        <Text
          fontSize="2xl"
          color="muted"
          fontWeight="semibold"
          className="opacity-80"
        >
          {getTermDisplay("sprintTerm", {
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
