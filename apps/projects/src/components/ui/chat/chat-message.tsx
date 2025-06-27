import { AiIcon, UserIcon } from "icons";
import { Avatar, Box, Text, Flex } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import type { User } from "@/types";
import type { Message } from "./types";

type ChatMessageProps = {
  message: Message;
  profile: User | undefined;
};

export const ChatMessage = ({ message, profile }: ChatMessageProps) => {
  return (
    <Flex
      className={cn({
        "flex-row-reverse": message.sender === "user",
      })}
      gap={3}
    >
      <Flex align="center" className="size-8" justify="center">
        {message.sender === "ai" ? (
          <AiIcon className="h-6" />
        ) : (
          <Avatar
            className=""
            color="primary"
            name={profile?.fullName || profile?.username}
            size="sm"
            src={profile?.avatarUrl}
          />
        )}
      </Flex>

      <Flex
        className={cn("max-w-[75%] flex-1", {
          "items-end": message.sender === "user",
        })}
        direction="column"
      >
        <Box
          className={cn("rounded-2xl p-4", {
            "bg-primary text-white": message.sender === "user",
            "bg-gray-50 dark:bg-dark-100": message.sender === "ai",
          })}
        >
          <Text
            as="div"
            className={cn({
              "text-white": message.sender === "user",
            })}
          >
            <Markdown>{message.content}</Markdown>
          </Text>
        </Box>
        <Text className="mt-2 px-1" color="muted" fontSize="sm">
          {message.timestamp.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </Text>
      </Flex>
    </Flex>
  );
};
