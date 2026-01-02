import {
  Box,
  Flex,
  Text,
  Avatar,
  TimeAgo,
  Tooltip,
  Button,
  Dialog,
  Skeleton,
} from "ui";
import Link from "next/link";
import { cn } from "lib";
import { useSession } from "next-auth/react";
import { DeleteIcon, EditIcon, ReplyIcon } from "icons";
import { useState } from "react";
import type { Comment } from "@/types";
import { useMembers } from "@/lib/hooks/members";
import { useStoryCommentsInfinite } from "@/modules/story/hooks/story-comments";
import { CommentInput } from "@/modules/story/components/comment-input";
import { useDeleteCommentMutation } from "@/lib/hooks/delete-comment-mutation";

const MainComment = ({
  storyId,
  comment: { id, userId, createdAt, comment, subComments },
  isSubComment = false,
  className,
  teamId,
}: {
  comment: Comment;
  isSubComment?: boolean;
  className?: string;
  storyId: string;
  teamId: string;
}) => {
  const { data: members = [] } = useMembers();
  const { data: session } = useSession();
  const { mutate: deleteComment } = useDeleteCommentMutation();
  const [isOpen, setIsOpen] = useState(false);
  const [isReplying, setIsReplying] = useState(false);
  const [isEditing, setIsEditing] = useState(false);
  const isOwner = userId === session?.user?.id;

  const member = members.find(
    (member) => member.id === (userId || session?.user?.id),
  );

  const handleCancel = () => {
    setIsReplying(false);
    setIsEditing(false);
  };

  const handleDelete = () => {
    setIsOpen(false);
    deleteComment({ commentId: id, storyId });
  };

  return (
    <Box
      className={cn(
        "relative pb-3",
        {
          "ml-10 border-l-2 border-border pb-0 pl-1.5 pt-1 d":
            isSubComment,
        },
        className,
      )}
    >
      <Flex align="center" className="group" gap={1}>
        <Tooltip
          className="py-2.5"
          title={
            member ? (
              <Box>
                <Flex align="center" gap={2}>
                  <Avatar
                    className="mt-0.5"
                    name={member.fullName}
                    src={member.avatarUrl}
                  />
                  <Box>
                    <Link
                      className="mb-2 flex gap-1"
                      href={`/profile/${member.id}`}
                    >
                      <Text fontSize="md">{member.fullName}</Text>
                      <Text color="muted">({member.username})</Text>
                    </Link>
                    <Button
                      className="mb-0.5 ml-px px-2"
                      color="tertiary"
                      href={`/profile/${member.id}`}
                      size="xs"
                    >
                      Go to profile
                    </Button>
                  </Box>
                </Flex>
              </Box>
            ) : null
          }
        >
          <Flex className="cursor-pointer" gap={1}>
            <Box className="relative top-px flex aspect-square items-center rounded-full bg-white p-[0.3rem] bg-surface-elevated">
              <Avatar
                name={member?.fullName}
                size="xs"
                src={member?.avatarUrl}
              />
            </Box>
            <Text className="relative top-0.5 ml-1 text-black dark:text-white">
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
        <Flex
          align="center"
          className="pointer-events-none ml-2 gap-2.5 opacity-0 transition duration-200 ease-linear group-hover:pointer-events-auto group-hover:opacity-100"
        >
          {isOwner ? (
            <button
              onClick={() => {
                setIsEditing(true);
              }}
              title="Edit"
              type="button"
            >
              <EditIcon className="h-[1.2rem] transition hover:text-dark dark:hover:text-white" />
              <span className="sr-only">Edit</span>
            </button>
          ) : null}
          {isOwner ? (
            <button
              onClick={() => {
                setIsOpen(true);
              }}
              title="Delete"
              type="button"
            >
              <DeleteIcon className="h-[1.1rem] transition hover:text-dark dark:hover:text-white" />
              <span className="sr-only">Delete</span>
            </button>
          ) : null}
        </Flex>
      </Flex>
      {!isEditing && (
        <Box
          className="prose prose-stone ml-9 mt-0.5 max-w-full leading-6 dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:bg-gray-50 prose-pre:text-dark-200 dark:prose-pre:bg-dark-200/80 dark:prose-pre:text-gray-200"
          html={comment}
        />
      )}

      {subComments.length > 0 && (
        <Box className="mt-2">
          {subComments.map((subComment) => (
            <MainComment
              comment={subComment}
              isSubComment
              key={subComment.id}
              storyId={storyId}
              teamId={teamId}
            />
          ))}
        </Box>
      )}
      {!isSubComment && !isReplying && !isEditing && (
        <Button
          className="ml-10 mt-3 px-2"
          color="tertiary"
          leftIcon={<ReplyIcon className="h-4" />}
          onClick={() => {
            setIsReplying(true);
          }}
          size="sm"
        >
          Reply
        </Button>
      )}
      {isReplying ? (
        <Box className="mt-3 pl-[2.4rem] pr-1">
          <CommentInput
            className="mb-2 min-h-12 focus-within:shadow-none"
            onCancel={handleCancel}
            parentId={id}
            storyId={storyId}
            teamId={teamId}
          />
        </Box>
      ) : null}

      {isEditing ? (
        <Box className="mt-3 pl-8 pr-1">
          <CommentInput
            className="mb-2 min-h-12 focus-within:shadow-none"
            commentId={id}
            initialComment={comment}
            onCancel={handleCancel}
            storyId={storyId}
            teamId={teamId}
          />
        </Box>
      ) : null}

      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content>
          <Dialog.Header className="px-6 pt-6">
            <Dialog.Title className="text-lg">
              Are you sure you want to delete this comment?
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="pt-0">
            <Text color="muted">
              {isSubComment
                ? "This comment will be permanently deleted."
                : "This comment will be permanently deleted and all replies will be deleted."}
            </Text>
            <Flex align="center" className="mt-4" gap={2} justify="end">
              <Button
                color="tertiary"
                onClick={() => {
                  setIsOpen(false);
                }}
              >
                Cancel
              </Button>
              <Button
                leftIcon={
                  <DeleteIcon className="h-[1.15rem] text-white dark:text-gray-200" />
                }
                onClick={handleDelete}
              >
                Delete
              </Button>
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};

export const Comments = ({
  storyId,
  teamId,
}: {
  storyId: string;
  teamId: string;
}) => {
  const {
    data: infiniteData,
    hasNextPage,
    fetchNextPage,
    isFetchingNextPage,
  } = useStoryCommentsInfinite(storyId);

  const allComments =
    infiniteData?.pages.flatMap((page) => page.comments) ?? [];

  const handleLoadMore = () => {
    fetchNextPage();
  };

  return (
    <Box>
      {allComments.map((comment) => (
        <MainComment
          comment={comment}
          key={comment.id}
          storyId={storyId}
          teamId={teamId}
        />
      ))}

      {isFetchingNextPage ? (
        <Box className="mt-4 space-y-3">
          {Array.from({ length: 2 }).map((_, i) => (
            <Box className="flex gap-3" key={i}>
              <Skeleton className="h-8 w-8 rounded-full" />
              <Box className="flex-1 space-y-2">
                <Skeleton className="h-4 w-32" />
                <Skeleton className="h-12 w-full" />
              </Box>
            </Box>
          ))}
        </Box>
      ) : null}

      {hasNextPage ? (
        <Button
          className="ml-6 px-3 text-[0.95rem]"
          color="tertiary"
          disabled={isFetchingNextPage}
          onClick={handleLoadMore}
          size="sm"
          variant="naked"
        >
          {isFetchingNextPage ? "Loading..." : "Load more comments"}
        </Button>
      ) : null}
    </Box>
  );
};
