import { ArrowLeftIcon, StoryMissingIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { NewStoryButton } from "@/components/ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <StoryMissingIcon className="h-20 w-auto rotate-12" />
        <Text className="mb-6 mt-10" fontSize="3xl">
          404: Project Detour
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Oops! It seems the project path hit a snag. Our team&lsquo;s on it!
          While we clear the roadblock, why not explore other routes to
          productivity?
        </Text>
        <Flex gap={2}>
          <Button
            className="gap-1 pl-2"
            color="tertiary"
            href="/"
            leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
          >
            Back to home
          </Button>
          <NewStoryButton
            className="dark:bg-opacity-20 dark:hover:bg-opacity-40"
            size="md"
          >
            Create story
          </NewStoryButton>
        </Flex>
      </Box>
    </Box>
  );
}
