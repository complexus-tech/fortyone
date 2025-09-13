import { ArrowLeftIcon, StoryMissingIcon } from "icons";
import { Box, Button, Text } from "ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <StoryMissingIcon className="h-20 w-auto rotate-12" />
        <Text className="mb-6 mt-10" fontSize="3xl">
          404: Team not found
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          This team might not exist or you do not belong to this team.
        </Text>
        <Button
          className="gap-1 pl-2"
          color="tertiary"
          href="/my-work"
          leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
        >
          Go to my work
        </Button>
      </Box>
    </Box>
  );
}
