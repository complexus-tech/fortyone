import Link from "next/link";
import { Avatar, Button, Flex, Menu, Text } from "ui";
import {
  ArchiveIcon,
  Building03Icon,
  DocsIcon,
  InternetIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  ShareIcon,
  StoryIcon,
  UndoIcon,
} from "icons";
import { cn } from "lib";
import type { Member } from "@/types";
import { FavoriteButton, FavoriteMenuItem } from "@/shared/favorites";
import { formatDocumentRelativeTime } from "./relative-time";
import type { RelatedWork, WorkspaceDocumentSummary } from "./types";

export const DOCUMENT_LIBRARY_COLUMNS =
  "grid grid-cols-[minmax(0,1fr)_2.5rem] items-center gap-x-4 @2xl:grid-cols-[minmax(0,1fr)_8rem_6rem_2.5rem] @4xl:grid-cols-[minmax(13rem,2fr)_minmax(8rem,1fr)_minmax(9rem,1.2fr)_8rem_6rem_2.5rem]";

type DocumentRowProps = {
  document: WorkspaceDocumentSummary;
  isArchived: boolean;
  isArchiving: boolean;
  isOwner: boolean;
  isRestoring: boolean;
  onArchive: () => void;
  onRestore: () => void;
  owner?: Pick<Member, "fullName" | "avatarUrl">;
  withWorkspace: (path: string) => string;
};

const RelatedWorkLink = ({
  work,
  withWorkspace,
}: {
  work: RelatedWork;
  withWorkspace: (path: string) => string;
}) => {
  const Icon = work.entityType === "story" ? StoryIcon : ObjectiveIcon;
  let path: string | null = null;
  if (work.entityType === "story" && work.reference) {
    path = `/work/${encodeURIComponent(work.reference)}`;
  } else if (work.entityType === "objective" && work.teamId) {
    path = `/teams/${work.teamId}/objectives/${work.entityId}`;
  }
  const content = (
    <>
      <Icon aria-hidden className="size-5 shrink-0" />
      <span className="truncate">{work.title}</span>
    </>
  );

  return path ? (
    <Link
      className="text-text-muted hover:text-foreground focus-visible:ring-ring/40 flex min-w-0 items-center gap-2 rounded-md outline-none focus-visible:ring-2"
      href={withWorkspace(path)}
      title={work.title}
    >
      {content}
    </Link>
  ) : (
    <span
      className="text-text-muted flex min-w-0 items-center gap-2"
      title={work.title}
    >
      {content}
    </span>
  );
};

export const DocumentRow = ({
  document,
  isArchived,
  isArchiving,
  isOwner,
  isRestoring,
  onArchive,
  onRestore,
  owner,
  withWorkspace,
}: DocumentRowProps) => {
  let access = { icon: Building03Icon, label: "Workspace" };
  if (document.visibility === "private") {
    access = { icon: LockKeyholeIcon, label: "Private" };
  } else if (document.visibility === "restricted") {
    access = { icon: ShareIcon, label: "Invited people" };
  }
  if (document.publicEnabled) {
    access = { icon: InternetIcon, label: "Public" };
  }
  const AccessIcon = access.icon;
  const relatedWork = document.relatedWork?.[0];
  const additionalWork = Math.max(0, (document.relatedWork?.length ?? 0) - 1);
  const ownerName = owner?.fullName || (isOwner ? "You" : "Workspace member");
  const date = isArchived
    ? document.archivedAt ?? document.updatedAt
    : document.updatedAt;
  const title = (
    <>
      <DocsIcon
        aria-hidden
        className="text-text-muted size-5 shrink-0"
        strokeWidth={2}
      />
      <span className="min-w-0 truncate font-medium">{document.title}</span>
    </>
  );

  return (
    <li
      className={cn(
        DOCUMENT_LIBRARY_COLUMNS,
        "group border-border/70 hover:bg-state-hover/50 border-b-[0.5px] py-2 transition-colors",
      )}
    >
      <div className="min-w-0">
        <Flex align="center" className="relative min-w-0" gap={2}>
          {!isArchived ? (
            <FavoriteButton
              className="pointer-events-none absolute top-1/2 -left-8 hidden h-8 w-8 -translate-y-1/2 text-base opacity-0 transition-opacity group-focus-within:pointer-events-auto group-focus-within:opacity-100 group-hover:pointer-events-auto group-hover:opacity-100 md:flex"
              item={{ kind: "document", id: document.id }}
              name={document.title}
              tooltip="Favorite"
            />
          ) : null}
          {isArchived ? (
            <span className="flex min-w-0 items-center gap-2.5">{title}</span>
          ) : (
            <Link
              className="focus-visible:ring-ring/40 flex min-w-0 flex-1 items-center gap-2.5 rounded-md outline-none focus-visible:ring-2"
              href={withWorkspace(`/docs/${document.id}`)}
              title={document.title}
            >
              {title}
            </Link>
          )}
        </Flex>
        <Flex
          className="text-text-muted mt-2 min-w-0 gap-x-3 gap-y-2 pl-7 @4xl:hidden"
          wrap
        >
          <span className="truncate">
            {ownerName}
            {isOwner && owner ? " (You)" : ""}
          </span>
          {relatedWork ? (
            <RelatedWorkLink withWorkspace={withWorkspace} work={relatedWork} />
          ) : null}
        </Flex>
        <Flex
          align="center"
          className="text-text-muted mt-2 gap-x-4 gap-y-2 pl-7 @2xl:hidden"
          wrap
        >
          <span className="flex items-center gap-2">
            <AccessIcon aria-hidden className="size-5" />
            {access.label}
          </span>
          <time dateTime={date}>{formatDocumentRelativeTime(date)}</time>
        </Flex>
      </div>
      <Flex align="center" className="hidden min-w-0 @4xl:flex" gap={2}>
        <Avatar name={ownerName} size="sm" src={owner?.avatarUrl} />
        <Text className="truncate" color="muted" title={ownerName}>
          {ownerName}
          {isOwner && owner ? " (You)" : ""}
        </Text>
      </Flex>
      <Flex align="center" className="hidden min-w-0 @4xl:flex" gap={2}>
        {relatedWork ? (
          <>
            <RelatedWorkLink withWorkspace={withWorkspace} work={relatedWork} />
            {additionalWork > 0 ? (
              <Text
                as="span"
                className="shrink-0"
                color="muted"
                title={`${additionalWork} more linked items`}
              >
                +{additionalWork}
              </Text>
            ) : null}
          </>
        ) : (
          <Text color="muted">—</Text>
        )}
      </Flex>
      <Flex align="center" className="hidden @2xl:flex" gap={2}>
        <AccessIcon
          aria-hidden
          className="text-text-muted size-5 shrink-0"
          strokeWidth={2}
        />
        <Text color="muted">{access.label}</Text>
      </Flex>
      <time
        className="text-text-muted hidden @2xl:block"
        dateTime={date}
        title={new Date(date).toLocaleString()}
      >
        {formatDocumentRelativeTime(date)}
      </time>
      <div className="self-start @2xl:self-auto">
        {isArchived ? (
          <Button
            aria-label={`Restore ${document.title}`}
            asIcon
            color="tertiary"
            disabled={isRestoring}
            loading={isRestoring}
            onClick={onRestore}
            title="Restore document"
            variant="naked"
          >
            <UndoIcon />
          </Button>
        ) : null}
        {!isArchived ? (
          <Menu>
            <Menu.Button>
              <Button
                aria-label={`Actions for ${document.title}`}
                asIcon
                className="opacity-100 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100 data-[state=open]:opacity-100 [@media(hover:hover)]:opacity-0"
                color="tertiary"
                disabled={isArchiving}
                variant="naked"
              >
                <MoreHorizontalIcon />
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="min-w-52">
              <Menu.Group>
                <Menu.Item asChild>
                  <Link href={withWorkspace(`/docs/${document.id}`)}>
                    <DocsIcon className="size-5" strokeWidth={2} />
                    Open document
                  </Link>
                </Menu.Item>
                <FavoriteMenuItem
                  item={{ kind: "document", id: document.id }}
                  name={document.title}
                />
                {isOwner ? (
                  <Menu.Item onSelect={onArchive}>
                    <ArchiveIcon className="size-5" strokeWidth={2} />
                    Archive document
                  </Menu.Item>
                ) : null}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        ) : null}
      </div>
    </li>
  );
};
