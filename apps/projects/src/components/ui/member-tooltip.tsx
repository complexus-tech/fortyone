import { Member } from "@/types";
import React from "react";
import { Avatar, Box, Button, Flex, Text, Tooltip } from "ui";
import Link from "next/link";

export const MemberTooltip = ({
  member,
  children,
}: {
  member?: Member;
  children: React.ReactNode;
}) => {
  if (!member) {
    return children;
  }

  return (
    <Tooltip
      className="mr-2 py-2.5"
      title={
        <Box>
          <Flex gap={2}>
            <Avatar
              className="mt-0.5"
              name={member?.fullName || member?.username}
              src={member?.avatarUrl}
            />
            <Box>
              <Link className="mb-2 flex gap-1" href={`/profile/${member?.id}`}>
                <Text fontSize="md" fontWeight="medium">
                  {member?.fullName}
                </Text>
                <Text color="muted" fontSize="md">
                  ({member?.username})
                </Text>
              </Link>
              <Button
                className="mb-0.5 ml-px px-2"
                color="tertiary"
                href={`/profile/${member?.id}`}
                size="xs"
              >
                Go to profile
              </Button>
            </Box>
          </Flex>
        </Box>
      }
    >
      {children}
    </Tooltip>
  );
};
