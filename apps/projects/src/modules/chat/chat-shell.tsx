"use client";

import type { ReactNode } from "react";
import { useCallback, useMemo, useState } from "react";
import { cn } from "lib";
import { Button, Text } from "ui";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { NativeLinkNavigationProvider } from "@/components/ui/link-preview/native-navigation";
import type { NativeLinkNavigation } from "@/components/ui/link-preview/native-navigation";
import type { NativeLinkTarget } from "@/components/ui/link-preview/native-target";
import { AttachmentThumbnailCacheProvider } from "@/components/ui/attachment-preview/thumbnail-cache-context";
import { ChatStoryPanel } from "./story-panel";
import { ThumbnailAccessGuard } from "./thumbnail-access-guard";
import { AudioAccessGuard } from "./audio-access-guard";
import { AudioPlaybackProvider } from "./audio-playback-context";
import { ConversationList } from "./conversation-list";
import { ConversationView } from "./conversation-view";
import { ThreadPanel } from "./thread-panel";
import { MessageEditDraftProvider } from "./message-edit-drafts";
import { ChatIdentityProvider } from "./chat-context";
import { useChat } from "./use-chat";
import { useChatMayaRuns } from "./use-chat-maya-runs";
import { NewMessage } from "./new-message";
import { ChatAttachmentPreviewProvider } from "./attachment-preview-context";

export const ChatShell = ({
  mayaAvatarUrl,
  navigationControl,
}: {
  mayaAvatarUrl?: string | null;
  navigationControl?: ReactNode;
}) => {
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  const {
    isComposing,
    startNewMessage,
    chooseNewMessageRecipient,
    newMessageDraft,
    updateNewMessageDraft,
    activeId,
    activeTab,
    closeThread,
    closeThreadRef,
    conversation,
    conversationHeadingRef,
    conversationListRef,
    conversations,
    conversationList,
    detail,
    draft,
    history,
    hasNewer,
    jumpToLatest,
    markVisible,
    recordVisibleMessages,
    openThread,
    openUnreadThread,
    selectConversation,
    selectTab,
    sendMessage,
    retryMessage,
    showConversation,
    showList,
    targetMessageId,
    thread,
    threadDraft,
    threadKey,
    threadTargetId,
    replies,
    repliesQuery,
    toggleReaction,
    editMessage,
    deleteMessage,
    updateDraft,
    viewMessage,
    sentMessage,
  } = useChat();
  const [openedStory, setOpenedStory] = useState<{
    target: Extract<NativeLinkTarget, { kind: "story" }>;
    trigger: HTMLElement;
    scope: string;
  } | null>(null);
  const storyScope = `${workspaceSlug}:${session?.user.id ?? "anonymous"}:${activeId}`;
  const activeStory =
    openedStory?.scope === storyScope && !isComposing ? openedStory : null;
  const openNative: NativeLinkNavigation = useCallback(
    (target, trigger) => {
      if (
        target.kind !== "story" ||
        target.workspaceSlug !== workspaceSlug.toLowerCase()
      )
        return false;
      if (thread) closeThread();
      setOpenedStory({ target, trigger, scope: storyScope });
      return true;
    },
    [closeThread, storyScope, thread, workspaceSlug],
  );
  const closeStory = () => {
    const trigger = activeStory?.trigger;
    setOpenedStory(null);
    requestAnimationFrame(() => {
      if (trigger?.isConnected) trigger.focus();
      else conversationHeadingRef.current?.focus();
    });
  };
  const maya = useChatMayaRuns(activeId, { poll: true });
  const identity = useMemo(
    () => ({
      userId: session?.user.id ?? "",
      conversationId: activeId,
      members: conversation?.members ?? [],
      mayaEnabled: conversation?.kind === "team" && maya.data?.enabled === true,
    }),
    [
      activeId,
      conversation?.kind,
      conversation?.members,
      maya.data?.enabled,
      session?.user.id,
    ],
  );
  const error =
    conversationList.error ??
    detail.error ??
    (history.data ? null : history.error);

  return (
    <NativeLinkNavigationProvider onOpen={openNative}>
      <ChatIdentityProvider value={identity}>
        <AttachmentThumbnailCacheProvider
          scope={`${workspaceSlug}:${session?.user.id ?? "anonymous"}`}
        >
          <ThumbnailAccessGuard />
          <MessageEditDraftProvider>
            <ChatAttachmentPreviewProvider
              conversationName={conversation?.name}
              key={activeId}
              onViewMessage={(id) => {
                void viewMessage(id);
              }}
            >
              <AudioPlaybackProvider
                key={`${workspaceSlug}:${session?.user.id ?? "anonymous"}:${activeId}`}
                onViewMessage={(id) => {
                  void viewMessage(id);
                }}
              >
                <AudioAccessGuard />
                <div
                  className="relative flex h-full min-h-0 overflow-hidden"
                  data-chat-active-tab={activeTab}
                >
                  <aside
                    aria-label="Conversations"
                    className={cn(
                      "border-border flex h-full w-full shrink-0 flex-col border-r-[0.5px] outline-none lg:w-[340px]",
                      showConversation ? "hidden lg:flex" : "flex",
                    )}
                    ref={conversationListRef}
                    tabIndex={-1}
                  >
                    <ConversationList
                      activeId={activeId}
                      activeThreadId={thread?.id}
                      conversations={conversations}
                      navigationControl={navigationControl}
                      onNewMessage={() => {
                        setOpenedStory(null);
                        startNewMessage();
                      }}
                      onSelect={(id) => {
                        setOpenedStory(null);
                        selectConversation(id);
                      }}
                      onSelectThread={(message, trigger) => {
                        setOpenedStory(null);
                        openUnreadThread(message, trigger);
                      }}
                    />
                  </aside>
                  {isComposing ? (
                    <NewMessage
                      draft={newMessageDraft}
                      onBack={() => {
                        setOpenedStory(null);
                        showList();
                      }}
                      onDraftChange={updateNewMessageDraft}
                      onSelect={chooseNewMessageRecipient}
                    />
                  ) : null}
                  {!isComposing && error ? (
                    <div
                      className={cn(
                        "flex min-w-0 flex-1 flex-col items-center justify-center gap-3 p-6 text-center",
                        !showConversation && "hidden lg:flex",
                      )}
                      role="alert"
                    >
                      <Text fontWeight="medium">
                        Could not load this conversation
                      </Text>
                      <Text color="muted">{error.message}</Text>
                      <Button
                        color="tertiary"
                        onClick={() => {
                          void conversationList.refetch();
                          if (activeId) {
                            void detail.refetch();
                            void history.refetch();
                          }
                        }}
                      >
                        Try again
                      </Button>
                    </div>
                  ) : null}
                  {!isComposing && !error && conversation ? (
                    <ConversationView
                      activeTab={activeTab}
                      activeThreadRootId={thread?.id}
                      conversation={conversation}
                      draft={draft}
                      fetchingHistory={history.isFetching}
                      hasNewer={hasNewer}
                      hasOlder={history.hasPreviousPage}
                      hasThread={Boolean(thread || activeStory)}
                      headingRef={conversationHeadingRef}
                      historyError={history.error}
                      key={conversation.id}
                      loadingOlder={history.isFetchingPreviousPage}
                      mayaAvatarUrl={mayaAvatarUrl}
                      onBack={() => {
                        setOpenedStory(null);
                        showList();
                      }}
                      onDeleteMessage={deleteMessage}
                      onDraftChange={(value) => {
                        updateDraft(activeId, value);
                      }}
                      onEditMessage={editMessage}
                      onJumpToLatest={jumpToLatest}
                      onLoadOlder={() =>
                        history.fetchPreviousPage({ cancelRefetch: false })
                      }
                      onReply={(message, trigger) => {
                        setOpenedStory(null);
                        openThread(message, trigger);
                      }}
                      onRetryMessage={retryMessage}
                      onSend={() => {
                        void sendMessage(activeId);
                      }}
                      onTabChange={(tab) => {
                        setOpenedStory(null);
                        selectTab(tab);
                      }}
                      onToggleReaction={toggleReaction}
                      onViewMessage={(id) => {
                        void viewMessage(id);
                      }}
                      onVisibleMessage={markVisible}
                      onVisibleMessages={recordVisibleMessages}
                      scrollToMessageId={
                        sentMessage?.conversationId === activeId &&
                        !sentMessage.threadRootId
                          ? sentMessage.id
                          : null
                      }
                      showConversation={showConversation}
                      targetMessageId={targetMessageId}
                    />
                  ) : null}
                  {!isComposing && !error && !conversation ? (
                    <div
                      className="text-text-muted hidden min-w-0 flex-1 items-center justify-center p-6 text-center lg:flex"
                      role="status"
                    >
                      {conversationList.isPending || detail.isFetching
                        ? "Loading conversations…"
                        : "Choose a conversation or start a direct message."}
                    </div>
                  ) : null}
                  {activeStory ? (
                    <ChatStoryPanel
                      onClose={closeStory}
                      target={activeStory.target}
                    />
                  ) : null}
                  {!isComposing && !activeStory && thread && conversation ? (
                    <ThreadPanel
                      allowMaya={identity.mayaEnabled}
                      closeButtonRef={closeThreadRef}
                      draft={threadDraft}
                      error={repliesQuery.error}
                      hasNewer={repliesQuery.hasNextPage}
                      hasOlder={repliesQuery.hasPreviousPage}
                      loading={repliesQuery.isFetching}
                      mayaAvatarUrl={mayaAvatarUrl}
                      onClose={closeThread}
                      onDeleteMessage={deleteMessage}
                      onDraftChange={(value) => {
                        updateDraft(threadKey, value);
                      }}
                      onEditMessage={editMessage}
                      onLoadNewer={() => {
                        void repliesQuery.fetchNextPage();
                      }}
                      onLoadOlder={() =>
                        repliesQuery.fetchPreviousPage({ cancelRefetch: false })
                      }
                      onRetry={() => {
                        void repliesQuery.refetch();
                      }}
                      onRetryMessage={retryMessage}
                      onSend={() => {
                        void sendMessage(threadKey, thread.id);
                      }}
                      onToggleReaction={toggleReaction}
                      onVisibleMessage={markVisible}
                      onVisibleMessages={recordVisibleMessages}
                      replies={replies}
                      scrollToMessageId={
                        sentMessage?.threadRootId === thread.id
                          ? sentMessage.id
                          : null
                      }
                      targetMessageId={
                        threadTargetId === sentMessage?.id
                          ? null
                          : threadTargetId
                      }
                      thread={thread}
                    />
                  ) : null}
                </div>
              </AudioPlaybackProvider>
            </ChatAttachmentPreviewProvider>
          </MessageEditDraftProvider>
        </AttachmentThumbnailCacheProvider>
      </ChatIdentityProvider>
    </NativeLinkNavigationProvider>
  );
};
