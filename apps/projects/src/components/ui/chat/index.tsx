"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useChat } from "@ai-sdk/react";
import { Dialog, Flex } from "ui";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";

export const Chat = () => {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);

  const { messages, input, handleInputChange, handleSubmit, status } = useChat({
    onFinish: (message) => {
      message.parts?.forEach((part) => {
        if (part.type === "tool-invocation") {
          // eslint-disable-next-line no-console -- debug
          console.log(part.toolInvocation);

          if (part.toolInvocation.toolName === "navigation") {
            if (part.toolInvocation.state === "result") {
              router.push(part.toolInvocation.result.route as string);
            }
          }
        }
      });
    },
    initialMessages: [
      {
        id: "1",
        content:
          "Hi! I'm Maya, your AI assistant. I can help you navigate the app, manage stories, get sprint insights, and provide project management assistance. How can I help you today?",
        role: "assistant",
      },
    ],
  });

  const isLoading = status === "submitted";

  const handleSendMessage = (content: string) => {
    if (!content.trim()) return;
    handleSubmit(new Event("submit") as any);
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
          className="max-w-[36rem] rounded-[2rem] font-medium dark:bg-dark-300 md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
          tabIndex={0}
        >
          <Dialog.Header className="border-b-[0.5px] border-gray-100 px-6 py-5 dark:border-dark-100">
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
                onChange={handleInputChange}
                onSend={handleSend}
                value={input}
              />
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
