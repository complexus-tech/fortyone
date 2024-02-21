import { Button, Flex, Text } from "ui";
import { DeleteIcon, ProjectsIcon, SprintsIcon } from "../icons";

export const IssuesToolbar = () => {
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
        leftIcon={<SprintsIcon className="h-[1.15rem] w-auto dark:text-gray" />}
        variant="outline"
      >
        Add to sprint
      </Button>
      <Button
        color="tertiary"
        leftIcon={
          <ProjectsIcon className="h-[1.15rem] w-auto dark:text-gray" />
        }
        variant="outline"
      >
        Add to project
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
