import { Button, Flex, Text } from "ui";
import { DeleteIcon, ObjectiveIcon, MilestonesIcon } from "icons";

export const StoriesToolbar = () => {
  return (
    <Flex
      align="center"
      className="sticky bottom-8 left-1/2 right-1/2 z-50 w-max -translate-x-1/2 rounded-lg border border-gray-100 bg-white/60 px-2.5 py-2 shadow-lg shadow-dark/10 backdrop-blur dark:border-dark-100/50 dark:bg-dark-200/70 dark:shadow-dark/20"
      gap={2}
    >
      <Text className="ml-2 mr-4" color="muted">
        2 selected
      </Text>
      <Button
        color="tertiary"
        leftIcon={
          <MilestonesIcon className="h-[1.15rem] w-auto dark:text-gray" />
        }
        variant="outline"
      >
        Add to sprint
      </Button>
      <Button
        color="tertiary"
        leftIcon={
          <ObjectiveIcon className="h-[1.15rem] w-auto dark:text-gray" />
        }
        variant="outline"
      >
        Add to objective
      </Button>
      <Button
        className="border-opacity-30"
        color="danger"
        leftIcon={<DeleteIcon className="h-[1.15rem] w-auto" />}
        variant="outline"
      >
        Delete
      </Button>
    </Flex>
  );
};
