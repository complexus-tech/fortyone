"use client";

import { useRef, useState } from "react";
import Link from "next/link";
import { ArchiveIcon, EditIcon, MoreHorizontalIcon } from "icons";
import { Button, Flex, Menu, Text } from "ui";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { SavedViewIcon } from "@/shared/views/icons";
import { usePresetMutations } from "./hooks";
import { PresetNameDialog } from "./name-dialog";
import { ViewFavoriteButton } from "./view-favorite-button";
import { savedViewPath } from "./view-link";
import type { SavedView } from "./resolve-views";

export const ViewRow = ({
  view,
  detail,
  onArchived,
}: {
  view: SavedView;
  detail: string;
  onArchived?: () => void;
}) => {
  const [renaming, setRenaming] = useState(false);
  const manageButton = useRef<HTMLButtonElement>(null);
  const { rename, archive } = usePresetMutations(view.teamId, "view");
  const { withWorkspace } = useWorkspacePath();
  const archiveView = async () => {
    if (archive.isPending) return;
    try {
      await archive.mutateAsync(view.id);
      onArchived?.();
      toast.success("View archived");
    } catch (error) {
      toast.error("Could not archive view", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };
  return (
    <>
      <Flex
        align="center"
        className="hover:bg-state-hover gap-2 rounded-lg px-3 py-2"
      >
        <Link
          className="focus-visible:ring-ring flex min-w-0 flex-1 items-center gap-3 rounded-md outline-none focus-visible:ring-2"
          href={withWorkspace(savedViewPath(view.teamId, view.id))}
        >
          <SavedViewIcon
            className="text-text-muted h-5 w-auto shrink-0"
            configuration={view.configuration}
          />
          <Text
            className="min-w-0 flex-1 truncate"
            title={`${view.name} · ${view.configuration.description || detail}`}
          >
            {view.name}
          </Text>
          <Text className="hidden max-w-64 truncate md:block" color="muted">
            {detail}
          </Text>
        </Link>
        <ViewFavoriteButton name={view.name} view={view} />
        {view.canEdit ? (
          <Menu>
            <Menu.Button>
              <Button
                aria-label={`Manage ${view.name}`}
                asIcon
                color="tertiary"
                ref={manageButton}
                size="sm"
                variant="naked"
              >
                <MoreHorizontalIcon />
              </Button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Item
                onSelect={() => {
                  openDialogAfterMenuClose(() => {
                    setRenaming(true);
                  });
                }}
              >
                <EditIcon />
                Rename
              </Menu.Item>
              <Menu.Item
                disabled={archive.isPending}
                onSelect={() => void archiveView()}
              >
                <ArchiveIcon />
                Archive
              </Menu.Item>
            </Menu.Items>
          </Menu>
        ) : null}
      </Flex>
      {renaming ? (
        <PresetNameDialog
          initialName={view.name}
          onOpenChange={setRenaming}
          onReturnFocus={() => manageButton.current?.focus()}
          onSave={(name) => rename.mutateAsync({ id: view.id, name })}
          open
          title="Rename view"
        />
      ) : null}
    </>
  );
};
