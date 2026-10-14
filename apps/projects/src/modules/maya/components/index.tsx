"use client";
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
import { useChatRuntime } from "@/context/chat-context";
import { useProfile } from "@/lib/hooks/profile";
import { useMayaMessageAvailability } from "@/shared/maya-usage/use-maya-message-availability";
import { Header } from "./header";
import { RecentWork } from "./recent-work";
import styles from "./index.module.css";

export const MayaChat = () => {
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
  } = useChatRuntime();

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
      onValueChange={setInput}
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
      {!isEmptyState ? (
        <Header
          currentChatId={currentChatId}
          handleChatSelect={handleChatSelect}
          handleNewChat={handleNewChat}
        />
      ) : null}
      <BodyContainer className="flex h-auto min-h-0 flex-1 flex-col overflow-hidden">
        {isEmptyState ? (
          <Box className="relative flex min-h-0 flex-1 overflow-y-auto">
            <Box className="mx-auto my-auto flex w-full max-w-5xl flex-col px-5 pt-20 pb-10 sm:px-8 md:pt-28 md:pb-14">
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
                <RecentWork />
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
