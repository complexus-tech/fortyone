import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
import { Box, Button, Text } from "ui";

export default function NotFound() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <StoryMissingIcon className="h-20 w-auto rotate-12" />
        <Text className="mb-6 mt-10" fontSize="3xl">
          404: Objective Detour
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Oops! It seems the objective path hit a snag. Our team&lsquo;s on it!
          While we clear the roadblock, why not explore other routes to
          productivity?
        </Text>
        <Button
          className="gap-1 pl-2"
          color="tertiary"
          href="/my-work"
          leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
        >
          Go to my work
        </Button>
      </Box>
    </Box>
  );
}
