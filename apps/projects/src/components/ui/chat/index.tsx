"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useTheme } from "next-themes";
import { useChat } from "@ai-sdk/react";
import { Dialog, Flex } from "ui";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";

export const Chat = () => {
  const router = useRouter();
  const { resolvedTheme, setTheme } = useTheme();
  const [isOpen, setIsOpen] = useState(false);
  const [isStoryOpen, setIsStoryOpen] = useState(false);
  const [isObjectiveOpen, setIsObjectiveOpen] = useState(false);
  const [isSprintOpen, setIsSprintOpen] = useState(false);

  const { messages, input, status, setInput, append } = useChat({
    onFinish: (message) => {
      message.parts?.forEach((part) => {
        if (part.type === "tool-invocation") {
          if (part.toolInvocation.toolName === "navigation") {
            if (part.toolInvocation.state === "result") {
              router.push(part.toolInvocation.result.route as string);
            }
          }
          if (part.toolInvocation.toolName === "theme") {
            if (part.toolInvocation.state === "result") {
              const requestedTheme = part.toolInvocation.result.theme as string;
              if (requestedTheme === "toggle") {
                const newTheme = resolvedTheme === "dark" ? "light" : "dark";
                setTheme(newTheme);
              } else {
                setTheme(requestedTheme);
              }
            }
          }
          if (part.toolInvocation.toolName === "quickCreate") {
            if (part.toolInvocation.state === "result") {
              const action = part.toolInvocation.result.action as string;
              switch (action) {
                case "story":
                  setIsStoryOpen(true);
                  break;
                case "objective":
                  setIsObjectiveOpen(true);
                  break;
                case "sprint":
                  setIsSprintOpen(true);
                  break;
              }
            }
          }
        }
      });
    },
    initialMessages: [
      {
        id: "1",
        content:
          "Hi! I'm Maya, your AI assistant. I can help you navigate the app, change your theme, create new items, manage stories, get sprint insights, and provide project management assistance. How can I help you today?",
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
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="max-w-[36rem] rounded-[2rem] bg-white/85 font-medium backdrop-blur dark:bg-dark-300/80 md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Header className="px-6 py-5">
            <Dialog.Title className="text-lg">
              <ChatHeader />
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="h-[80dvh] max-h-[80dvh] p-0">
            <Flex className="h-full" direction="column">
              <ChatMessages isLoading={isLoading} messages={messages} />
              {messages.length === 1 && (
                <SuggestedPrompts onPromptSelect={handleSuggestedPrompt} />
              )}
              <ChatInput
                isLoading={isLoading}
                onChange={(e) => {
                  setInput(e.target.value);
                }}
                onSend={handleSend}
                value={input}
              />
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
      />
      <NewSprintDialog isOpen={isSprintOpen} setIsOpen={setIsSprintOpen} />
    </>
  );
};
