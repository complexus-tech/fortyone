"use client";

import { Box, Flex, Text, Input } from "ui";

export const ProfileSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Profile Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <Box className="border-b border-gray-100 p-6 dark:border-dark-100">
          <Text as="h3" className="font-medium">
            Personal Information
          </Text>
          <Text className="mt-1" color="muted">
            Update your personal information and email address.
          </Text>
        </Box>

        <Box className="p-6">
          <Flex direction="column" gap={6}>
            <Box>
              <Input
                helpText="Your full name will be displayed on your profile"
                label="Full name"
                name="name"
                placeholder="Enter your full name"
              />
            </Box>

            <Box>
              <Input
                helpText="Your email address is used for notifications and sign-in"
                label="Email address"
                name="email"
                placeholder="Enter your email address"
                type="email"
              />
            </Box>
          </Flex>
        </Box>
      </Box>
    </Box>
  );
};
