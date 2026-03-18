"use client";
import { useRef, useEffect, useState, useCallback } from "react";
import { Box, Flex } from "ui";
import type { ChatStatus } from "ai";
import { useProfile } from "@/lib/hooks/profile";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { ChatMessage } from "./chat-message";
import { ChatLoading } from "./chat-loading";

type ChatMessagesProps = {
  messages: MayaUIMessage[];
  status: ChatStatus;
  value: string;
  regenerate: (messageId?: string) => void;
  onPromptSelect: (prompt: string) => void;
};

export const ChatMessages = ({
  messages,
  status,
  value,
  regenerate,
  onPromptSelect,
}: ChatMessagesProps) => {
  const { data: profile } = useProfile();
  const messagesEndRef = useRef<HTMLDivElement>(null);

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

  useEffect(() => {
    if (messages.length > 0 && shouldAutoScroll) {
      scrollToBottom();
    }
  }, [messages, shouldAutoScroll]);

  useEffect(() => {
    if (value === "") {
      scrollToBottom();
    }
  }, [value]);

  return (
    <Box
      className="hide-scrollbar flex-1 overflow-y-auto p-6"
      onScroll={handleScroll}
    >
      <Flex direction="column" gap={6}>
        {messages.map((message, idx) => (
          <ChatMessage
            isLast={idx === messages.length - 1}
            key={message.id}
            message={message}
            onPromptSelect={onPromptSelect}
            profile={profile}
            regenerate={regenerate}
            status={status}
          />
        ))}
        {/* Show loading only before an assistant message exists */}
        {status !== "ready" &&
          messages[messages.length - 1]?.role !== "assistant" && (
            <ChatLoading />
          )}
        {status === "streaming" ? <div className="h-32" /> : null}
        <div ref={messagesEndRef} />
      </Flex>
    </Box>
  );
};
