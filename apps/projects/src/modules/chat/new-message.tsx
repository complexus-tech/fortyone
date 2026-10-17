"use client";

import type { KeyboardEvent } from "react";
import { useEffect, useId, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ChevronLeftIcon } from "icons";
import { cn } from "lib";
import { Avatar, Button, Input, Tooltip } from "ui";
import { useDebounce } from "@/hooks/debounce";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSession } from "@/lib/auth/client";
import { useMembersInfinite } from "@/lib/hooks/members";
import type { Member } from "@/types/member";
import { createDirectConversation } from "./api";
import { ConversationMark } from "./avatars";
import type { ChatDraft } from "./chat-rich-text";
import { ICON_BUTTON_CLASS } from "./chat-styles";
import { Composer } from "./composer";
import { chatKeys } from "./constants";
import { useChatConversations } from "./queries";
import type { ChatConversation } from "./types";

type Recipient =
  | { kind: "person"; id: string; name: string; person: Member }
  | { kind: "team"; id: string; name: string; conversation: ChatConversation };

const ignoreSend = () => undefined;

const getRecipientHint = (
  recipient: Recipient,
  openingPersonId: string | null,
) => {
  if (recipient.kind === "person" && recipient.person.id === openingPersonId) {
    return "Opening…";
  }
  return recipient.kind === "person" ? "Person" : "Team";
};

const RecipientQueryStatus = ({
  kind,
  error,
  hasMore,
  loadingMore,
  disabled,
  onRetry,
  onLoadMore,
}: {
  kind: "people" | "teams";
  error: boolean;
  hasMore: boolean;
  loadingMore: boolean;
  disabled: boolean;
  onRetry: () => void;
  onLoadMore: () => void;
}) => {
  if (error) {
    return (
      <div
        className="text-text-muted flex items-center gap-2 px-3 py-2"
        role="alert"
      >
        Could not load {kind}.
        <Button
          color="tertiary"
          disabled={disabled}
          onClick={onRetry}
          size="sm"
          variant="naked"
        >
          Retry {kind}
        </Button>
      </div>
    );
  }
  return hasMore ? (
    <Button
      className="m-1"
      color="tertiary"
      disabled={disabled || loadingMore}
      onClick={onLoadMore}
      size="sm"
      variant="naked"
    >
      {loadingMore ? `Loading ${kind}…` : `Load more ${kind}`}
    </Button>
  ) : null;
};

export const NewMessage = ({
  draft,
  onDraftChange,
  onSelect,
  onBack,
}: {
  draft: ChatDraft;
  onDraftChange: (draft: ChatDraft) => void;
  onSelect: (conversationId: string) => void;
  onBack: () => void;
}) => {
  const { workspaceSlug } = useWorkspacePath();
  const { data: session } = useSession();
  const client = useQueryClient();
  const [search, setSearch] = useState("");
  const [query, setQuery] = useState("");
  const [expanded, setExpanded] = useState(true);
  const [activeIndex, setActiveIndex] = useState(0);
  const searchLater = useDebounce(setQuery, 200);
  const people = useMembersInfinite(query, 15, true);
  const teams = useChatConversations(query);
  const inputId = useId();
  const listId = useId();
  const activeOptionRef = useRef<HTMLButtonElement>(null);
  const mounted = useRef(true);
  const pendingPerson = useRef<string | null>(null);
  const leaving = useRef(false);

  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);

  const create = useMutation({
    mutationFn: (userId: string) =>
      createDirectConversation({ workspaceSlug }, userId),
    onSuccess: (conversation) => {
      client.setQueryData(
        chatKeys.conversation(workspaceSlug, conversation.id),
        conversation,
      );
      void client.invalidateQueries({
        queryKey: chatKeys.conversations(workspaceSlug),
      });
      if (mounted.current && !leaving.current) {
        leaving.current = true;
        onSelect(conversation.id);
      }
    },
    onSettled: () => {
      pendingPerson.current = null;
    },
  });
  const matchingPeople: Recipient[] = (
    people.data?.pages.flatMap((page) => page.members) ?? []
  )
    .filter(
      (person) =>
        person.isActive && !person.isSystem && person.id !== session?.user.id,
    )
    .map((person) => ({
      kind: "person",
      id: `person:${person.id}`,
      name: person.fullName || person.username,
      person,
    }));
  const matchingTeams: Recipient[] = (
    teams.data?.pages.flatMap((page) => page.items) ?? []
  )
    .filter((conversation) => conversation.kind === "team")
    .map((conversation) => ({
      kind: "team",
      id: `team:${conversation.id}`,
      name: conversation.name,
      conversation,
    }));
  const recipients = [
    ...new Map(
      [...matchingPeople, ...matchingTeams].map((recipient) => [
        recipient.id,
        recipient,
      ]),
    ).values(),
  ];
  const waitingForQuery = search.trim() !== query;
  const activeRecipient = waitingForQuery
    ? undefined
    : recipients[Math.min(activeIndex, recipients.length - 1)];
  const activeOptionId = activeRecipient
    ? `${listId}-${activeRecipient.id}`
    : undefined;
  const loading = waitingForQuery || people.isPending || teams.isPending;

  useEffect(() => {
    if (expanded) activeOptionRef.current?.scrollIntoView({ block: "nearest" });
  }, [activeOptionId, expanded]);

  const selectRecipient = (recipient: Recipient) => {
    if (pendingPerson.current || leaving.current || waitingForQuery) return;
    if (recipient.kind === "team") {
      leaving.current = true;
      onSelect(recipient.conversation.id);
      return;
    }
    pendingPerson.current = recipient.person.id;
    create.mutate(recipient.person.id);
  };

  const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
    if (event.nativeEvent.isComposing) return;
    if (event.key === "Escape") {
      event.preventDefault();
      setExpanded(false);
      return;
    }
    if (event.key === "ArrowDown" || event.key === "ArrowUp") {
      event.preventDefault();
      setExpanded(true);
      if (!recipients.length || waitingForQuery) return;
      const direction = event.key === "ArrowDown" ? 1 : -1;
      setActiveIndex((current) =>
        !expanded
          ? 0
          : (current + direction + recipients.length) % recipients.length,
      );
    } else if (event.key === "Enter" && expanded && activeRecipient) {
      event.preventDefault();
      selectRecipient(activeRecipient);
    }
  };

  return (
    <section
      aria-label="New message"
      className="flex min-h-0 min-w-0 flex-1 flex-col"
    >
      <header className="border-border flex h-16 shrink-0 items-center gap-2 border-b-[0.5px] px-4">
        <Tooltip title="Back to conversations">
          <Button
            aria-label="Back to conversations"
            asIcon
            className={ICON_BUTTON_CLASS}
            color="tertiary"
            onClick={() => {
              leaving.current = true;
              onBack();
            }}
            size="sm"
            type="button"
            variant="naked"
          >
            <ChevronLeftIcon className="size-4" />
          </Button>
        </Tooltip>
        <h2 className="font-semibold">New message</h2>
      </header>
      <div
        className="flex min-h-0 flex-1 flex-col"
        onBlur={(event) => {
          if (!event.currentTarget.contains(event.relatedTarget as Node | null))
            setExpanded(false);
        }}
      >
        <div className="border-border/60 flex shrink-0 items-center gap-2 border-b-[0.5px] px-5 py-2">
          <label className="text-text-muted shrink-0" htmlFor={inputId}>
            To:
          </label>
          <div className="min-w-0 flex-1">
            <Input
              aria-activedescendant={expanded ? activeOptionId : undefined}
              aria-autocomplete="list"
              aria-controls={expanded ? listId : undefined}
              aria-expanded={expanded}
              aria-label="Find a person or team"
              autoComplete="off"
              autoFocus
              className="h-10 border-0 bg-transparent px-2 text-base focus-visible:ring-0 dark:bg-transparent"
              disabled={create.isPending}
              id={inputId}
              onChange={(event) => {
                setSearch(event.target.value);
                searchLater(event.target.value.trim());
                setActiveIndex(0);
                setExpanded(true);
              }}
              onFocus={() => {
                setExpanded(true);
              }}
              onKeyDown={handleKeyDown}
              placeholder="Find a person or team…"
              role="combobox"
              value={search}
            />
          </div>
        </div>
        <div className="min-h-0 flex-1 overflow-y-auto px-5 py-3">
          {expanded ? (
            <div className="border-border/60 overflow-hidden rounded-xl border">
              <div
                aria-busy={loading || create.isPending}
                aria-label="People and teams"
                id={listId}
                role="listbox"
              >
                {!waitingForQuery
                  ? recipients.map((recipient, index) => (
                      <button
                        aria-selected={activeRecipient?.id === recipient.id}
                        className={cn(
                          "hover:bg-state-hover/60 focus-visible:ring-primary/40 flex w-full items-center gap-3 px-3 py-2.5 text-left outline-none focus-visible:ring-2 disabled:cursor-wait",
                          activeRecipient?.id === recipient.id &&
                            "bg-state-hover/60",
                        )}
                        disabled={create.isPending}
                        id={`${listId}-${recipient.id}`}
                        key={recipient.id}
                        onClick={() => {
                          selectRecipient(recipient);
                        }}
                        onMouseDown={(event) => {
                          event.preventDefault();
                        }}
                        onMouseEnter={() => {
                          setActiveIndex(index);
                        }}
                        ref={
                          activeRecipient?.id === recipient.id
                            ? activeOptionRef
                            : undefined
                        }
                        role="option"
                        tabIndex={-1}
                        type="button"
                      >
                        <span className="flex size-7 shrink-0 items-center justify-center">
                          {recipient.kind === "person" ? (
                            <Avatar
                              name={recipient.name}
                              size="sm"
                              src={recipient.person.avatarUrl}
                            />
                          ) : (
                            <ConversationMark
                              conversation={recipient.conversation}
                            />
                          )}
                        </span>
                        <span className="min-w-0 flex-1 truncate font-medium">
                          {recipient.name}
                        </span>
                        <span className="text-text-muted shrink-0">
                          {getRecipientHint(
                            recipient,
                            create.isPending ? create.variables : null,
                          )}
                        </span>
                      </button>
                    ))
                  : null}
              </div>
              {loading ? (
                <p className="text-text-muted px-3 py-3" role="status">
                  Finding people and teams…
                </p>
              ) : null}
              {!loading &&
              !recipients.length &&
              !people.isError &&
              !teams.isError ? (
                <p className="text-text-muted px-3 py-4" role="status">
                  No people or teams found.
                </p>
              ) : null}
              <RecipientQueryStatus
                disabled={create.isPending || waitingForQuery}
                error={people.isError}
                hasMore={Boolean(people.hasNextPage)}
                kind="people"
                loadingMore={people.isFetchingNextPage}
                onLoadMore={() => {
                  void people.fetchNextPage();
                }}
                onRetry={() => {
                  void people.refetch();
                }}
              />
              <RecipientQueryStatus
                disabled={create.isPending || waitingForQuery}
                error={teams.isError}
                hasMore={Boolean(teams.hasNextPage)}
                kind="teams"
                loadingMore={teams.isFetchingNextPage}
                onLoadMore={() => {
                  void teams.fetchNextPage();
                }}
                onRetry={() => {
                  void teams.refetch();
                }}
              />
            </div>
          ) : null}
          {create.isError ? (
            <div
              className="text-danger mt-3 flex flex-wrap items-center gap-2"
              role="alert"
            >
              <span>
                Could not start this conversation. {create.error.message}
              </span>
              <Button
                color="tertiary"
                onClick={() => {
                  if (
                    !create.variables ||
                    pendingPerson.current ||
                    leaving.current
                  )
                    return;
                  pendingPerson.current = create.variables;
                  create.mutate(create.variables);
                }}
                size="sm"
                type="button"
                variant="naked"
              >
                Try again
              </Button>
            </div>
          ) : null}
        </div>
      </div>
      <Composer
        allowMaya={false}
        draft={draft}
        label="Start a new message…"
        onDraftChange={onDraftChange}
        onSend={ignoreSend}
      />
    </section>
  );
};
