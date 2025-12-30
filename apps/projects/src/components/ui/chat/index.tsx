"use client";

import { useState, useEffect } from "react";
import { generateId } from "ai";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { ReloadIcon } from "icons";
import { useHotkeys } from "react-hotkeys-hook";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useMayaChat } from "@/modules/maya";
import { useUserRole } from "@/hooks/role";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";
import { LimitReached } from "./limit-reached";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

type ChatProps = {
  isOpen?: boolean;
  setIsOpen?: (open: boolean) => void;
  initialMessage?: string;
  onMessageSent?: () => void;
  isFromContext?: boolean;
};

export const Chat = ({
  isOpen: externalIsOpen,
  setIsOpen: externalSetIsOpen,
  initialMessage,
  onMessageSent,
  isFromContext,
}: ChatProps = {}) => {
  const [internalIsOpen, setInternalIsOpen] = useState(false);

  // Use external state if provided, otherwise use internal state
  const isOpen = externalIsOpen ?? internalIsOpen;
  const setIsOpen = externalSetIsOpen ?? setInternalIsOpen;

  const { userRole } = useUserRole();

  useHotkeys("shift+m", () => {
    if (userRole !== "guest" && !isFromContext) {
      setIsOpen(!isOpen);
    }
  });

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
  } = useMayaChat({
    currentChatId: generateId(),
  });
  const { data: totalMessages = 0 } = useTotalMessages();
  const { withinLimit } = useSubscriptionFeatures();
  const needsUpgrade = !withinLimit("maxAiMessages", totalMessages);

  // Handle initial message when chat opens
  useEffect(() => {
    if (initialMessage && isOpen) {
      handleSuggestedPrompt(initialMessage);
      onMessageSent?.();
    }
  }, [isOpen, initialMessage, handleSuggestedPrompt, onMessageSent]);

  return (
    <>
      <ChatButton
        isOpen={isOpen}
        onOpen={() => {
          setIsOpen(true);
        }}
      />
      <NewStoryDialog isOpen={isStoryOpen} setIsOpen={setIsStoryOpen} />
      <NewObjectiveDialog
        isOpen={isObjectiveOpen}
        setIsOpen={setIsObjectiveOpen}
      />
      <NewSprintDialog isOpen={isSprintOpen} setIsOpen={setIsSprintOpen} />
      <Dialog onOpenChange={setIsOpen} open={isOpen}>
        <Dialog.Content
          className="max-w-xl rounded-3xl border-[0.5px] border-gray-200/90 font-medium shadow-2xl outline-none backdrop-blur-lg dark:bg-dark-300/60 md:mb-[2.5dvh] md:mt-auto"
          hideClose
          overlayClassName="justify-end pr-[1.5vh] bg-dark/[0.07] dark:bg-dark/20"
        >
          <Dialog.Header className="flex h-[4.6rem] items-center px-6">
            <Dialog.Title className="w-full text-lg">
              <ChatHeader
                currentChatId={currentChatId}
                handleChatSelect={handleChatSelect}
                handleNewChat={handleNewChat}
                setIsOpen={setIsOpen}
              />
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description className="sr-only">
            Maya is your AI assistant.
          </Dialog.Description>
          <Dialog.Body className="h-208 max-h-[75dvh] p-0">
            <Flex className="h-full" direction="column">
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
                    leftIcon={
                      <ReloadIcon className="text-white dark:text-white" />
                    }
                    onClick={() => {
                      regenerate();
                    }}
                  >
                    Retry
                  </Button>
                </Box>
              ) : null}
              {messages.length === 0 && (
                <SuggestedPrompts
                  fromIndex={needsUpgrade ? 1 : 0}
                  onPromptSelect={handleSuggestedPrompt}
                />
              )}
              {needsUpgrade && <LimitReached />}
              <ChatInput
                attachments={attachments}
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
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
