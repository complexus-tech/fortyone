"use client";
import { Button, Flex, Text } from "ui";
import { DeleteIcon, ObjectiveIcon, SprintsIcon } from "icons";
import { useBoard } from "./stories-board";

export const StoriesToolbar = () => {
  const { selectedStories } = useBoard();
  return (
    <Flex
      align="center"
      className="fixed bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-[0.55rem] border-[0.5px] border-gray-100 bg-white/60 px-2.5 py-2 shadow-lg shadow-dark/10 backdrop-blur dark:border-dark-50 dark:bg-dark-300/70 dark:shadow-dark/20"
      gap={2}
    >
      <Text className="ml-2 mr-4 opacity-80">
        {selectedStories.length} selected
      </Text>
      <Button
        color="tertiary"
        leftIcon={<SprintsIcon className="h-[1.15rem] w-auto" />}
        variant="outline"
      >
        Add to sprint
      </Button>
      <Button
        color="tertiary"
        leftIcon={<ObjectiveIcon className="h-[1.15rem] w-auto" />}
        variant="outline"
      >
        Add to objective
      </Button>
      <Button
        color="danger"
        leftIcon={<DeleteIcon className="h-[1.15rem] w-auto" />}
      >
        Delete
      </Button>
    </Flex>
  );
};
