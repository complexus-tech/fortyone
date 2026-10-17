import { useCurrentWorkspace } from "@/lib/hooks/workspaces";
import { useChatIdentity } from "./chat-context";
import { useChatTyping } from "./typing";
import { useChatMayaAction, useChatMayaRuns } from "./use-chat-maya-runs";
import type { ChatMessage } from "./types";

export const TypingIndicator = ({
  conversationId,
  threadRootId,
  messages,
  active = true,
}: {
  conversationId: string;
  threadRootId?: string;
  messages: ChatMessage[];
  active?: boolean;
}) => {
  const { workspace } = useCurrentWorkspace();
  const { userId } = useChatIdentity();
  const people = useChatTyping(
    workspace?.id ?? "",
    conversationId,
    threadRootId,
  ).filter((person) => person.id !== userId);
  const maya = useChatMayaRuns(conversationId);
  const action = useChatMayaAction(conversationId);
  const sourceIds = new Set(messages.map((message) => message.id));
  if (threadRootId) sourceIds.add(threadRootId);
  const thinking = maya.data?.runs.find(
    (run) =>
      (run.status === "queued" || run.status === "running") &&
      sourceIds.has(run.sourceId),
  );
  let text = "";
  if (people.length === 1) text = `${people[0].fullName} is typing…`;
  else if (people.length === 2)
    text = `${people[0].fullName} and ${people[1].fullName} are typing…`;
  else if (people.length > 2)
    text = `${people[0].fullName}, ${people[1].fullName} and ${people.length - 2} others are typing…`;
  if (thinking)
    text = text ? `${text} · Maya is thinking…` : "Maya is thinking…";
  return (
    <div
      aria-atomic="true"
      aria-live="polite"
      className="text-text-muted flex min-h-7 shrink-0 items-center gap-2 pt-1 pr-5 pb-2 pl-6"
      role="status"
    >
      {active && text ? (
        <>
          <span aria-hidden className="flex gap-0.5">
            <span className="size-1 animate-pulse rounded-full bg-current motion-reduce:animate-none" />
            <span className="size-1 animate-pulse rounded-full bg-current [animation-delay:150ms] motion-reduce:animate-none" />
            <span className="size-1 animate-pulse rounded-full bg-current [animation-delay:300ms] motion-reduce:animate-none" />
          </span>
          <span className="truncate">{text}</span>
          {thinking?.requesterId === userId ? (
            <button
              className="text-text-muted hover:text-foreground ml-auto shrink-0 rounded px-1 underline underline-offset-2"
              disabled={action.isPending}
              onClick={() => {
                action.mutate({ run: thinking, action: "cancel" });
              }}
              type="button"
            >
              Cancel
            </button>
          ) : null}
        </>
      ) : null}
    </div>
  );
};
