import React, { useState } from "react";
import { Row, Text, StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { Pressable } from "react-native";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";
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
  const { colorScheme } = useColorScheme();
  const [isOpened, setIsOpened] = useState(false);
  return (
    <>
      <Row className="mb-2" asContainer justify="between" align="center">
        <Text fontSize="2xl" fontWeight="semibold">
          My Work
        </Text>
        <Pressable
          className="p-2 rounded-xl active:bg-gray-50 dark:active:bg-dark-300"
          onPress={() => setIsOpened(true)}
        >
          <SymbolView
            name="line.3.horizontal.decrease"
            tintColor={
              colorScheme === "light" ? colors.dark[50] : colors.gray[300]
            }
          />
        </Pressable>
      </Row>
      <StoriesOptionsSheet
        isOpened={isOpened}
        setIsOpened={setIsOpened}
        viewOptions={viewOptions}
        setViewOptions={setViewOptions}
        resetViewOptions={resetViewOptions}
      />
    </>
  );
};
