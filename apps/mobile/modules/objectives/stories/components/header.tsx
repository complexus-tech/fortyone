import React, { useState } from "react";
import { Pressable } from "react-native";
import { Row, Text, Back, StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";
import { useObjective } from "@/modules/objectives/hooks/use-objectives";
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
  const { objectiveId } = useGlobalSearchParams<{
    objectiveId: string;
  }>();
  const { data: objective } = useObjective(objectiveId);
  const { getTermDisplay } = useTerminology();
  const [isOpened, setIsOpened] = useState(false);

  // truncate objective name to 16 characters and add ellipsis if needed
  const objectiveName =
    objective?.name && objective.name.length > 16
      ? objective.name.slice(0, 16) + "..."
      : objective?.name;

  return (
    <Row className="mb-3" asContainer gap={2} justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {objectiveName} /{" "}
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
