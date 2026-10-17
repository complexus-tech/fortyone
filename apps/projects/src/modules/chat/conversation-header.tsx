import type { Ref } from "react";
import { ArrowLeftIcon, ChatIcon, LinkIcon } from "icons";
import { cn } from "lib";
import { Button, Tabs } from "ui";
import { ConversationMark, PersonAvatar } from "./avatars";
import { useChatIdentity } from "./chat-context";
import { ICON_BUTTON_CLASS } from "./chat-styles";
import type { ConversationFeed } from "./types";

export const ConversationHeader = ({
  conversation,
  headingRef,
  mayaAvatarUrl,
  onBack,
}: {
  conversation: ConversationFeed;
  headingRef: Ref<HTMLHeadingElement>;
  mayaAvatarUrl?: string | null;
  onBack: () => void;
}) => {
  const { userId } = useChatIdentity();
  const recipient = conversation.members.find((person) => person.id !== userId);

  return (
    <div className="flex min-w-0 items-center justify-between gap-3">
      <div className="flex min-w-0 flex-1 items-center gap-3">
        <Button
          aria-label="Back to conversations"
          asIcon
          className={cn(ICON_BUTTON_CLASS, "lg:hidden")}
          color="tertiary"
          onClick={onBack}
          size="sm"
          variant="naked"
        >
          <ArrowLeftIcon className="size-4" />
        </Button>
        {conversation.kind === "dm" && recipient ? (
          <PersonAvatar mayaAvatarUrl={mayaAvatarUrl} person={recipient} />
        ) : (
          <ConversationMark conversation={conversation} />
        )}
        <h2
          className="truncate font-semibold outline-none"
          ref={headingRef}
          tabIndex={-1}
          title={conversation.name}
        >
          {conversation.name}
        </h2>
      </div>
      <div className="flex shrink-0 items-center gap-4">
        {conversation.kind === "team" ? (
          <div
            aria-label="Team members"
            className="isolate hidden items-center -space-x-1.5 @min-[700px]/conversation:flex"
          >
            {conversation.members.slice(0, 4).map((person) => (
              <span
                className="bg-surface ring-surface relative inline-flex rounded-full ring-2"
                key={person.id}
              >
                <PersonAvatar mayaAvatarUrl={mayaAvatarUrl} person={person} />
              </span>
            ))}
            <span className="text-text-muted ml-3 pl-2">
              {conversation.memberCount}
            </span>
          </div>
        ) : null}
        <Tabs.List
          aria-label="Conversation content"
          className="mx-0 flex-nowrap md:mx-0"
        >
          <Tabs.Tab
            aria-label="Messages"
            leftIcon={<ChatIcon className="h-[1.1rem] w-auto" />}
            value="messages"
          >
            <span className="hidden @min-[500px]/conversation:inline">
              Messages
            </span>
          </Tabs.Tab>
          <Tabs.Tab
            aria-label="Files & links"
            leftIcon={<LinkIcon className="h-[1.1rem] w-auto" />}
            value="files"
          >
            <span className="hidden @min-[500px]/conversation:inline">
              Files & links
            </span>
          </Tabs.Tab>
        </Tabs.List>
      </div>
    </div>
  );
};
