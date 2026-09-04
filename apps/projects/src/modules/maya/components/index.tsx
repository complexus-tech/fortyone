"use client";
import { useCallback } from "react";
import { Box, Button, Text } from "ui";
import { ReloadIcon } from "icons";
import { cn } from "lib";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { ChatMessages } from "@/components/ui/chat/chat-messages";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { SuggestedPrompts } from "@/components/ui/chat/suggested-prompts";
import { LimitReached } from "@/components/ui/chat/limit-reached";
import { BodyContainer } from "@/components/shared";
import { useWalkthrough } from "@/components/walkthrough/walkthrough-provider";
import { useProfile } from "@/lib/hooks/profile";
import { useMayaChat } from "../hooks/use-maya-chat";
import { useMayaNavigation } from "../hooks/use-maya-navigation";
import { useMayaMessageAvailability } from "../hooks/use-maya-message-availability";
import type { MayaChatConfig } from "../types";
import { Header } from "./header";
import styles from "./index.module.css";

export const MayaChat = () => {
  const { chatRef, getInitialChatId, updateChatRef, clearChatRef } =
    useMayaNavigation();
  const { completeWalkthroughAction } = useWalkthrough();
  const handleUserMessageSubmitted = useCallback(() => {
    completeWalkthroughAction("maya-message-completed");
  }, [completeWalkthroughAction]);
  const config: MayaChatConfig = {
    currentChatId: getInitialChatId(),
    hasSelectedChat: Boolean(chatRef),
    onUserMessageSubmitted: handleUserMessageSubmitted,
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
    googleDriveFiles,
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
    removeGoogleDriveFile,

    // Dialog states
    isStoryOpen,
    setIsStoryOpen,
    isObjectiveOpen,
    setIsObjectiveOpen,
    isSprintOpen,
    setIsSprintOpen,
  } = useMayaChat(config);

  const { isLimited: needsUpgrade } = useMayaMessageAvailability();
  const { data: profile } = useProfile();
  const isEmptyState = displayMessages.length === 0;
  const firstName =
    profile?.fullName.trim().split(/\s+/)[0] || profile?.username || "there";

  const errorNotice =
    error || realtimeVoice.error ? (
      <Box
        className={cn("mb-4", {
          "px-1": isEmptyState,
          "px-6": !isEmptyState,
        })}
      >
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
    ) : null;

  const composer = (
    <ChatInput
      attachments={attachments}
      googleDriveFiles={googleDriveFiles}
      isEmptyState={isEmptyState}
      isOnPage
      liveVoiceDisabled={needsUpgrade}
      messagesCount={displayMessages.length}
      onAttachmentsChange={setAttachments}
      onChange={(e) => {
        setInput(e.target.value);
      }}
      onGoogleDriveFileRemove={removeGoogleDriveFile}
      onSend={handleSend}
      onStop={handleStop}
      realtimeVoice={realtimeVoice}
      status={status}
      value={input}
    />
  );

  return (
    <Box
      className={cn(
        "relative flex h-full min-h-0 flex-col overflow-hidden",
        styles.page,
      )}
    >
      <Header
        currentChatId={currentChatId}
        handleChatSelect={handleChatSelect}
        handleNewChat={handleNewChat}
      />
      <BodyContainer className="flex h-auto min-h-0 flex-1 flex-col overflow-hidden">
        {isEmptyState ? (
          <Box className="relative flex min-h-0 flex-1 overflow-y-auto">
            <Box className="mx-auto flex min-h-full w-full max-w-5xl flex-col justify-center px-5 py-10 sm:px-8 md:py-14">
              <Box className="mb-7 flex flex-col items-center text-center md:mb-9">
                <Text
                  as="h1"
                  className="text-4xl font-semibold tracking-[-0.045em] md:text-[3rem]"
                >
                  Hi, {firstName}! Ask Maya anything.
                </Text>
                <Text
                  className="mt-4 max-w-lg text-base leading-6 md:text-lg"
                  color="muted"
                >
                  Plan what&apos;s next, find what matters, or move work
                  forward.
                </Text>
              </Box>

              <Box className="mx-auto w-full max-w-4xl">
                {errorNotice}
                {needsUpgrade ? <LimitReached isOnPage /> : null}
                {composer}
                <SuggestedPrompts
                  isOnPage
                  onPromptSelect={handleSuggestedPrompt}
                />
              </Box>
            </Box>
          </Box>
        ) : (
          <>
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
              {errorNotice}
              {needsUpgrade ? <LimitReached isOnPage /> : null}
              {composer}
            </Box>
          </>
        )}
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
