import React from "react";
import { Avatar, Row, Text } from "@/components/ui";
import { View } from "react-native";
import { formatDistanceToNow } from "date-fns";
import { useMembers } from "@/modules/members/hooks/use-members";
import type { Comment } from "@/types";
import { RichTextViewer } from "@/components/rich-text/viewer";
import { plainTextToHtml } from "@/components/rich-text/content";
import { resolveCommentAuthor } from "./comment-author";

export const CommentItem = ({ userId, user, comment, createdAt }: Comment) => {
  const { data: members = [] } = useMembers();
  const author = resolveCommentAuthor({ userId, user }, members);

  return (
    <View className="my-[6px] rounded-[20px] bg-gray-50 p-[14px] dark:bg-dark-100">
      <Row align="center" wrap>
        <Avatar
          name={author.name}
          src={author.avatarUrl}
          size="xs"
          className="mr-2"
        />
        <Text fontSize="sm" fontWeight="bold">
          {author.name}
        </Text>
        <Text fontSize="xs" color="muted" className="ml-2">
          {formatDistanceToNow(new Date(createdAt), { addSuffix: true })}
        </Text>
      </Row>
      <View className="mt-[6px]">
        <RichTextViewer
          html={
            /<(?:p|ul|ol|blockquote|h[1-6]|pre|a|strong|em|span)(?:\s|>)/i.test(
              comment,
            )
              ? comment
              : plainTextToHtml(comment)
          }
        />
      </View>
    </View>
  );
};
