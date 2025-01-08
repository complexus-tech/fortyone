import { Box, Flex, Text, Avatar, TimeAgo, Tooltip, Button } from "ui";
import { useStoryComments } from "@/modules/story/hooks/story-comments";
import { useMembers } from "@/lib/hooks/members";
import { Comment } from "@/types";
import Link from "next/link";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import { ReplyIcon } from "icons";

const MainComment = ({
  comment: { userId, createdAt, comment, subComments },
  isSubComment = false,
  className,
}: {
  comment: Comment;
  isSubComment?: boolean;
  className?: string;
}) => {
  const { data: members = [] } = useMembers();
  const { data: session } = useSession();
  const member = members.find(
    (member) => member.id === (userId || session?.user?.id),
  );
  return (
    <Box
      className={cn(
        "relative pb-3",
        {
          "ml-10 border-l-2 border-gray-200 pb-0 pl-1.5 pt-1 dark:border-dark-50":
            isSubComment,
        },
        className,
      )}
    >
      <Flex align="center" gap={1}>
        <Tooltip
          className="py-2.5"
          title={
            member && (
              <Box>
                <Flex gap={2}>
                  <Avatar
                    name={member?.fullName}
                    src={member?.avatarUrl}
                    className="mt-0.5"
                  />
                  <Box>
                    <Link
                      href={`/profile/${member?.id}`}
                      className="mb-2 flex gap-1"
                    >
                      <Text fontWeight="medium" fontSize="md">
                        {member?.fullName}
                      </Text>
                      <Text color="muted" fontSize="md">
                        ({member?.username})
                      </Text>
                    </Link>
                    <Button
                      size="xs"
                      color="tertiary"
                      className="mb-0.5 ml-px px-2"
                      href={`/profile/${member?.id}`}
                    >
                      Go to profile
                    </Button>
                  </Box>
                </Flex>
              </Box>
            )
          }
        >
          <Flex gap={1} className="cursor-pointer">
            <Box className="relative top-[1px] flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-300">
              <Avatar
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
            </Box>
            <Text
              className="relative top-0.5 ml-1 text-black dark:text-white"
              fontWeight="medium"
            >
              {member?.username}
            </Text>
          </Flex>
        </Tooltip>
        <Text className="mx-0.5 text-[0.95rem]" color="muted">
          ·
        </Text>
        <Text className="text-[0.95rem]" color="muted">
          <TimeAgo timestamp={createdAt} />
        </Text>
      </Flex>
      <Box
        className="prose prose-stone ml-9 mt-0.5 max-w-full leading-6 antialiased dark:prose-invert prose-headings:font-medium prose-a:text-primary prose-pre:bg-gray-50 prose-pre:text-dark-200 dark:prose-pre:bg-dark-200/80 dark:prose-pre:text-gray-200"
        html={comment}
      />
      {subComments.length > 0 && (
        <Box className="mt-2">
          {subComments.map((subComment) => (
            <MainComment
              key={subComment.id}
              comment={subComment}
              isSubComment
            />
          ))}
        </Box>
      )}
      {!isSubComment && (
        <Button
          size="sm"
          color="tertiary"
          className="ml-10 mt-3 px-2"
          leftIcon={<ReplyIcon className="h-4" />}
        >
          Reply
        </Button>
      )}
    </Box>
  );
};

export const Comments = ({ storyId }: { storyId: string }) => {
  const { data: comments = [] } = useStoryComments(storyId);

  return (
    <Box>
      {comments.map((comment) => (
        <MainComment key={comment.id} comment={comment} />
      ))}
    </Box>
  );
};
