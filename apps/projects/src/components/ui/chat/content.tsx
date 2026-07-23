"use client";

import { ReloadIcon } from "icons";
import { cn } from "lib";
import { Box, Button, Flex, Text } from "ui";
import { NewObjectiveDialog, NewStoryDialog } from "@/components/ui";
import { NewSprintDialog } from "@/components/ui/new-sprint-dialog";
import { useChatContext } from "@/context/chat-context";
import { useSession } from "@/lib/auth/client";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useTotalMessages } from "@/modules/ai-chats/hooks/use-total-messages";
import { shouldShowMayaMessageLimit } from "@/modules/maya/utils/message-limit";
import { ChatHeader } from "./chat-header";
import { ChatInput } from "./chat-input";
import { ChatMessages } from "./chat-messages";
import { LimitReached } from "./limit-reached";
import { SuggestedPrompts } from "./suggested-prompts";

export const ChatContent = ({ isPopup = false }: { isPopup?: boolean }) => {
  const { chat, setIsOpen } = useChatContext();
  const { data: totalMessages = 0 } = useTotalMessages();
  const { data: session } = useSession();
  const { getLimit } = useSubscriptionFeatures();
  const isInternalUser = session?.user.isInternal === true;
  const needsUpgrade = shouldShowMayaMessageLimit({
    isInternalUser,
    limit: getLimit("maxAiMessages"),
    totalMessages,
  });
  const isWorking = chat.status === "submitted" || chat.status === "streaming";

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

      <Flex
        className={cn("h-full min-h-0 flex-col", {
          "bg-transparent": isPopup,
          "bg-background dark:bg-sidebar/40": !isPopup,
        })}
      >
        <Box
          className={cn("flex shrink-0 items-center", {
            "h-[4.75rem] border-b border-black/[0.06] px-5 py-3 dark:border-white/[0.06]":
              isPopup,
            "h-[3.6rem] px-4": !isPopup,
          })}
        >
          <ChatHeader
            currentChatId={chat.currentChatId}
            handleChatSelect={chat.handleChatSelect}
            handleNewChat={chat.handleNewChat}
            hasMessages={chat.displayMessages.length > 0}
            isPopup={isPopup}
            setIsOpen={setIsOpen}
          />
        </Box>

        <Flex className="min-h-0 flex-1" direction="column">
          {chat.displayMessages.length > 0 || isWorking ? (
            <ChatMessages
              isPopup={isPopup}
              isVoiceSpeaking={chat.realtimeVoice.isSpeaking}
              messages={chat.displayMessages}
              onPromptSelect={chat.handleSuggestedPrompt}
              regenerate={chat.regenerate}
              status={chat.status}
              value={chat.input}
            />
          ) : null}
          {chat.error || chat.realtimeVoice.error ? (
            <Box className="mb-4 px-6">
              <Text>
                {chat.realtimeVoice.error ||
                  chat.error?.message ||
                  "An error occurred."}{" "}
              </Text>
              <Button
                className="mt-2"
                leftIcon={<ReloadIcon className="text-current" />}
                onClick={() => {
                  chat.regenerate();
                }}
              >
                Retry
              </Button>
            </Box>
          ) : null}
          {chat.displayMessages.length === 0 && !isWorking ? (
            <Box className="min-h-0 flex-1 overflow-y-auto">
              <SuggestedPrompts
                fromIndex={needsUpgrade ? 1 : 0}
                isPopup={isPopup}
                onPromptSelect={chat.handleSuggestedPrompt}
              />
            </Box>
          ) : null}
          {needsUpgrade ? <LimitReached isOnPage={!isPopup} /> : null}
          <ChatInput
            attachments={chat.attachments}
            isOnPage={!isPopup}
            isPopup={isPopup}
            liveVoiceDisabled={needsUpgrade}
            messagesCount={chat.displayMessages.length}
            onAttachmentsChange={chat.setAttachments}
            onChange={(event) => {
              chat.setInput(event.target.value);
            }}
            onSend={chat.handleSend}
            onStop={chat.handleStop}
            realtimeVoice={chat.realtimeVoice}
            status={chat.status}
            value={chat.input}
          />
        </Flex>
      </Flex>
    </>
  );
};
