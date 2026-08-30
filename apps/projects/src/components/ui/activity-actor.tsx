import Link from "next/link";
import { Avatar, Box, Button, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import { MayaAvatar } from "./maya-avatar";

type ActivityActorMember = {
  avatarUrl?: string | null;
  fullName: string;
  id: string;
  isSystem?: boolean;
  username: string;
};

type ActivityActorProps = {
  avatarSurfaceClassName?: string;
  displayName: string;
  displayUsername: string;
  isSelfActivity: boolean;
  member: ActivityActorMember;
  withWorkspace: (path: string) => string;
};

export const ActivityActor = ({
  avatarSurfaceClassName,
  displayName,
  displayUsername,
  isSelfActivity,
  member,
  withWorkspace,
}: ActivityActorProps) => (
  <Tooltip
    className="py-2.5"
    title={
      <Box>
        <Flex gap={2}>
          {member.isSystem ? (
            <MayaAvatar
              className="mt-0.5"
              name={member.fullName}
              size="md"
              src={member.avatarUrl}
            />
          ) : (
            <Avatar
              className="mt-0.5"
              name={member.fullName}
              src={member.avatarUrl}
            />
          )}
          <Box>
            <Link
              className={cn("mb-2 flex gap-1", {
                "mb-0": member.isSystem,
              })}
              href={member.isSystem ? "" : `/profile/${member.id}`}
            >
              <Text fontSize="md" fontWeight="medium">
                {displayName}
              </Text>
              {!isSelfActivity ? (
                <Text color="muted" fontSize="md">
                  ({member.username})
                </Text>
              ) : null}
            </Link>
            {!member.isSystem ? (
              <Button
                className="mb-0.5 ml-px px-2"
                color="tertiary"
                href={withWorkspace(`/profile/${member.id}`)}
                size="xs"
              >
                Go to profile
              </Button>
            ) : (
              <Text color="muted" fontSize="md">
                ({member.username === "maya" ? "AI Agent" : "Bot"})
              </Text>
            )}
          </Box>
        </Flex>
      </Box>
    }
  >
    <Flex align="center" className="shrink-0 cursor-pointer" gap={1}>
      <Box
        className={cn(
          "bg-surface relative left-px flex aspect-square items-center rounded-full p-[0.3rem]",
          avatarSurfaceClassName,
        )}
      >
        {member.isSystem ? (
          <MayaAvatar name={member.fullName} size="xs" src={member.avatarUrl} />
        ) : (
          <Avatar name={member.fullName} size="xs" src={member.avatarUrl} />
        )}
      </Box>
      <Text
        className="relative ml-1 text-sm text-black md:text-[0.95rem] dark:text-white"
        fontWeight="medium"
      >
        {displayUsername}
      </Text>
    </Flex>
  </Tooltip>
);
