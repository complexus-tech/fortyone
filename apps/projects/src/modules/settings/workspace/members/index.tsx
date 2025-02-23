"use client";

import { Box, Flex, Text, Button } from "ui";
import { TeamIcon } from "icons";
import { useMembers } from "@/lib/hooks/members";
import { SectionHeader } from "../../components";
import { WorkspaceMember } from "./components/member";

export const WorkspaceMembersSettings = () => {
  const { data: members = [] } = useMembers();
  return (
    <Box>
      <Text as="h1" className="mb-6 text-2xl font-medium">
        Members
      </Text>

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button color="tertiary" leftIcon={<TeamIcon />} variant="outline">
              Invite Members
            </Button>
          }
          description="Manage members of your workspace and their roles."
          title="Workspace members"
        />

        <Flex
          className="divide-y divide-gray-100 dark:divide-dark-100"
          direction="column"
        >
          {members.map((member) => (
            <WorkspaceMember key={member.id} {...member} />
          ))}
        </Flex>
      </Box>
    </Box>
  );
};
