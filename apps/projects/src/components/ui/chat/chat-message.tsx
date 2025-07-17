import { Avatar, Box, Text, Flex } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import type { Message } from "@ai-sdk/react";
import { useEffect, useState } from "react";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { AiIcon } from "./ai";
import { Thinking } from "./thinking";

type ChatMessageProps = {
  isFullScreen: boolean;
  message: Message;
  profile: User | undefined;
  isStreaming?: boolean;
};

const RenderMessage = ({
  isFullScreen,
  message,
  isStreaming,
}: {
  isFullScreen: boolean;
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
      {isProcessing ? <Thinking /> : null}
      {message.parts?.map((part, index) => {
        if (part.type === "text") {
          return (
            <Box
              className={cn(
                "chat-tables prose prose-stone leading-normal dark:prose-invert prose-table:border prose-table:border-gray-100 prose-img:size-10 prose-img:rounded-full prose-img:object-cover dark:prose-table:border-dark-100",
                {
                  "text-white dark:text-dark": message.role === "user",
                },
              )}
              key={part.text}
            >
              <Markdown rehypePlugins={[rehypeRaw]} remarkPlugins={[remarkGfm]}>
                {part.text}
              </Markdown>
            </Box>
          );
        } else if (part.type === "tool-invocation") {
          const toolInvocation = part.toolInvocation;
          if (toolInvocation.state === "call") {
            return (
              <Thinking
                key={index}
                message={`Using ${toolInvocation.toolName} tool...`}
              />
            );
          }
        }
        return null;
      })}

      {message.parts?.map((part, index) => {
        if (part.type === "tool-invocation") {
          const toolInvocation = part.toolInvocation;
          if (toolInvocation.state === "result") {
            if (toolInvocation.toolName === "sprints") {
              const { result } = toolInvocation;
              if (result?.analytics?.burndown && !isStreaming) {
                return (
                  <Box className="mb-3" key={index}>
                    <Text
                      as="h3"
                      className="mb-1 mt-4 text-xl font-semibold antialiased"
                    >
                      Burndown graph
                    </Text>
                    <BurndownChart
                      burndownData={result?.analytics?.burndown}
                      className={cn("h-72", {
                        "h-80": isFullScreen,
                      })}
                    />
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
  isFullScreen,
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
          "max-w-[85%]": message.role === "assistant",
          "md:max-w-[90%]": isFullScreen,
        })}
        direction="column"
      >
        <Box
          className={cn("rounded-2xl p-4", {
            "rounded-tr-md bg-dark dark:bg-white dark:text-dark":
              message.role === "user",
            "bg-transparent p-0": message.role === "assistant",
          })}
        >
          <RenderMessage
            isFullScreen={isFullScreen}
            isStreaming={isStreaming}
            message={message}
          />
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
