import React, { useState } from "react";
import { Pressable } from "react-native";
import { StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useColorScheme } from "nativewind";
import type { StoriesViewOptions } from "@/types/stories-view-options";

type StoryOptionsButtonProps = {
  viewOptions: StoriesViewOptions;
  setViewOptions: (options: Partial<StoriesViewOptions>) => void;
  resetViewOptions: () => void;
};

export const StoryOptionsButton = ({
  viewOptions,
  setViewOptions,
  resetViewOptions,
}: StoryOptionsButtonProps) => {
  const { colorScheme } = useColorScheme();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <>
      <Pressable
        className="p-2 rounded-xl active:bg-gray-50 dark:active:bg-dark-300"
        onPress={() => setIsOpened(true)}
      >
        <SymbolView
          name="slider.horizontal.3"
          weight="medium"
          tintColor={
            colorScheme === "light" ? colors.dark[50] : colors.gray[200]
          }
        />
      </Pressable>
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
