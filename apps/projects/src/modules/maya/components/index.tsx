import { Box, Button, Text } from "ui";
import { ReloadIcon } from "icons";
import { NewStoryDialog, NewObjectiveDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { ChatMessages } from "@/components/ui/chat/chat-messages";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { SuggestedPrompts } from "@/components/ui/chat/suggested-prompts";
import { BodyContainer } from "@/components/shared";
import type { MayaChatConfig } from "../types";
import { useMayaChat } from "../hooks/use-maya-chat";
import { Header } from "./header";

type MayaChatProps = {
  config: MayaChatConfig;
};

export const MayaChat = ({ config }: MayaChatProps) => {
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
          isLoading={isLoading}
          isOnPage
          isStreaming={status === "streaming"}
          messages={messages}
          onPromptSelect={handleSuggestedPrompt}
          reload={reload}
          value={input}
        />
        {error ? (
          <Box className="mb-4 px-6">
            <Text>{error.message || "An error occurred."} </Text>
            <Button
              className="mt-2"
              leftIcon={<ReloadIcon className="text-white dark:text-white" />}
              onClick={() => reload()}
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
          isLoading={isLoading}
          isOnPage
          messagesCount={messages.length}
          onAttachmentsChange={setAttachments}
          onChange={(e) => {
            setInput(e.target.value);
          }}
          onSend={handleSend}
          onStop={handleStop}
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
