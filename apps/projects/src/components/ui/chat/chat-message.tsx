import { Avatar, Box, Flex, Button, Tooltip } from "ui";
import { cn } from "lib";
import type { ChatStatus } from "ai";
import { useState, type ComponentProps } from "react";
import { CheckIcon, CopyIcon, PlusIcon, ReloadIcon } from "icons";
import { Streamdown, type StreamdownProps } from "streamdown";
import type { User } from "@/types";
import { useCopyToClipboard, useTerminology } from "@/hooks";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { NewStoryDialog } from "../new-story-dialog";
import { AttachmentsDisplay } from "./attachments-display";
import { getMessageText } from "./chat-message-utils";
import { isToolMessagePart } from "./tool-output-policy";
import { ToolOutputRenderer } from "./tool-output-renderer";

type ChatMessageProps = {
  isAnimating?: boolean;
  isLast: boolean;
  message: MayaUIMessage;
  profile: User | undefined;
  status: ChatStatus;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

const LinkText = ({ children }: ComponentProps<"a">) => <>{children}</>;

const STREAMDOWN_COMPONENTS: NonNullable<StreamdownProps["components"]> = {
  a: LinkText,
};

const RenderMessage = ({
  isAnimating,
  message,
  onPromptSelect,
}: {
  isAnimating: boolean;
  message: MayaUIMessage;
  onPromptSelect: (prompt: string) => void;
}) => {
  return (
    <>
      {message.parts.map((part, index) => {
        if (part.type === "text") {
          return (
            <Streamdown
              className="chat-tables"
              components={STREAMDOWN_COMPONENTS}
              controls={{
                table: true,
                code: true,
                mermaid: {
                  download: true,
                  copy: true,
                  fullscreen: true,
                  panZoom: true,
                },
              }}
              isAnimating={isAnimating}
              key={`${message.id}-text-${index}`}
            >
              {part.text}
            </Streamdown>
          );
        }

        if (isToolMessagePart(part)) {
          return (
            <ToolOutputRenderer
              key={`${message.id}-${part.type}-${index}`}
              onPromptSelect={onPromptSelect}
              part={part}
            />
          );
        }

        return null;
      })}
    </>
  );
};

export const ChatMessage = ({
  isAnimating = false,
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
  const content = getMessageText(message);
  return (
    <>
      <Flex
        className={cn({
          "flex-row-reverse": message.role === "user",
        })}
        gap={message.role === "user" ? 3 : 0}
      >
        {message.role === "user" ? (
          <Avatar
            color="tertiary"
            name={profile?.fullName || profile?.username}
            src={profile?.avatarUrl}
          />
        ) : null}
        <Flex
          className={cn("flex-1", {
            "items-end": message.role === "user",
            "max-w-[80%]": message.role === "user",
            "max-w-full": message.role === "assistant",
          })}
          direction="column"
        >
          <Box
            className={cn("mb-2 rounded-2xl px-4 py-3", {
              "bg-state-hover/80 rounded-tr-md dark:bg-white/[0.08]":
                message.role === "user",
              "bg-transparent p-0": message.role === "assistant",
            })}
          >
            <RenderMessage
              isAnimating={isAnimating}
              message={message}
              onPromptSelect={onPromptSelect}
            />
          </Box>
          <AttachmentsDisplay message={message} />
          <Flex className="mt-2 px-0.5" justify="between">
            {message.role === "assistant" &&
            status !== "streaming" &&
            !isAnimating ? (
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
                {isLast && message.metadata?.source !== "voice" ? (
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
            ) : null}
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
