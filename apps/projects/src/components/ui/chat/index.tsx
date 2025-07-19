"use client";

import { useState } from "react";
import { generateId } from "ai";
import { Box, Button, Dialog, Flex, Text } from "ui";
import { ReloadIcon } from "icons";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useMayaChat } from "@/modules/maya";
import { ChatButton } from "./chat-button";
import { ChatHeader } from "./chat-header";
import { ChatMessages } from "./chat-messages";
import { ChatInput } from "./chat-input";
import { SuggestedPrompts } from "./suggested-prompts";

export const Chat = () => {
  const [isOpen, setIsOpen] = useState(false);
  const {
    // Chat state
    messages,
    input,
    status,
    isLoading,
    error,
    attachments,
    currentChatId,

    // Chat actions
    setInput,
    handleSend,
    handleStop,
    reload,
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
          className="max-w-[36rem] rounded-[1.8rem] border-[0.5px] border-gray-200/90 font-medium outline-none backdrop-blur-lg dark:bg-dark-300/90 md:mb-[2.6dvh] md:mt-auto"
          hideClose
          overlayClassName="justify-end pr-[1.5vh]"
        >
          <Dialog.Header className="flex h-[4.5rem] items-center px-6">
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
          <Dialog.Body className="h-[80dvh] max-h-[80dvh] p-0">
            <Flex className="h-full" direction="column">
              <ChatMessages
                isLoading={isLoading}
                isStreaming={status === "streaming"}
                messages={messages}
                reload={reload}
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
                    onClick={() => reload()}
                  >
                    Retry
                  </Button>
                </Box>
              ) : null}
              {messages.length === 0 && (
                <SuggestedPrompts onPromptSelect={handleSuggestedPrompt} />
              )}
              <ChatInput
                attachments={attachments}
                isLoading={isLoading}
                onAttachmentsChange={setAttachments}
                onChange={(e) => {
                  setInput(e.target.value);
                }}
                onSend={handleSend}
                onStop={handleStop}
                value={input}
              />
            </Flex>
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
