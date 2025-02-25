"use client";

import { Button, Box, Flex, Text, Wrapper, Avatar } from "ui";
import type { FormEvent } from "react";

export const CreateAccountForm = () => {
  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    // if new user, redirect to /onboarding/account
    // if existing user, redirect to workspace
  };

  return (
    <form className="space-y-5" onSubmit={handleSubmit}>
      <Wrapper className="py-3">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" gap={2}>
            <Avatar name="John Doe" rounded="md" />
            <Box>
              <Text>John Doe</Text>
              <Text color="muted" fontSize="sm">
                3 members
              </Text>
            </Box>
          </Flex>

          <Button color="tertiary" size="sm">
            Join
          </Button>
        </Flex>
      </Wrapper>
      <Flex align="center" className="my-3 gap-4" justify="between">
        <Box className="h-px w-full bg-white/10" />
        <Text className="text-[0.95rem] opacity-40" color="white">
          OR
        </Text>
        <Box className="h-px w-full bg-white/10" />
      </Flex>
      <Button
        align="center"
        className="mt-4"
        color="tertiary"
        fullWidth
        href="/onboarding/create"
      >
        Create your own workspace
      </Button>
    </form>
  );
};
