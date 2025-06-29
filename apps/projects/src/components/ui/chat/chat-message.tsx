import { Avatar, Box, Text, Flex } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Message } from "@ai-sdk/react";
import { useEffect, useState } from "react";
import type { User } from "@/types";
import { AiIcon } from "./ai";

type ChatMessageProps = {
  message: Message;
  profile: User | undefined;
};

const RenderMessage = ({ message }: { message: Message }) => {
  const [hasText, setHasText] = useState(false);

  useEffect(() => {
    if (message.parts?.some((p) => p.type === "text")) {
      setHasText(true);
    }
  }, [message.parts]);

  const isProcessing =
    !hasText && message.parts?.some((p) => p.type === "step-start");

  return (
    <>
      {isProcessing ? <Text>Processing…</Text> : null}
      {message.parts?.map((part, index) => {
        if (part.type === "text") {
          return (
            <Box
              className={cn(
                "chat-tables prose prose-stone leading-normal dark:prose-invert prose-a:text-primary",
                {
                  "text-white": message.role === "user",
                },
              )}
              key={part.text}
            >
              <Markdown remarkPlugins={[remarkGfm]}>{part.text}</Markdown>
            </Box>
          );
        } else if (part.type === "tool-invocation") {
          const toolInvocation = part.toolInvocation;
          if (toolInvocation.state === "call") {
            return (
              <Text key={index}>Using {toolInvocation.toolName} tool...</Text>
            );
          }
        }
        return null;
      })}
    </>
  );
};

export const ChatMessage = ({ message, profile }: ChatMessageProps) => {
  const createdAt = message.createdAt || new Date();
  return (
    <Flex
      className={cn({
        "flex-row-reverse": message.role === "user",
      })}
      gap={3}
    >
      {message.role === "assistant" ? (
        <AiIcon />
      ) : (
        <Avatar
          color="tertiary"
          name={profile?.fullName || profile?.username}
          size="sm"
          src={profile?.avatarUrl}
        />
      )}
      <Flex
        className={cn("max-w-[75%] flex-1", {
          "items-end": message.role === "user",
        })}
        direction="column"
      >
        <Box
          className={cn("rounded-full p-4", {
            "bg-primary": message.role === "user",
            "bg-transparent p-0": message.role === "assistant",
          })}
        >
          <RenderMessage message={message} />
        </Box>
        <Text className="mt-2 px-1" color="muted" fontSize="sm">
          {createdAt.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
          })}
        </Text>
      </Flex>
    </Flex>
  );
};
