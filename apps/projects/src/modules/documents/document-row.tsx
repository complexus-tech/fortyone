import Link from "next/link";
import { Button, Flex, Menu, Text } from "ui";
import {
  ArchiveIcon,
  DocsIcon,
  LockKeyholeIcon,
  MoreHorizontalIcon,
  ShareIcon,
} from "icons";
import { formatDocumentRelativeTime } from "./relative-time";
import type { WorkspaceDocumentSummary } from "./types";

type DocumentRowProps = {
  document: WorkspaceDocumentSummary;
  isArchiving: boolean;
  isOwner: boolean;
  onArchive: () => void;
  withWorkspace: (path: string) => string;
};

export const DocumentRow = ({
  document,
  isArchiving,
  isOwner,
  onArchive,
  withWorkspace,
}: DocumentRowProps) => {
  let DocumentIcon = DocsIcon;
  if (document.visibility === "private") DocumentIcon = LockKeyholeIcon;
  if (document.visibility === "restricted") DocumentIcon = ShareIcon;

  return (
    <Flex
      align="center"
      className="border-border/80 hover:bg-state-hover/50 border-b-[0.5px] py-[0.655rem] transition-colors"
      gap={2}
    >
      <Link
        className="focus-visible:ring-ring/40 flex min-w-0 flex-1 items-center gap-2 rounded-sm outline-none focus-visible:ring-1"
        href={withWorkspace(`/docs/${document.id}`)}
      >
        <DocumentIcon
          className="text-text-muted size-[1.1rem] shrink-0"
          strokeWidth={2}
        />
        <span className="min-w-0 flex-1 truncate">{document.title}</span>
        <Text as="span" className="shrink-0" color="muted">
          {formatDocumentRelativeTime(document.updatedAt)}
        </Text>
      </Link>
      {isOwner ? (
        <Menu>
          <Menu.Button>
            <Button
              aria-label={`Actions for ${document.title}`}
              asIcon
              color="tertiary"
              disabled={isArchiving}
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon />
            </Button>
          </Menu.Button>
          <Menu.Items align="end" className="min-w-52">
            <Menu.Group>
              <Menu.Item
                className="text-danger dark:!text-danger"
                onSelect={onArchive}
              >
                <ArchiveIcon
                  className="text-danger dark:!text-danger size-4"
                  strokeWidth={2}
                />
                Archive document
              </Menu.Item>
            </Menu.Group>
          </Menu.Items>
        </Menu>
      ) : null}
    </Flex>
  );
};
