"use client";
import { useRef, useEffect, useState } from "react";
import { Box, Flex } from "ui";
import type { ChatStatus } from "ai";
import { useProfile } from "@/lib/hooks/profile";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { useTerminology } from "@/hooks";
import {
  ChatMessage,
  getMessageProgressLabel,
  hasVisibleMessageContent,
} from "./chat-message";
import { Thinking } from "./thinking";

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
  const { getTermDisplay } = useTerminology();
  const { data: profile } = useProfile();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const [shouldAutoScroll, setShouldAutoScroll] = useState(true);
  const isWorking = status === "submitted" || status === "streaming";
  const latestAssistantMessage = messages.findLast(
    (message) => message.role === "assistant",
  );
  const latestAssistantMessageId = latestAssistantMessage?.id;
  const visibleMessages = messages.filter((message) =>
    hasVisibleMessageContent(
      message,
      isWorking && message.id === latestAssistantMessageId,
    ),
  );
  let rawProgressLabel: string | undefined;
  if (isWorking) {
    rawProgressLabel = latestAssistantMessage
      ? getMessageProgressLabel(latestAssistantMessage)
      : "Thinking";
  }
  const progressLabel = rawProgressLabel
    ?.replaceAll("stories", getTermDisplay("storyTerm", { variant: "plural" }))
    .replaceAll("story", getTermDisplay("storyTerm"));

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  // Handle scroll position detection
  const handleScroll = (e: React.UIEvent<HTMLDivElement>) => {
    const { scrollTop, scrollHeight, clientHeight } = e.currentTarget;
    const distanceFromBottom = scrollHeight - scrollTop - clientHeight;
    const threshold = 100; // 100px threshold

    const isNearBottom = distanceFromBottom < threshold;
    setShouldAutoScroll(isNearBottom);
  };

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
        {visibleMessages.map((message) => {
          const deferToolOutputs =
            isWorking && message.id === latestAssistantMessageId;

          return (
            <ChatMessage
              deferToolOutputs={deferToolOutputs}
              isLast={message.id === messages.at(-1)?.id}
              key={message.id}
              message={message}
              onPromptSelect={onPromptSelect}
              profile={profile}
              regenerate={regenerate}
              status={status}
            />
          );
        })}
        {progressLabel ? <Thinking message={progressLabel} /> : null}
        {status === "streaming" ? <div className="h-32" /> : null}
        <div ref={messagesEndRef} />
      </Flex>
    </Box>
  );
};
