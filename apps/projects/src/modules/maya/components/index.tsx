"use client";
import { useCallback } from "react";
import { Box, Button, Text } from "ui";
import { ReloadIcon } from "icons";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { ChatMessages } from "@/components/ui/chat/chat-messages";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { SuggestedPrompts } from "@/components/ui/chat/suggested-prompts";
import { LimitReached } from "@/components/ui/chat/limit-reached";
import { BodyContainer } from "@/components/shared";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useMayaChat } from "../hooks/use-maya-chat";
import { useMayaNavigation } from "../hooks/use-maya-navigation";
import { useMayaMessageAvailability } from "../hooks/use-maya-message-availability";
import type { MayaChatConfig } from "../types";
import { Header } from "./header";

export const MayaChat = () => {
  const { chatRef, getInitialChatId, isNewChat, updateChatRef, clearChatRef } =
    useMayaNavigation();
  const { completeWalkthroughAction } = useWalkthrough();
  const handleUserMessageCompleted = useCallback(() => {
    completeWalkthroughAction("maya-message-completed");
  }, [completeWalkthroughAction]);
  const config: MayaChatConfig = {
    currentChatId: getInitialChatId(),
    hasSelectedChat: Boolean(chatRef),
    isNewChat: isNewChat(),
    onUserMessageCompleted: handleUserMessageCompleted,
    updateChatRef,
    clearChatRef,
  };
  const {
    // Chat state
    displayMessages,
    input,
    status,
    error,
    attachments,
    currentChatId,
    realtimeVoice,

    // Chat actions
    setInput,
    handleSend,
    handleStop,
    regenerate,
    handleNewChat,
    handleChatSelect,
    handleSuggestedPrompt,
    addToolApprovalResponse,
    setAttachments,

    // Dialog states
    isStoryOpen,
    setIsStoryOpen,
    isObjectiveOpen,
    setIsObjectiveOpen,
    isSprintOpen,
    setIsSprintOpen,
  } = useMayaChat(config);

  const { isLimited: needsUpgrade } = useMayaMessageAvailability();

  return (
    <Box className="flex h-full min-h-0 flex-col overflow-hidden">
      <Header
        currentChatId={currentChatId}
        handleChatSelect={handleChatSelect}
        handleNewChat={handleNewChat}
      />
      <BodyContainer className="flex h-auto min-h-0 flex-1 flex-col overflow-hidden">
        <ChatMessages
          isOnPage
          isVoiceSpeaking={realtimeVoice.isSpeaking}
          messages={displayMessages}
          onPromptSelect={handleSuggestedPrompt}
          onToolApproval={addToolApprovalResponse}
          regenerate={regenerate}
          status={status}
          value={input}
        />
        <Box className="mx-auto flex w-full max-w-3xl shrink-0 flex-col">
          {error || realtimeVoice.error ? (
            <Box className="mb-4 px-6">
              <Text>
                {realtimeVoice.error || error?.message || "An error occurred."}{" "}
              </Text>
              <Button
                className="mt-2"
                leftIcon={<ReloadIcon className="text-current" />}
                onClick={() => {
                  regenerate();
                }}
              >
                Retry
              </Button>
            </Box>
          ) : null}

          {displayMessages.length === 0 ? (
            <SuggestedPrompts isOnPage onPromptSelect={handleSuggestedPrompt} />
          ) : null}
          {needsUpgrade ? <LimitReached isOnPage /> : null}
          <ChatInput
            attachments={attachments}
            isOnPage
            liveVoiceDisabled={needsUpgrade}
            messagesCount={displayMessages.length}
            onAttachmentsChange={setAttachments}
            onChange={(e) => {
              setInput(e.target.value);
            }}
            onSend={handleSend}
            onStop={handleStop}
            realtimeVoice={realtimeVoice}
            status={status}
            value={input}
          />
        </Box>
      </BodyContainer>

      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
      />
      <NewSprintDialog isOpen={isSprintOpen} setIsOpen={setIsSprintOpen} />
    </Box>
  );
};
