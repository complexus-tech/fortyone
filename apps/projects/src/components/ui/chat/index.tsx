"use client";

import { useRef, useState, useEffect } from "react";
import { usePathname, useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useChat } from "@ai-sdk/react";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { useQueryClient } from "@tanstack/react-query";
import { cn } from "lib";
import { ReloadIcon } from "icons";
import { generateId } from "ai";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import {
  notificationKeys,
  sprintKeys,
  statusKeys,
  teamKeys,
} from "@/constants/keys";
import { storyKeys } from "@/modules/stories/constants";
import { objectiveKeys } from "@/modules/objectives/constants";
import { useMediaQuery } from "@/hooks";
import { useSubscription } from "@/lib/hooks/subscriptions/subscription";
import { useTeams } from "@/modules/teams/hooks/teams";
import { fileToBase64 } from "@/lib/utils/files";
import { useAiChatMessages } from "@/modules/ai-chats/hooks/use-ai-chat-messages";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { History } from "./history";
import { SuggestedPrompts } from "./suggested-prompts";

export const Chat = () => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const pathname = usePathname();
  const { data: subscription } = useSubscription();
  const { data: teams = [] } = useTeams();

  const { resolvedTheme, theme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [isFullScreen, setIsFullScreen] = useState(false);
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);
  const [isHistoryOpen, setIsHistoryOpen] = useState(false);
  const [attachments, setAttachments] = useState<File[]>([]);
  const isMobile = useMediaQuery("(max-width: 768px)");
  const idRef = useRef(generateId());
  const { data: aiChatMessages = [], refetch } = useAiChatMessages(
    idRef.current,
  );

  const handleNewChat = () => {
    idRef.current = generateId();
    setMessages([]);
  };

  const handleChatSelect = async (chatId: string) => {
    idRef.current = chatId;
    const { data: newMessages = [] } = await refetch();
    setMessages(newMessages);
    setInput("");
    setAttachments([]);
  };

  const {
    messages,
    input,
    status,
    setInput,
    append,
    stop: handleStop,
    reload,
    error,
    setMessages,
  } = useChat({
    body: {
      currentPath: pathname,
      currentTheme: theme,
      resolvedTheme,
      subscription: {
        tier: subscription?.tier,
        billingInterval: subscription?.billingInterval,
        billingEndsAt: subscription?.billingEndsAt,
        status: subscription?.status,
      },
      teams,
      id: idRef.current,
    },
    sendExtraMessageFields: true, // send id and createdAt for each message
    onFinish: (message) => {
      message.parts?.forEach((part) => {
        if (part.type === "tool-invocation") {
          if (part.toolInvocation.toolName === "navigation") {
            if (part.toolInvocation.state === "result") {
              const result = part.toolInvocation.result;
              if (result.route) {
                router.push(result.route as string);
                setIsOpen(false);
              }
            }
          } else if (part.toolInvocation.toolName === "theme") {
            if (part.toolInvocation.state === "result") {
              const requestedTheme = part.toolInvocation.result.theme as string;
              if (requestedTheme === "toggle") {
                const newTheme = resolvedTheme === "dark" ? "light" : "dark";
                setTheme(newTheme);
              } else {
                setTheme(requestedTheme);
              }
            }
          } else if (part.toolInvocation.toolName === "quickCreate") {
            if (part.toolInvocation.state === "result") {
              const action = part.toolInvocation.result.action as string;
              switch (action) {
                case "story":
                  setIsStoryOpen(true);
                  setIsOpen(false);
                  break;
                case "objective":
                  setIsObjectiveOpen(true);
                  setIsOpen(false);
                  break;
                case "sprint":
                  setIsSprintOpen(true);
                  setIsOpen(false);
                  break;
              }
            }
          } else if (part.toolInvocation.toolName === "teams") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: teamKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "notifications") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: notificationKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "statuses") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: statusKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "stories") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: storyKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "sprints") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: sprintKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "objectives") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: objectiveKeys.all,
              });
            }
          } else if (part.toolInvocation.toolName === "objectiveStatuses") {
            if (part.toolInvocation.state === "result") {
              queryClient.invalidateQueries({
                queryKey: objectiveKeys.statuses(),
              });
            }
          }
        }
      });
    },
    initialMessages: aiChatMessages,
  });

  // Sync messages when aiChatMessages changes (when selecting a different chat)
  useEffect(() => {
    if (aiChatMessages.length > 0) {
      setMessages(aiChatMessages);
    }
  }, [aiChatMessages, setMessages]);

  const isLoading = status === "submitted";

  const handleSendMessage = async (content: string) => {
    if (!content.trim() && attachments.length === 0) return;

    // Convert attachments to base64 for AI SDK
    const attachmentData = await Promise.all(
      attachments.map(async (file) => ({
        name: file.name,
        contentType: file.type,
        url: await fileToBase64(file),
      })),
    );

    append({
      role: "user",
      content:
        content ||
        `Analyze the attached file${attachmentData.length > 1 ? "s" : ""}.`,
      experimental_attachments:
        attachmentData.length > 0 ? attachmentData : undefined,
    });

    setInput("");
    setAttachments([]);
  };

  const handleSuggestedPrompt = (prompt: string) => {
    handleSendMessage(prompt);
  };

  const handleSend = () => {
    if (!input.trim() && attachments.length === 0) return;
    handleSendMessage(input);
  };

  return (
    <>
      <ChatButton
        isOpen={isOpen}
        onOpen={() => {
          setIsOpen(true);
        }}
      />
      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
      />
      <NewSprintDialog isOpen={isSprintOpen} setIsOpen={setIsSprintOpen} />
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className={cn(
            "max-w-[36rem] rounded-[2rem] border-[0.5px] border-gray-200/90 font-medium outline-none backdrop-blur-lg dark:bg-dark-300/90 md:mb-[2.6vh] md:mt-auto",
            {
              "m-0 h-dvh w-screen max-w-[100vw] rounded-none border-0 backdrop-blur-lg dark:bg-dark md:mb-0 md:mt-0":
                isFullScreen || isMobile,
            },
          )}
          hideClose
          overlayClassName={cn("justify-end pr-[1.5vh]", {
            "pr-0 justify-center": isFullScreen || isMobile,
          })}
        >
          <Dialog.Header
            className={cn("flex h-[4.5rem] items-center px-6", {
              "absolute left-0 right-0 top-0": isFullScreen || isMobile,
              "border-b-[0.5px] border-gray-100 dark:border-dark-100":
                isHistoryOpen && !isFullScreen && !isMobile,
            })}
          >
            <Dialog.Title className="w-full text-lg">
              <ChatHeader
                handleNewChat={handleNewChat}
                isFullScreen={isFullScreen}
                isHistoryOpen={isHistoryOpen}
                setIsFullScreen={setIsFullScreen}
                setIsHistoryOpen={setIsHistoryOpen}
                setIsOpen={setIsOpen}
              />
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="sr-only">
            Maya is your AI assistant.
          </Dialog.Description>
          <Dialog.Body
            className={cn("h-[83dvh] max-h-[83dvh] p-0", {
              "h-dvh max-h-dvh md:h-dvh md:max-h-dvh": isFullScreen || isMobile,
            })}
          >
            <Flex
              className={cn("h-full", {
                "mx-auto h-dvh max-w-3xl pt-16": isFullScreen || isMobile,
                "pt-4": isHistoryOpen && !isFullScreen && !isMobile,
              })}
              direction="column"
            >
              {isHistoryOpen ? (
                <History
                  currentChatId={idRef.current}
                  handleChatSelect={handleChatSelect}
                  handleNewChat={handleNewChat}
                  setIsHistoryOpen={setIsHistoryOpen}
                />
              ) : (
                <>
                  <ChatMessages
                    isFullScreen={isFullScreen}
                    isLoading={isLoading}
                    isStreaming={status === "streaming"}
                    messages={messages}
                    reload={reload}
                    value={input}
                  />
                  {error ? (
                    <Box className="mb-4 px-6">
                      <Text>{error.message || "An error occurred."} </Text>
                      <Button
                        className="mt-2"
                        leftIcon={
                          <ReloadIcon className="text-white dark:text-white" />
                        }
                        onClick={() => reload()}
                      >
                        Retry
                      </Button>
                    </Box>
                  ) : null}
                  {messages.length === 0 && (
                    <SuggestedPrompts onPromptSelect={handleSuggestedPrompt} />
                  )}
                  <ChatInput
                    attachments={attachments}
                    isLoading={isLoading}
                    onAttachmentsChange={setAttachments}
                    onChange={(e) => {
                      setInput(e.target.value);
                    }}
                    onSend={handleSend}
                    onStop={handleStop}
                    value={input}
                  />
                </>
              )}
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
