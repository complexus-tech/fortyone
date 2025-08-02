import { Avatar, Box, Text, Flex, Button, Tooltip } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import type { Message } from "@ai-sdk/react";
import { useEffect, useState } from "react";
import { CheckIcon, CopyIcon, PlusIcon, ReloadIcon } from "icons";
import type { ChatRequestOptions } from "ai";
import { usePathname } from "next/navigation";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import { NewStoryDialog } from "../new-story-dialog";
import { AiIcon } from "./ai";
import { Thinking } from "./thinking";
import { AttachmentsDisplay } from "./attachments-display";

type ChatMessageProps = {
  isLast: boolean;
  message: Message;
  profile: User | undefined;
  isStreaming?: boolean;
  reload: (
    chatRequestOptions?: ChatRequestOptions,
  ) => Promise<string | null | undefined>;
  onPromptSelect: (prompt: string) => void;
  isOnPage?: boolean;
};

const RenderMessage = ({
  isLast,
  message,
  isStreaming,
  onPromptSelect,
  isOnPage,
}: {
  isLast: boolean;
  message: Message;
  isStreaming?: boolean;
  onPromptSelect: (prompt: string) => void;
  isOnPage?: boolean;
}) => {
  const [hasText, setHasText] = useState(false);
  const pathname = usePathname();

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
                  "text-white": message.role === "user",
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
                message={`Invoking ${toolInvocation.toolName}`}
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
                        "h-80": pathname === "/maya",
                      })}
                    />
                  </Box>
                );
              }
            } else if (toolInvocation.toolName === "suggestions") {
              const { result } = toolInvocation;
              if (result?.suggestions && isLast && !isStreaming) {
                return (
                  <Flex
                    className="mt-2"
                    gap={2}
                    key={`${index}-suggestions`}
                    wrap
                  >
                    {result.suggestions.map(
                      (suggestion: string, index: number) => (
                        <Button
                          color="tertiary"
                          key={index}
                          onClick={() => {
                            onPromptSelect(suggestion);
                          }}
                          size={isOnPage ? "md" : "sm"}
                        >
                          {suggestion}
                        </Button>
                      ),
                    )}
                  </Flex>
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
  isLast,
  message,
  profile,
  isStreaming,
  reload,
  onPromptSelect,
  isOnPage,
}: ChatMessageProps) => {
  const [_, copy] = useCopyToClipboard();
  const [hasCopied, setHasCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  return (
    <>
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
            src={profile?.avatarUrl}
          />
        )}
        <Flex
          className={cn("max-w-[80%] flex-1", {
            "items-end": message.role === "user",
            "max-w-[85%]": message.role === "assistant",
          })}
          direction="column"
        >
          <Box
            className={cn("rounded-2xl px-4 py-3", {
              "rounded-tr-md bg-primary": message.role === "user",
              "bg-transparent p-0": message.role === "assistant",
            })}
          >
            <RenderMessage
              isLast={isLast}
              isOnPage={isOnPage}
              isStreaming={isStreaming}
              message={message}
              onPromptSelect={onPromptSelect}
            />
          </Box>
          <AttachmentsDisplay attachments={message.experimental_attachments} />
          <Flex className="mt-2 px-0.5" justify="between">
            {message.role === "assistant" && !isStreaming && (
              <Flex gap={2} justify="end">
                <Tooltip title={`Create ${getTermDisplay("storyTerm")}`}>
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      setIsOpen(true);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <PlusIcon />
                  </Button>
                </Tooltip>
                <Tooltip title="Copy">
                  <Button
                    asIcon
                    color="tertiary"
                    onClick={() => {
                      copy(message.content).then(() => {
                        setHasCopied(true);
                        setTimeout(() => {
                          setHasCopied(false);
                        }, 1500);
                      });
                    }}
                    size="sm"
                    variant="naked"
                  >
                    {hasCopied ? <CheckIcon /> : <CopyIcon />}
                  </Button>
                </Tooltip>
                {isLast ? (
                  <Tooltip title="Retry">
                    <Button
                      asIcon
                      color="tertiary"
                      onClick={() => {
                        reload();
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <ReloadIcon strokeWidth={2} />
                    </Button>
                  </Tooltip>
                ) : null}
              </Flex>
            )}
          </Flex>
        </Flex>
      </Flex>
      <NewStoryDialog
        description={message.content}
        isOpen={isOpen}
        setIsOpen={setIsOpen}
      />
    </>
  );
};
