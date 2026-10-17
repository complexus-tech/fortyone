import { Avatar } from "ui";
import { MayaAvatar } from "@/components/ui/maya-avatar";
import { useChatIdentity } from "./chat-context";
import type { ChatConversation, ChatPerson } from "./types";

export const PersonAvatar = ({
  person,
  mayaAvatarUrl,
  small = false,
}: {
  person: ChatPerson;
  mayaAvatarUrl?: string | null;
  small?: boolean;
}) => (
  <span
    aria-label={person.fullName}
    className="inline-flex shrink-0 self-start"
    title={person.fullName}
  >
    {person.isSystem ? (
      <MayaAvatar
        size={small ? "xs" : "md"}
        src={person.avatarUrl ?? mayaAvatarUrl}
      />
    ) : (
      <Avatar
        name={person.fullName}
        size={small ? "xs" : "md"}
        src={person.avatarUrl}
      />
    )}
  </span>
);

export const ConversationMark = ({
  conversation,
  active = false,
}: {
  conversation: ChatConversation;
  active?: boolean;
}) => {
  const { userId } = useChatIdentity();
  return conversation.kind === "dm" ? (
    <Avatar
      name={conversation.name}
      size="xs"
      src={
        conversation.members.find((member) => member.id !== userId)?.avatarUrl
      }
    />
  ) : (
    <span
      aria-hidden
      className="text-text-muted flex size-5 shrink-0 items-center justify-center text-[19px] leading-none font-normal"
      style={{
        color: active
          ? "var(--color-primary)"
          : conversation.color ?? undefined,
      }}
    >
      #
    </span>
  );
};
