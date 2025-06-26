"use client";

import { useState } from "react";
import { Dialog, Flex } from "ui";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";
import type { Message } from "./types";

export const Chat = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<Message[]>([
    {
      id: "1",
      content:
        "Hi! I'm Maya, your AI assistant. I can help you with project insights, sprint planning, story management, and more. How can I assist you today?",
      sender: "ai",
      timestamp: new Date(),
    },
  ]);
  const [inputValue, setInputValue] = useState("");
  const [isLoading, setIsLoading] = useState(false);

  const handleSendMessage = (content: string) => {
    if (!content.trim()) return;

    const newMessage: Message = {
      id: Date.now().toString(),
      content,
      sender: "user",
      timestamp: new Date(),
    };

    setMessages((prev) => [...prev, newMessage]);
    setInputValue("");
    setIsLoading(true);

    // Simulate AI response (replace with actual API call)
    setTimeout(() => {
      const aiResponse: Message = {
        id: (Date.now() + 1).toString(),
        content:
          "I understand your request. While I'm currently in development, I'll soon be able to provide detailed insights about your projects, analyze sprint data, and help you make data-driven decisions. Is there anything specific you'd like to know about your current sprint?",
        sender: "ai",
        timestamp: new Date(),
      };
      setMessages((prev) => [...prev, aiResponse]);
      setIsLoading(false);
    }, 1000);
  };

  const handleSuggestedPrompt = (prompt: string) => {
    handleSendMessage(prompt);
  };

  const handleSend = () => {
    handleSendMessage(inputValue);
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
          className="max-w-[37rem] rounded-[2rem] antialiased md:mb-[2.6vh] md:mt-auto"
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Header className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
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
                onChange={setInputValue}
                onSend={handleSend}
                value={inputValue}
              />
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
