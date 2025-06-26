import { AiIcon } from "icons";
import { Avatar, Box, Text, Flex } from "ui";
import { cn } from "lib";
import type { Message } from "./types";

type ChatMessageProps = {
  message: Message;
};

export const ChatMessage = ({ message }: ChatMessageProps) => {
  return (
    <Flex
      className={cn({
        "flex-row-reverse": message.sender === "user",
      })}
      gap={3}
    >
      <Avatar
        className={cn({
          "bg-gradient-to-br from-primary to-secondary":
            message.sender === "ai",
        })}
        color={message.sender === "ai" ? "primary" : "secondary"}
        name={message.sender === "user" ? "You" : "Maya"}
        size="md"
      >
        {message.sender === "ai" && <AiIcon />}
      </Avatar>
      <Flex
        className={cn("max-w-[75%] flex-1", {
          "items-end": message.sender === "user",
        })}
        direction="column"
      >
        <Box
          className={cn("rounded-2xl rounded-tl-none p-4", {
            "bg-primary text-white": message.sender === "user",
            "bg-gray-50 dark:bg-dark-50": message.sender === "ai",
          })}
        >
          <Text
            className={cn({
              "text-white": message.sender === "user",
            })}
            fontSize="md"
          >
            {message.content}
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
