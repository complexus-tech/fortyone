"use client";

import { Badge, Box, Button, Flex, Text } from "ui";
import { UserIcon, ChatIcon, PreferencesIcon } from "icons";
import { ActionCard } from "./components/action-card";

export const Welcome = () => {
  const handleNext = async () => {
    // TODO: Navigate to main app
  };

  return (
    <Box className="min-h-screen">
      <Box className="min-h-screen">
        <Flex align="center" className="min-h-screen" justify="center">
          <Box className="w-full max-w-xl px-4">
            <Text align="center" as="h1" className="mb-6" fontSize="3xl">
              You&apos;re good to go
            </Text>
            <Text align="center" className="max-w-l mx-auto mb-6" color="muted">
              Go ahead and explore the app. When you&apos;re ready, create your
              first story by pressing{" "}
              <Badge
                className="inline-flex rounded-[0.4rem] font-semibold"
                color="tertiary"
              >
                shift+c
              </Badge>
            </Text>
            <Box className="grid gap-4">
              <ActionCard
                description="Make sure to invite your team members."
                icon={<UserIcon className="h-5 w-5" />}
                title="Tell your team"
              />
              <ActionCard
                description="Link your pull requests and create issues from Slack."
                icon={<ChatIcon className="h-5 w-5" />}
                title="Integrate GitHub & Slack"
              />
              <ActionCard
                description="Learn the keyboard commands by pressing ?"
                icon={<PreferencesIcon className="h-5 w-5" />}
                title="Keyboard shortcuts"
              />
            </Box>
            <Button
              align="center"
              className="mt-4"
              fullWidth
              href="/my-work"
              onClick={handleNext}
              rounded="lg"
              size="lg"
            >
              Open Complexus
            </Button>
          </Box>
        </Flex>
      </Box>
    </Box>
  );
};
