import { Avatar, Box, Text, Flex, Button, Tooltip } from "ui";
import { cn } from "lib";
import Markdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeRaw from "rehype-raw";
import type { ChatStatus } from "ai";
import { useEffect, useState } from "react";
import { CheckIcon, CopyIcon, PlusIcon, ReloadIcon } from "icons";
import { usePathname } from "next/navigation";
import type { User } from "@/types";
import { BurndownChart } from "@/modules/sprints/stories/burndown";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { NewStoryDialog } from "../new-story-dialog";
import { AiIcon } from "./ai";
import { Thinking } from "./thinking";
import { AttachmentsDisplay } from "./attachments-display";

type ChatMessageProps = {
  isLast: boolean;
  message: MayaUIMessage;
  profile: User | undefined;
  status: ChatStatus;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

const RenderMessage = ({
  message,
  onPromptSelect,
  status,
}: {
  isLast: boolean;
  message: MayaUIMessage;
  status: ChatStatus;
  onPromptSelect: (prompt: string) => void;
}) => {
  const [hasText, setHasText] = useState(false);
  const pathname = usePathname();

  useEffect(() => {
    if (message.parts.some((p) => p.type === "text")) {
      setHasText(true);
    }
  }, [message.parts]);

  const isProcessing =
    !hasText && message.parts.some((p) => p.type === "step-start");

  return (
    <>
      {isProcessing ? <Thinking /> : null}
      {message.parts.map((part, index) => {
        if (part.type === "text") {
          return (
            <Box
              className={cn(
                "chat-tables prose prose-stone leading-normal dark:prose-invert prose-table:border prose-table:border-gray-100 prose-img:size-10 prose-img:rounded-full prose-img:object-cover dark:prose-table:border-dark-100",
                {
                  "text-white": message.role === "user",
                },
              )}
              key={index}
            >
              <Markdown rehypePlugins={[rehypeRaw]} remarkPlugins={[remarkGfm]}>
                {part.text}
              </Markdown>
            </Box>
          );
        } else if (part.type === "tool-getSprintDetailsTool") {
          if (part.state === "input-available") {
            return <Thinking key={index} message="Getting sprint details" />;
          }
        } else if (part.type === "tool-listRunningSprints") {
          if (part.state === "input-available") {
            return <Thinking key={index} message="Getting active sprints" />;
          }
        }
        return null;
      })}

      {status === "ready" ? (
        <>
          {message.parts.map((part, index) => {
            if (part.type === "tool-getSprintDetailsTool") {
              if (part.state === "output-available") {
                return (
                  <Box className="mb-3" key={index}>
                    <Text
                      as="h3"
                      className="mb-1 mt-4 text-xl font-semibold antialiased"
                    >
                      Burndown graph
                    </Text>
                    <BurndownChart
                      burndownData={part.output.analyticsReport?.burndown ?? []}
                      className={cn("h-72", {
                        "h-80": pathname === "/maya",
                      })}
                    />
                  </Box>
                );
              }
            } else if (part.type === "tool-suggestions") {
              if (part.state === "output-available") {
                return (
                  <Flex
                    className="mt-2"
                    gap={2}
                    key={`${index}-suggestions`}
                    wrap
                  >
                    {part.output.suggestions.map((suggestion, index) => (
                      <Button
                        color="tertiary"
                        key={index}
                        onClick={() => {
                          onPromptSelect(suggestion);
                        }}
                        size="sm"
                      >
                        {suggestion}
                      </Button>
                    ))}
                  </Flex>
                );
              }
            }
            return null;
          })}
        </>
      ) : null}
    </>
  );
};

export const ChatMessage = ({
  isLast,
  message,
  profile,
  status,
  regenerate,
  onPromptSelect,
}: ChatMessageProps) => {
  const [_, copy] = useCopyToClipboard();
  const [hasCopied, setHasCopied] = useState(false);
  const [isOpen, setIsOpen] = useState(false);
  const { getTermDisplay } = useTerminology();
  const content = message.parts.find((p) => p.type === "text")?.text ?? "";
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
            className={cn("mb-2 rounded-2xl px-4 py-3", {
              "rounded-tr-md bg-primary": message.role === "user",
              "bg-transparent p-0": message.role === "assistant",
            })}
          >
            <RenderMessage
              isLast={isLast}
              message={message}
              onPromptSelect={onPromptSelect}
              status={status}
            />
          </Box>
          <AttachmentsDisplay message={message} />
          <Flex className="mt-2 px-0.5" justify="between">
            {message.role === "assistant" && status !== "streaming" && (
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
                      copy(content).then(() => {
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
                        regenerate();
                      }}
                      size="sm"
                      variant="naked"
                    >
                      <ReloadIcon strokeWidth={2.8} />
                    </Button>
                  </Tooltip>
                ) : null}
              </Flex>
            )}
          </Flex>
        </Flex>
      </Flex>
      <NewStoryDialog
        description={content}
        isOpen={isOpen}
        setIsOpen={setIsOpen}
      />
    </>
  );
};
