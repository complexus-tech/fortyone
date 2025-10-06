import React, { useState } from "react";
import { Pressable } from "react-native";
import { Row, Text, Back, StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants/colors";
import { useGlobalSearchParams } from "expo-router";
import { useSprint } from "@/modules/sprints/hooks/use-sprints";
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
  const { sprintId } = useGlobalSearchParams<{ sprintId: string }>();
  const { data: sprint } = useSprint(sprintId);
  const { getTermDisplay } = useTerminology();
  const [isOpened, setIsOpened] = useState(false);
  // truncate sprint name to 16 characters and add ellipsis if needed
  const sprintName =
    sprint?.name && sprint.name.length > 16
      ? sprint.name.slice(0, 16) + "..."
      : sprint?.name;

  return (
    <Row className="mb-3" asContainer justify="between" align="center">
      <Back />
      <Text fontSize="2xl" fontWeight="semibold">
        {sprintName} /{" "}
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
