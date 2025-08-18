"use client";
import { Box, Button, Text } from "ui";
import { ReloadIcon } from "icons";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { ChatMessages } from "@/components/ui/chat/chat-messages";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { SuggestedPrompts } from "@/components/ui/chat/suggested-prompts";
import { BodyContainer } from "@/components/shared";
import { useMayaChat } from "../hooks/use-maya-chat";
import { useMayaNavigation } from "../hooks/use-maya-navigation";
import type { MayaChatConfig } from "../types";
import { Header } from "./header";

export const MayaChat = () => {
  const { chatRef, getInitialChatId, isNewChat } = useMayaNavigation();
  const config: MayaChatConfig = {
    currentChatId: getInitialChatId(),
    hasSelectedChat: Boolean(chatRef),
    isNewChat: isNewChat(),
  };
  const {
    // Chat state
    messages,
    input,
    status,
    error,
    attachments,
    currentChatId,

    // Chat actions
    setInput,
    handleSend,
    handleStop,
    regenerate,
    handleNewChat,
    handleChatSelect,
    handleSuggestedPrompt,
    setAttachments,

    // Dialog states
    isStoryOpen,
    setIsStoryOpen,
    isObjectiveOpen,
    setIsObjectiveOpen,
    isSprintOpen,
    setIsSprintOpen,
  } = useMayaChat(config);

  return (
    <>
      <Header
        currentChatId={currentChatId}
        handleChatSelect={handleChatSelect}
        handleNewChat={handleNewChat}
      />
      <BodyContainer className="mx-auto flex max-w-4xl flex-col">
        <ChatMessages
          messages={messages}
          onPromptSelect={handleSuggestedPrompt}
          regenerate={regenerate}
          status={status}
          value={input}
        />
        {error ? (
          <Box className="mb-4 px-6">
            <Text>{error.message || "An error occurred."} </Text>
            <Button
              className="mt-2"
              leftIcon={<ReloadIcon className="text-white dark:text-white" />}
              onClick={() => {
                regenerate();
              }}
            >
              Retry
            </Button>
          </Box>
        ) : null}
        {messages.length === 0 ? (
          <SuggestedPrompts isOnPage onPromptSelect={handleSuggestedPrompt} />
        ) : null}
        <ChatInput
          attachments={attachments}
          isOnPage
          messagesCount={messages.length}
          onAttachmentsChange={setAttachments}
          onChange={(e) => {
            setInput(e.target.value);
          }}
          onSend={handleSend}
          onStop={handleStop}
          status={status}
          value={input}
        />
      </BodyContainer>

      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
      />
      <NewSprintDialog isOpen={isSprintOpen} setIsOpen={setIsSprintOpen} />
    </>
  );
};
