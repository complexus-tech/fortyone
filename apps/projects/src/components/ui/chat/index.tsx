"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useChat } from "ai/react";
import { Dialog, Flex } from "ui";
import Markdown from "react-markdown";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";
import type { Message } from "./types";

export const Chat = () => {
  const router = useRouter();
  const [isOpen, setIsOpen] = useState(false);

  const { messages, input, handleInputChange, handleSubmit, isLoading } =
    useChat({
      api: "/api/chat",
      onFinish: (message) => {
        // Check if the AI used the navigation tool
        if (message.toolInvocations) {
          for (const toolInvocation of message.toolInvocations) {
            if (
              toolInvocation.toolName === "navigate" &&
              toolInvocation.result?.route
            ) {
              // Navigate to the specified route
              setTimeout(() => {
                router.push(toolInvocation.result.route);
                setIsOpen(false);
              }, 2000); // Give user 2 seconds to read the response
              break;
            }
          }
        }
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

  // Convert AI SDK messages to our Message format
  const convertedMessages: Message[] = messages.map((msg) => ({
    id: msg.id,
    content: msg.content,
    sender: msg.role === "user" ? "user" : "ai",
    timestamp: new Date(msg.createdAt || Date.now()),
  }));

  return (
    <>
      <ChatButton
        onOpen={() => {
          setIsOpen(true);
        }}
      />
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="max-w-[37rem] rounded-[2rem] font-normal md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Header className="border-b border-gray-100 px-6 py-5 dark:border-dark-100">
            <Dialog.Title className="text-lg">
              <ChatHeader />
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Body className="h-[80dvh] max-h-[80dvh] p-0">
            <Flex className="h-full" direction="column">
              <ChatMessages
                isLoading={isLoading}
                messages={convertedMessages}
              />
              {convertedMessages.length === 1 && (
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
