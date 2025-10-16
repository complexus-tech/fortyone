import React, { useState } from "react";
import { Pressable } from "react-native";
import { StoriesOptionsSheet } from "@/components/ui";
import { SymbolView } from "expo-symbols";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
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
  const { resolvedTheme } = useTheme();
  const [isOpened, setIsOpened] = useState(false);

  return (
    <>
      <Pressable
        className="p-2 rounded-xl active:bg-gray-50 dark:active:bg-dark-300"
        onPress={() => setIsOpened(true)}
      >
        <SymbolView
          name="line.3.horizontal.decrease"
          size={25}
          weight="semibold"
          tintColor={
            resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]
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
