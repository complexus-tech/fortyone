import React, { useState } from "react";
import { Pressable } from "react-native";
import { Row, Text, Back, StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";
import { useTeams } from "@/modules/teams/hooks/use-teams";
import { useTerminology } from "@/hooks/use-terminology";
import type { StoriesViewOptions } from "@/types/stories-view-options";

type HeaderProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

export const Header = ({
  viewOptions,
  setViewOptions,
  resetViewOptions,
}: HeaderProps) => {
  const { teamId } = useGlobalSearchParams<{ teamId: string }>();
  const { data: teams = [] } = useTeams();
  const team = teams.find((team) => team.id === teamId)!;
  const { getTermDisplay } = useTerminology();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <Row className="mb-3" asContainer align="center" justify="between">
      <Row align="center" gap={3}>
        <Back />
        <Text fontSize="2xl" fontWeight="semibold">
          {team?.name} /{" "}
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
      </Row>
      <Pressable
        className="p-2 rounded-md"
        style={({ pressed }) => [
          pressed && { backgroundColor: colors.gray[50] },
        ]}
        onPress={() => setIsOpened(true)}
      >
        <SymbolView
          name="line.3.horizontal.decrease"
          tintColor={colors.dark[50]}
        />
      </Pressable>
      <StoriesOptionsSheet
        isOpened={isOpened}
        setIsOpened={setIsOpened}
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
    </Row>
  );
};
