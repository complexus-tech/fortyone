"use client";

import { Box, Button, Divider, Input, Text } from "ui";
import { SectionHeader } from "../../components";
import { ProfilePicture } from "./components/profile-picture";

export const ProfileSettings = () => {
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-semibold">
        Profile Settings
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={<ProfilePicture />}
          description="Update your personal information and profile picture."
          title="Personal Information"
        />
        <Box className="p-6">
          <Box className="grid grid-cols-2 gap-6">
            <Input
              label="Full name"
              name="fullName"
              placeholder="Enter your full name"
            />

            <Input
              label="Email address"
              name="email"
              placeholder="Enter your email address"
              type="email"
            />

            <Input
              helpText="This is your unique identifier"
              label="Username"
              name="username"
              placeholder="Enter your username"
            />
          </Box>
          <Divider className="my-3" />
          <Button>Save changes</Button>
        </Box>
      </Box>

      <Box className="mt-6 rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          description="Leave your current workspace. This action cannot be undone."
          title="Leave Workspace"
        />
        <Box className="p-6">
          <Box>
            <Text color="muted">When you leave the workspace:</Text>
            <ul className="mt-2 list-disc pl-5 text-gray dark:text-gray-300">
              <li>You will lose access to all workspace data</li>
              <li>Your account will remain active</li>
              <li>You can be invited back by workspace admins</li>
            </ul>
          </Box>
          <Divider className="my-4" />
          <Button className="bg-primary dark:bg-primary" variant="naked">
            Leave Workspace
          </Button>
        </Box>
      </Box>
    </Box>
  );
};
