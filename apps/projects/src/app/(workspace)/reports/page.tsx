import { AnalyticsIcon, ArrowLeftIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { NewStoryButton } from "@/components/ui";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Reports",
};

export default function Page() {
  return (
    <Box className="flex h-screen items-center justify-center">
      <Box className="flex flex-col items-center">
        <AnalyticsIcon className="h-20 w-auto rotate-12" strokeWidth={1.3} />
        <Text className="mb-6 mt-10" fontSize="3xl">
          Coming soon...
        </Text>
        <Text className="mb-6 max-w-md text-center" color="muted">
          Oops! This page is under construction. Our team is working on it!
          While we clear the roadblock, why not explore other routes to
          productivity?
        </Text>
        <Flex gap={2}>
          <Button
            className="gap-1 pl-2"
            color="tertiary"
            href="/my-work"
            leftIcon={<ArrowLeftIcon className="h-[1.05rem] w-auto" />}
          >
            Goto my work
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
