import { Avatar, Box, Text, Flex } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import type { Message } from "@ai-sdk/react";
import { useEffect, useState } from "react";
import { BrainIcon } from "icons";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { AiIcon } from "./ai";

type ChatMessageProps = {
  message: Message;
  profile: User | undefined;
  isStreaming?: boolean;
};

const RenderMessage = ({
  message,
  isStreaming,
}: {
  message: Message;
  isStreaming?: boolean;
}) => {
  const [hasText, setHasText] = useState(false);

  useEffect(() => {
    if (message.parts?.some((p) => p.type === "text")) {
      setHasText(true);
    }
  }, [message.parts]);

  const isProcessing =
    !hasText &&
    message.parts?.some((p) => p.type === "step-start") &&
    !message.parts.some(
      (p) => p.type === "tool-invocation" && p.toolInvocation.state === "call",
    );

  return (
    <>
      {isProcessing ? (
        <Flex align="center" className="gap-1.5">
          <BrainIcon className="h-4 animate-pulse" />
          <Text>Thinking…</Text>
        </Flex>
      ) : null}
      {message.parts?.map((part, index) => {
        if (part.type === "text") {
          return (
            <Box
              className={cn(
                "chat-tables prose prose-stone leading-normal dark:prose-invert prose-a:text-primary prose-table:border prose-table:border-gray-100 prose-img:size-10 prose-img:rounded-full prose-img:object-cover dark:prose-table:border-dark-100",
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
              <Flex align="center" className="gap-1.5" key={index}>
                <BrainIcon className="h-4 animate-pulse" />
                <Text>Using {toolInvocation.toolName} tool...</Text>
              </Flex>
            );
          }
          if (toolInvocation.state === "result") {
            if (toolInvocation.toolName === "sprints") {
              const { result } = toolInvocation;
              if (result?.analytics?.burndown && !isStreaming) {
                return (
                  <Box className="mb-3" key={index}>
                    <Text
                      as="h3"
                      className="mb-2 text-lg font-semibold antialiased"
                    >
                      Burndown graph
                    </Text>
                    <BurndownChart burndownData={result?.analytics?.burndown} />
                  </Box>
                );
              }
            }
          }
        }
        return null;
      })}
    </>
  );
};

export const ChatMessage = ({
  message,
  profile,
  isStreaming,
}: ChatMessageProps) => {
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
          <RenderMessage isStreaming={isStreaming} message={message} />
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
