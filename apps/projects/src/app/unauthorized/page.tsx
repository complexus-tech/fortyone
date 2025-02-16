import { StoryMissingIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";

export default function Unauthorized() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <StoryMissingIcon className="h-20 w-auto rotate-12" />
        <Text className="mb-6 mt-10" fontSize="3xl">
          Unauthorized Access
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          You don&apos;t have access to this workspace. This could be because it
          doesn&apos;t exist or you haven&apos;t been granted permission to view
          it. Please check the URL or contact your workspace administrator.
        </Text>
        <Flex gap={2}>
          <Button className="gap-1 pl-2" color="tertiary">
            Back to Home
          </Button>
          <Button color="tertiary" href="https://complexus.app/contact">
            Contact Support
          </Button>
        </Flex>
      </Box>
    </Box>
  );
}
