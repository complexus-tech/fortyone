import { Button, Flex } from "ui";
import { ArrowUp2Icon, ReplyIcon } from "icons";
import { cn } from "lib";
import type {
  CommentEmoji,
  CommentReaction,
} from "../types/comment-collaboration";
import { CommentReactions } from "./comment-reactions";

export const CommentFooter = ({
  commentId,
  reactions,
  disabled,
  onReact,
  onReply,
  onCollapse,
}: {
  commentId: string;
  reactions: CommentReaction[];
  disabled: boolean;
  onReact: (emoji: CommentEmoji, reacted: boolean) => void;
  onReply?: () => void;
  onCollapse?: () => void;
}) => (
  <Flex align="center" className="comment-footer mt-2 ml-9 flex-wrap" gap={3}>
    {onReply ? (
      <Button
        className="-mx-2 text-[0.95rem]"
        color="tertiary"
        disabled={disabled}
        leftIcon={<ReplyIcon className="h-4" />}
        onClick={onReply}
        size="sm"
        variant="naked"
      >
        Reply
      </Button>
    ) : null}
    {onCollapse ? (
      <Button
        aria-controls={`comment-replies-${commentId}`}
        aria-expanded
        className={cn("-mr-2 text-[0.95rem]", !onReply && "-ml-2")}
        color="tertiary"
        leftIcon={<ArrowUp2Icon className="h-4" />}
        onClick={onCollapse}
        size="sm"
        variant="naked"
      >
        Hide replies
      </Button>
    ) : null}
    <CommentReactions
      disabled={disabled}
      onReact={onReact}
      reactions={reactions}
    />
  </Flex>
);
