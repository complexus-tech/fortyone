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
            <Text align="center" as="h1" className="mb-6" fontSize="4xl">
              Welcome to Complexus!
            </Text>
            <Text align="center" className="max-w-l mx-auto mb-6" color="muted">
              Your workspace is ready. Create your first story by pressing{" "}
              <Badge
                className="inline-flex rounded-[0.4rem] font-semibold"
                color="tertiary"
              >
                shift+n
              </Badge>{" "}
              or explore the features below.
            </Text>
            <Box className="grid gap-4">
              <ActionCard
                description="Collaborate better by inviting your teammates to join your workspace."
                icon={<UserIcon />}
                title="Build your team"
              />
              <ActionCard
                description="Connect your GitHub repositories and Slack channels for seamless integration."
                icon={<ChatIcon />}
                title="Set up integrations"
              />
              <ActionCard
                description="Boost your productivity with keyboard shortcuts (press ? for help)"
                icon={<PreferencesIcon />}
                title="Master shortcuts"
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
              Get Started
            </Button>
          </Box>
        </Flex>
      </Box>
    </Box>
  );
};
