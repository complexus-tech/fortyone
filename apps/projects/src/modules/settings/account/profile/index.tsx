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
          <Button variant="solid">Save changes</Button>
        </Box>
      </Box>
    </Box>
  );
};
