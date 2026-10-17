import type { UIEvent } from "react";
import { useDeferredValue, useId, useState } from "react";
import { SearchIcon } from "icons";
import { Button, Divider, Input, Popover } from "ui";
import { MenuLoadingSkeleton } from "@/components/ui/menu-loading-skeleton";
import { MEMBER_MENU_PAGE_SIZE } from "@/lib/hooks/members";
import { useTeamMembersInfinite } from "@/lib/hooks/team-members";
import { PersonAvatar } from "./avatars";
import { useChatIdentity } from "./chat-context";
import type { ChatConversation, ChatPerson } from "./types";

type MembersPopoverProps = {
  conversation: ChatConversation;
  mayaAvatarUrl?: string | null;
};

const MembersList = ({ conversation, mayaAvatarUrl }: MembersPopoverProps) => {
  const { userId } = useChatIdentity();
  const [search, setSearch] = useState("");
  const query = useDeferredValue(search.trim());
  const membersQuery = useTeamMembersInfinite(
    conversation.teamId ?? undefined,
    query,
    MEMBER_MENU_PAGE_SIZE,
  );
  const initialMembers = membersQuery.isError
    ? []
    : conversation.members.filter((member) => !member.isSystem);
  const members: ChatPerson[] =
    membersQuery.data?.pages
      .flatMap((page) => page.members)
      .filter((member) => member.isActive && !member.isSystem)
      .map((member) => ({
        id: member.id,
        fullName: member.fullName || member.username,
        avatarUrl: member.avatarUrl,
        isSystem: false,
      })) ?? initialMembers;
  const visibleMembers = [
    ...new Map(
      [
        ...members,
        ...conversation.members.filter((member) => member.isSystem),
      ].map((member) => [member.id, member]),
    ).values(),
  ].filter((member) =>
    member.fullName
      .toLocaleLowerCase()
      .includes(search.trim().toLocaleLowerCase()),
  );
  const loading = membersQuery.isPending || query !== search.trim();

  const loadMore = () => {
    if (
      !loading &&
      membersQuery.hasNextPage &&
      !membersQuery.isFetchingNextPage
    ) {
      void membersQuery.fetchNextPage();
    }
  };

  const handleScroll = (event: UIEvent<HTMLDivElement>) => {
    const { scrollHeight, scrollTop, clientHeight } = event.currentTarget;
    if (
      !membersQuery.isError &&
      scrollHeight - scrollTop - clientHeight <= 80
    ) {
      loadMore();
    }
  };

  return (
    <>
      <div className="px-2">
        <Input
          aria-label="Search team members"
          autoComplete="off"
          className="h-9 border-0 bg-transparent px-2 leading-normal focus-visible:ring-0 dark:bg-transparent"
          leftIcon={<SearchIcon className="text-text-muted size-4.5" />}
          onChange={(event) => {
            setSearch(event.target.value);
          }}
          placeholder="Search members…"
          value={search}
        />
      </div>
      <Divider className="my-2" />
      <div
        aria-label="Team members"
        className="max-h-80 overflow-y-auto px-1.5 outline-none md:max-h-100"
        onScroll={handleScroll}
      >
        <ul aria-label="Members" className="space-y-0.5">
          {visibleMembers.map((member) => (
            <li className="flex items-center gap-3 px-2 py-2" key={member.id}>
              <PersonAvatar mayaAvatarUrl={mayaAvatarUrl} person={member} />
              <span className="min-w-0 flex-1 truncate" title={member.fullName}>
                {member.fullName}
              </span>
              {member.id === userId ? (
                <span className="text-text-muted shrink-0">You</span>
              ) : null}
            </li>
          ))}
        </ul>
        {loading || membersQuery.isFetchingNextPage ? (
          <div className="p-2" role="status">
            <span className="sr-only">Loading members…</span>
            <MenuLoadingSkeleton avatar rows={2} trailing={false} />
          </div>
        ) : null}
        {!loading && !membersQuery.isError && !visibleMembers.length ? (
          <p className="text-text-muted px-2 py-4" role="status">
            No members found.
          </p>
        ) : null}
        {membersQuery.isError ? (
          <div className="px-2 py-2" role="alert">
            <p className="text-text-muted">Could not load members.</p>
            <Button
              className="mt-1"
              color="primary"
              disabled={membersQuery.isFetching}
              onClick={() => {
                if (membersQuery.isFetchNextPageError) {
                  loadMore();
                } else {
                  void membersQuery.refetch();
                }
              }}
              size="sm"
              variant="naked"
            >
              Try again
            </Button>
          </div>
        ) : null}
        {!membersQuery.isError && membersQuery.hasNextPage ? (
          <Button
            className="mx-2 my-1"
            color="primary"
            disabled={loading || membersQuery.isFetchingNextPage}
            onClick={loadMore}
            size="sm"
            variant="naked"
          >
            {membersQuery.isFetchingNextPage ? "Loading…" : "Load more members"}
          </Button>
        ) : null}
      </div>
    </>
  );
};

export const ConversationMembersPopover = ({
  conversation,
  mayaAvatarUrl,
}: MembersPopoverProps) => {
  const [open, setOpen] = useState(false);
  const titleId = useId();

  return (
    <Popover onOpenChange={setOpen} open={open}>
      <Popover.Trigger asChild>
        <Button
          aria-label={`View ${conversation.name} team members`}
          className="hidden shrink-0 gap-0 px-1.5 @min-[700px]/conversation:flex"
          color="tertiary"
          size="md"
          type="button"
          variant="naked"
        >
          <span aria-hidden className="isolate flex items-center -space-x-1.5">
            {conversation.members.slice(0, 4).map((person) => (
              <span
                className="bg-surface ring-surface relative inline-flex rounded-full ring-2"
                key={person.id}
              >
                <PersonAvatar mayaAvatarUrl={mayaAvatarUrl} person={person} />
              </span>
            ))}
          </span>
          <span className="text-text-muted ml-3">
            {conversation.memberCount}
          </span>
        </Button>
      </Popover.Trigger>
      <Popover.Content
        align="start"
        aria-labelledby={titleId}
        className="w-80 max-w-[calc(100vw-2rem)]"
      >
        <div className="flex items-center justify-between gap-3 px-3 pb-1">
          <h3 className="font-semibold" id={titleId}>
            Team members
          </h3>
          <span className="text-text-muted">{conversation.memberCount}</span>
        </div>
        <MembersList
          conversation={conversation}
          mayaAvatarUrl={mayaAvatarUrl}
        />
      </Popover.Content>
    </Popover>
  );
};
