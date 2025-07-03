"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useChat } from "@ai-sdk/react";
import { Dialog, Flex } from "ui";
import { useQueryClient } from "@tanstack/react-query";
import { cn } from "lib";
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
import { useProfile } from "@/lib/hooks/profile";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";

export const Chat = () => {
  const router = useRouter();
  const queryClient = useQueryClient();
  const { data: profile } = useProfile();
  const { resolvedTheme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [isFullScreen, setIsFullScreen] = useState(false);
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);

  const name = profile?.fullName.split(" ")[0] || profile?.username;

  const {
    messages,
    input,
    status,
    setInput,
    append,
    stop: handleStop,
  } = useChat({
    experimental_throttle: 100,
    onFinish: (message) => {
      message.parts?.forEach((part) => {
        if (part.type === "tool-invocation") {
          if (part.toolInvocation.toolName === "navigation") {
            if (part.toolInvocation.state === "result") {
              router.push(part.toolInvocation.result.route as string);
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
    initialMessages: [
      {
        id: "1",
        content: `Hi ${name}! I'm Maya, your AI assistant. I can help you navigate the app, change your theme, create new items, manage stories, get sprint insights, and provide project management assistance. How can I help you today?`,
        role: "assistant",
      },
    ],
  });

  const isLoading = status === "submitted";

  const handleSendMessage = (content: string) => {
    if (!content.trim()) return;
    append({
      role: "user",
      content,
    });
    setInput("");
  };

  const handleSuggestedPrompt = (prompt: string) => {
    handleSendMessage(prompt);
  };

  const handleSend = () => {
    if (!input.trim()) return;
    handleSendMessage(input);
  };

  return (
    <>
      <ChatButton
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
            "max-w-[36rem] rounded-[2rem] border border-gray-200/80 font-medium outline-none md:mb-[2.6vh] md:mt-auto",
            {
              "m-0 h-dvh w-screen max-w-[100vw] rounded-none border-0 md:mb-0 md:mt-0":
                isFullScreen,
            },
          )}
          hideClose
          overlayClassName={cn("justify-end pr-[1.5vh]", {
            "pr-0 justify-center": isFullScreen,
          })}
        >
          <Dialog.Header
            className={cn("flex h-[4.5rem] items-center px-6", {
              "absolute left-0 right-0 top-0 bg-white/40 backdrop-blur dark:bg-dark-200/30":
                isFullScreen,
            })}
          >
            <Dialog.Title className="w-full text-lg">
              <ChatHeader
                isFullScreen={isFullScreen}
                setIsFullScreen={setIsFullScreen}
                setIsOpen={setIsOpen}
              />
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="sr-only">
            Maya is your AI assistant.
          </Dialog.Description>
          <Dialog.Body
            className={cn("h-[82dvh] max-h-[82dvh] p-0", {
              "h-dvh max-h-dvh md:h-dvh md:max-h-dvh": isFullScreen,
            })}
          >
            <Flex
              className={cn("h-full pt-16", {
                "mx-auto h-dvh max-w-3xl": isFullScreen,
              })}
              direction="column"
            >
              <ChatMessages
                isLoading={isLoading}
                isStreaming={status === "streaming"}
                messages={messages}
                value={input}
              />
              {messages.length <= 1 && (
                <SuggestedPrompts onPromptSelect={handleSuggestedPrompt} />
              )}
              <ChatInput
                isLoading={isLoading}
                onChange={(e) => {
                  setInput(e.target.value);
                }}
                onSend={handleSend}
                onStop={handleStop}
                value={input}
              />
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
