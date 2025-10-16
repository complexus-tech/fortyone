import React, { useState } from "react";
import { StoriesOptionsSheet } from "@/components/ui";
import { colors } from "@/constants";
import { useTheme } from "@/hooks";
import type { StoriesViewOptions } from "@/types/stories-view-options";
import { Host, HStack, Image } from "@expo/ui/swift-ui";
import { frame, glassEffect } from "@expo/ui/swift-ui/modifiers";

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
      <Host matchContents style={{ width: 45, height: 45 }}>
        <HStack
          modifiers={[
            frame({ width: 45, height: 45 }),
            glassEffect({
              glass: {
                interactive: true,
                variant: "regular",
              },
            }),
          ]}
          onPress={() => setIsOpened(true)}
        >
          <Image
            systemName="line.3.horizontal.decrease"
            size={20}
            color={
              resolvedTheme === "light" ? colors.dark[50] : colors.gray[200]
            }
          />
        </HStack>
      </Host>
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
