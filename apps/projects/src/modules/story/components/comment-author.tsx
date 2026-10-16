import Link from "next/link";
import { Avatar, Box, Button, Flex, Text, Tooltip } from "ui";
import { cn } from "lib";
import type { Comment } from "@/types/comment";
import { isFormerUser } from "@/lib/former-user";
import { FormerUser } from "@/components/ui/former-user";
import { useWorkspacePath } from "@/hooks/use-workspace-path";

export const CommentAuthor = ({
  userId,
  user,
  isDialog,
}: Pick<Comment, "userId" | "user"> & { isDialog?: boolean }) => {
  const { withWorkspace } = useWorkspacePath();
  return (
    <>
      {isFormerUser(userId) ? (
        <FormerUser
          avatarSurfaceClassName={
            isDialog ? "dark:bg-surface-elevated/95" : undefined
          }
        />
      ) : (
        <Tooltip
          className="py-2.5"
          title={
            <Box>
              <Flex align="center" gap={2}>
                <Avatar
                  className="mt-0.5"
                  name={user.fullName}
                  src={user.avatarUrl}
                />
                <Box>
                  <Link
                    className="mb-2 flex gap-1"
                    href={
                      user.isSystem ? "" : withWorkspace(`/profile/${user.id}`)
                    }
                  >
                    <Text fontSize="md">{user.fullName}</Text>
                    <Text color="muted">({user.username})</Text>
                  </Link>
                  {!user.isSystem ? (
                    <Button
                      className="mb-0.5 ml-px px-2"
                      color="tertiary"
                      href={withWorkspace(`/profile/${user.id}`)}
                      size="xs"
                    >
                      Go to profile
                    </Button>
                  ) : (
                    <Text color="muted" fontSize="md">
                      (System Account)
                    </Text>
                  )}
                </Box>
              </Flex>
            </Box>
          }
        >
          <Flex className="cursor-pointer" gap={1}>
            <Box
              className={cn(
                "bg-surface relative top-px flex aspect-square items-center rounded-full p-[0.3rem]",
                { "dark:bg-surface-elevated/95": isDialog },
              )}
            >
              <Avatar name={user.fullName} size="xs" src={user.avatarUrl} />
            </Box>
            <Text className="relative top-0.5 ml-1 text-black dark:text-white">
              {user.username}
            </Text>
          </Flex>
        </Tooltip>
      )}
    </>
  );
};
