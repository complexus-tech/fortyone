import { useRef, useEffect } from "react";
import { Box, Flex } from "ui";
import { useProfile } from "@/lib/hooks/profile";
import { ChatMessage } from "./chat-message";
import { ChatLoading } from "./chat-loading";
import type { Message } from "./types";

type ChatMessagesProps = {
  messages: Message[];
  isLoading: boolean;
};

export const ChatMessages = ({ messages, isLoading }: ChatMessagesProps) => {
  const { data: profile } = useProfile();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  return (
    <Box className="flex-1 overflow-y-auto px-6 py-6">
      <Flex direction="column" gap={6}>
        {messages.map((message) => (
          <ChatMessage key={message.id} message={message} profile={profile} />
        ))}

        {isLoading ? <ChatLoading /> : null}
        <div ref={messagesEndRef} />
      </Flex>
    </Box>
  );
};
