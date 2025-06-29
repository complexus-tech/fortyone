import { useRef, useEffect, useState, useCallback } from "react";
import { Box, Flex } from "ui";
import type { Message } from "@ai-sdk/react";
import { useProfile } from "@/lib/hooks/profile";
import { ChatMessage } from "./chat-message";
import { ChatLoading } from "./chat-loading";

type ChatMessagesProps = {
  messages: Message[];
  isLoading: boolean;
  isStreaming?: boolean;
  value: string;
};

export const ChatMessages = ({
  messages,
  isLoading,
  isStreaming,
  value,
}: ChatMessagesProps) => {
  const { data: profile } = useProfile();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Smart scroll state management
  const [prevMessageCount, setPrevMessageCount] = useState(0);
  const [shouldAutoScroll, setShouldAutoScroll] = useState(true);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  // Handle scroll position detection
  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight;
    const threshold = 100; // 100px threshold

    const isNearBottom = distanceFromBottom < threshold;
    setShouldAutoScroll(isNearBottom);
  }, []);

  // Only scroll on new messages when user is near bottom
  useEffect(() => {
    const hasNewMessage = messages.length > prevMessageCount;
    if (hasNewMessage && shouldAutoScroll) {
      scrollToBottom();
    }
    setPrevMessageCount(messages.length);
  }, [messages.length, shouldAutoScroll, prevMessageCount]);

  useEffect(() => {
    if (value === "") {
      scrollToBottom();
    }
  }, [value]);

  return (
    <Box className="flex-1 overflow-y-auto px-6 py-6" onScroll={handleScroll}>
      <Flex direction="column" gap={6}>
        {messages.map((message) => (
          <ChatMessage
            isStreaming={isStreaming}
            key={message.id}
            message={message}
            profile={profile}
          />
        ))}
        {isLoading ? <ChatLoading /> : null}
        {isStreaming ? <div className="h-32" /> : null}
        <div ref={messagesEndRef} />
      </Flex>
    </Box>
  );
};
