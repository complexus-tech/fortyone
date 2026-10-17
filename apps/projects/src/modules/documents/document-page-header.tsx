import { cn } from "lib";
import { Avatar, Box, Button, Flex, Menu, Skeleton, Text } from "ui";
import {
  ArchiveIcon,
  HistoryIcon,
  ArrowLeftIcon,
  CopyIcon,
  DeleteIcon,
  DuplicateIcon,
  Comment01Icon,
  CheckIcon,
  InternetIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  UserMultiple02Icon,
} from "icons";
import type { WorkspaceDocument } from "./types";
import { DocumentAccessMenu } from "./document-access-menu";
import styles from "./document-page.module.css";

const documentAccessLabels = {
  private: "Private",
  restricted: "Shared",
  workspace: "Workspace",
} as const;
const DOCUMENT_HEADER_BACKDROP_CLASS_NAME =
  "pointer-events-none absolute inset-x-0 top-0 z-20 h-18";

type DocumentPageHeaderProps = {
  document: WorkspaceDocument;
  title: string;
  saveStatus: string;
  canManageDocument: boolean;
  canDuplicate: boolean;
  isDuplicating: boolean;
  peers: { id: string; name: string }[];
  commentCount: number;
  onBack: () => void;
  onCopyLink: () => void;
  onDuplicate: () => void;
  onShowHistory: () => void;
  onArchive: () => void;
  onDelete: () => void;
  onShowComments: () => void;
};

export function DocumentPageHeader({
  document,
  title,
  saveStatus,
  canManageDocument,
  canDuplicate,
  isDuplicating,
  peers,
  commentCount,
  onBack,
  onCopyLink,
  onDuplicate,
  onShowHistory,
  onArchive,
  onDelete,
  onShowComments,
}: DocumentPageHeaderProps) {
  const isPublic = document.publicEnabled ?? Boolean(document.publicToken);
  const accessLabel = isPublic
    ? "Public"
    : documentAccessLabels[document.visibility];
  let AccessIcon = UserMultiple02Icon;
  if (document.visibility === "private") AccessIcon = LockKeyholeIcon;
  if (isPublic) AccessIcon = InternetIcon;
  const saveFeedback = (
    <span
      aria-live="polite"
      className="text-text-muted flex min-w-0 items-center gap-1.5 text-base"
      role="status"
      title={saveStatus}
    >
      {saveStatus === "Saved" ? (
        <CheckIcon aria-hidden className="size-4 shrink-0" />
      ) : null}
      <span className="truncate">{saveStatus}</span>
    </span>
  );
  return (
    <Box
      className={cn(DOCUMENT_HEADER_BACKDROP_CLASS_NAME, styles.headerBackdrop)}
    >
      <Flex
        align="center"
        className="pointer-events-auto relative z-10 h-18 gap-2 px-4 text-base md:px-5"
        justify="between"
      >
        <Flex align="center" className="min-w-0 flex-1" gap={2}>
          <Button
            aria-label="Back to documents"
            asIcon
            className="md:hidden"
            color="tertiary"
            onClick={() => {
              onBack();
            }}
            variant="naked"
          >
            <ArrowLeftIcon />
          </Button>
          <Box className="min-w-0">
            <Text
              className="max-w-80 truncate"
              fontSize="lg"
              fontWeight="semibold"
            >
              {title || "Untitled document"}
            </Text>
            <Box className="md:hidden">{saveFeedback}</Box>
          </Box>
          <Menu>
            <Menu.Button>
              <Button
                aria-label="Document actions"
                asIcon
                className="shrink-0"
                color="tertiary"
                variant="naked"
              >
                <MoreHorizontalIcon />
              </Button>
            </Menu.Button>
            <Menu.Items align="start" className="min-w-52 text-base">
              <Menu.Group>
                <Menu.Item onSelect={onCopyLink}>
                  <CopyIcon />
                  Copy link
                </Menu.Item>
                <Menu.Item disabled={!canDuplicate} onSelect={onDuplicate}>
                  <DuplicateIcon />
                  {isDuplicating ? "Duplicating..." : "Duplicate document"}
                </Menu.Item>
              </Menu.Group>
              <Menu.Group>
                <Menu.Item
                  onSelect={() => {
                    onShowHistory();
                  }}
                >
                  <HistoryIcon />
                  Version history
                </Menu.Item>
                {canManageDocument ? (
                  <Menu.Item onSelect={onArchive}>
                    <ArchiveIcon />
                    Archive document
                  </Menu.Item>
                ) : (
                  <Menu.Item disabled>Archive unavailable</Menu.Item>
                )}
              </Menu.Group>
              <Menu.Separator />
              <Menu.Group>
                {canManageDocument ? (
                  <Menu.Item
                    className="text-danger dark:text-danger!"
                    onSelect={() => {
                      onDelete();
                    }}
                  >
                    <DeleteIcon className="text-danger dark:text-danger!" />
                    Delete permanently...
                  </Menu.Item>
                ) : (
                  <Menu.Item disabled>Delete unavailable</Menu.Item>
                )}
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
        <Flex align="center" className="shrink-0" gap={2}>
          <Box className="hidden max-w-52 md:block">{saveFeedback}</Box>
          <Flex
            align="center"
            aria-label={`Document access: ${accessLabel}`}
            className="text-text-muted shrink-0"
            gap={1}
            title={
              isPublic
                ? "Anyone with the public link can view"
                : `Workspace access: ${accessLabel}`
            }
          >
            <AccessIcon
              aria-hidden
              className="hidden size-4 shrink-0 lg:block"
            />
            <Text as="span" color="muted">
              {accessLabel}
            </Text>
          </Flex>
          <Flex className="hidden -space-x-2 lg:flex">
            {peers.slice(0, 3).map((peer) => (
              <span key={peer.id} title={peer.name}>
                <Avatar name={peer.name} size="sm" />
              </span>
            ))}
          </Flex>
          <Button
            aria-label={`${commentCount} ${commentCount === 1 ? "comment" : "comments"}`}
            className="gap-1.5 px-2 text-base sm:px-3"
            color="tertiary"
            leftIcon={<Comment01Icon className="size-5" strokeWidth={2} />}
            onClick={() => {
              onShowComments();
            }}
            variant="naked"
          >
            <span className="hidden xl:inline">Comments</span>
            {commentCount > 0 ? commentCount : null}
          </Button>
          {canManageDocument ? (
            <DocumentAccessMenu document={document} />
          ) : null}
        </Flex>
      </Flex>
    </Box>
  );
}

export const DocumentPageSkeleton = () => (
  <Box className="relative h-full min-h-0">
    <Box
      className={cn(DOCUMENT_HEADER_BACKDROP_CLASS_NAME, styles.headerBackdrop)}
    >
      <Flex
        align="center"
        className="pointer-events-auto relative z-10 h-18 px-5"
        justify="between"
      >
        <Skeleton className="h-5 w-44" />
        <Skeleton className="h-8 w-24" />
      </Flex>
    </Box>
    <Box className="mx-auto w-full max-w-5xl px-8 pt-34 pb-16 sm:px-10 lg:px-12 lg:pt-34">
      <Skeleton className="mb-8 h-12 w-3/4" />
      <Skeleton className="mb-3 h-6 w-full" />
      <Skeleton className="mb-3 h-6 w-5/6" />
      <Skeleton className="h-6 w-2/3" />
    </Box>
  </Box>
);
