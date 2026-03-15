"use client";

import { ReloadIcon } from "icons";
import { Box, Button, Flex, Text } from "ui";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useChatContext } from "@/context/chat-context";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { ChatHeader } from "./chat-header";
import { ChatInput } from "./chat-input";
import { ChatMessages } from "./chat-messages";
import { LimitReached } from "./limit-reached";
import { SuggestedPrompts } from "./suggested-prompts";

export const ChatContent = () => {
  const { chat, setIsOpen } = useChatContext();
  const { data: totalMessages = 0 } = useTotalMessages();
  const { withinLimit } = useSubscriptionFeatures();
  const needsUpgrade = !withinLimit("maxAiMessages", totalMessages);

  return (
    <>
      <NewStoryDialog
        isOpen={chat.isStoryOpen}
        setIsOpen={chat.setIsStoryOpen}
      />
      <NewObjectiveDialog
        isOpen={chat.isObjectiveOpen}
        setIsOpen={chat.setIsObjectiveOpen}
      />
      <NewSprintDialog
        isOpen={chat.isSprintOpen}
        setIsOpen={chat.setIsSprintOpen}
      />

      <Flex className="bg-background dark:bg-sidebar/80 h-full min-h-0 flex-col">
        <Box className="flex h-[3.6rem] shrink-0 items-center px-4">
          <ChatHeader
            currentChatId={chat.currentChatId}
            handleChatSelect={chat.handleChatSelect}
            handleNewChat={chat.handleNewChat}
            setIsOpen={setIsOpen}
          />
        </Box>

        <Flex className="min-h-0 flex-1" direction="column">
          <ChatMessages
            messages={chat.messages}
            onPromptSelect={chat.handleSuggestedPrompt}
            regenerate={chat.regenerate}
            status={chat.status}
            value={chat.input}
          />
          {chat.error ? (
            <Box className="mb-4 px-6">
              <Text>{chat.error.message || "An error occurred."} </Text>
              <Button
                className="mt-2"
                leftIcon={<ReloadIcon className="text-white dark:text-white" />}
                onClick={() => {
                  chat.regenerate();
                }}
              >
                Retry
              </Button>
            </Box>
          ) : null}
          {chat.messages.length === 0 ? (
            <SuggestedPrompts
              fromIndex={needsUpgrade ? 1 : 0}
              onPromptSelect={chat.handleSuggestedPrompt}
            />
          ) : null}
          {needsUpgrade ? <LimitReached isOnPage /> : null}
          <ChatInput
            attachments={chat.attachments}
            isOnPage
            messagesCount={chat.messages.length}
            onAttachmentsChange={chat.setAttachments}
            onChange={(event) => {
              chat.setInput(event.target.value);
            }}
            onSend={chat.handleSend}
            onStop={chat.handleStop}
            status={chat.status}
            value={chat.input}
          />
        </Flex>
      </Flex>
    </>
  );
};
