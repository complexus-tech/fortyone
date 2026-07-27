import { ArrowLeft2Icon, StoryMissingIcon } from "icons";
import { Box, Button, Text } from "ui";
import { withWorkspacePath } from "@/utils";

export const StoryRouteNotFound = ({
  workspaceSlug,
}: {
  workspaceSlug: string;
}) => (
  <Box className="flex h-screen items-center justify-center">
    <Box className="flex flex-col items-center">
      <StoryMissingIcon className="h-20 w-auto rotate-12" />
      <Text className="mt-10 mb-6" fontSize="3xl">
        404: Item not found
      </Text>
      <Text className="mb-6 max-w-md text-center" color="muted">
        This item might not exist or you do not have access to it.
      </Text>
      <Button
        className="gap-1 pl-2"
        color="tertiary"
        href={withWorkspacePath("/my-work", workspaceSlug)}
        leftIcon={<ArrowLeft2Icon className="h-[1.05rem] w-auto" />}
      >
        Go to my work
      </Button>
    </Box>
  </Box>
);
