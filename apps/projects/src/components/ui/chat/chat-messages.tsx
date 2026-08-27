"use client";
import { useRef, useEffect, useState } from "react";
import { Box, Flex } from "ui";
import { cn } from "lib";
import type { ChatAddToolApproveResponseFunction, ChatStatus } from "ai";
import { useProfile } from "@/lib/hooks/profile";
import type { MayaUIMessage } from "@/lib/ai/tools/types";
import { useTerminology } from "@/hooks";
import { ChatMessage } from "./chat-message";
import {
  getMessageProgressLabel,
  getVisibleMessageRenderEntries,
} from "./chat-message-utils";
import { Thinking } from "./thinking";

type ChatMessagesProps = {
  isOnPage?: boolean;
  isVoiceSpeaking?: boolean;
  isPopup?: boolean;
  messages: MayaUIMessage[];
  status: ChatStatus;
  value: string;
  regenerate: (messageId?: string) => void;
  onToolApproval: ChatAddToolApproveResponseFunction;
  onPromptSelect: (prompt: string) => void;
};

export const ChatMessages = ({
  isOnPage = false,
  isVoiceSpeaking = false,
  isPopup = false,
  messages,
  status,
  value,
  regenerate,
  onToolApproval,
  onPromptSelect,
}: ChatMessagesProps) => {
  const { getTermDisplay } = useTerminology();
  const { data: profile } = useProfile();
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const [shouldAutoScroll, setShouldAutoScroll] = useState(true);
  const isWorking = status === "submitted" || status === "streaming";
  const latestAssistantMessage = messages.findLast(
    (message) =>
      message.role === "assistant" && message.metadata?.source !== "voice",
  );
  const latestVoiceAssistantMessage = messages.findLast(
    (message) =>
      message.role === "assistant" && message.metadata?.source === "voice",
  );
  const visibleMessageEntries = getVisibleMessageRenderEntries(messages);
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
      className={cn("flex-1 overflow-y-auto", {
        "hide-scrollbar": !isOnPage,
        "px-[18px] pt-4 pb-5": isPopup,
      })}
      onScroll={handleScroll}
    >
      <Flex
        className={cn({
          "mx-auto w-full max-w-3xl px-6 py-5": isOnPage,
        })}
        direction="column"
        gap={6}
      >
        {visibleMessageEntries.map(({ isLast, message, renderKey }) => {
          return (
            <ChatMessage
              isAnimating={
                (status === "streaming" &&
                  message === latestAssistantMessage) ||
                (isVoiceSpeaking && message === latestVoiceAssistantMessage)
              }
              isLast={isLast}
              key={renderKey}
              message={message}
              onPromptSelect={onPromptSelect}
              onToolApproval={onToolApproval}
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
