"use client";

import type { ReactNode, RefObject } from "react";
import { useRef, useState } from "react";
import Link from "next/link";
import {
  ArchiveIcon,
  ArrowRight2Icon,
  EditIcon,
  MoreHorizontalIcon,
  ViewsIcon,
} from "icons";
import { Button, Flex, Menu, Text } from "ui";
import { toast } from "sonner";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { SavedViewIcon } from "@/shared/views/icons";
import { usePresetMutations, useWorkPresets } from "./hooks";
import { PresetNameDialog } from "./name-dialog";
import type { Preset } from "./types";
import type { SavedView } from "./resolve-views";
import { ViewFavoriteButton } from "./view-favorite-button";

export const SavedViewIdentity = ({
  view,
  triggerRef,
  onSelect,
  onArchived,
  selectionControl,
}: {
  selectionControl?: ReactNode;
  view: SavedView;
  triggerRef: RefObject<HTMLButtonElement | null>;
  onSelect: (preset: Preset) => void;
  onArchived?: () => void;
}) => {
  const [open, setOpen] = useState(false);
  const [renaming, setRenaming] = useState(false);
  const openingDialog = useRef(false);
  const { withWorkspace } = useWorkspacePath();
  const query = useWorkPresets(view.teamId, "view", open);
  const { rename, archive } = usePresetMutations(view.teamId, "view");
  const presets = query.data?.pages.flatMap((page) => page.items) ?? [];
  const archiveView = async () => {
    if (archive.isPending) return;
    try {
      await archive.mutateAsync(view.id);
      onArchived?.();
    } catch (error) {
      toast.error("Could not archive view", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };
  return (
    <>
      <Flex align="center" className="min-w-0 gap-2">
        {selectionControl ?? (
          <>
            <SavedViewIcon
              className="h-4.5 w-auto shrink-0"
              configuration={view.configuration}
            />
            <Text as="h1" className="min-w-0 truncate" title={view.name}>
              {view.name}
            </Text>
          </>
        )}
        <ViewFavoriteButton name={view.name} view={view} />
        <Menu onOpenChange={setOpen} open={open}>
          <Menu.Button>
            <Button
              aria-label={`Manage ${view.name}`}
              asIcon
              color="tertiary"
              data-view-menu-trigger
              ref={triggerRef}
              size="sm"
              variant="naked"
            >
              <MoreHorizontalIcon className="h-4.5 w-auto" />
            </Button>
          </Menu.Button>
          <Menu.Items
            align="end"
            className="w-56"
            onCloseAutoFocus={(event) => {
              if (!openingDialog.current) return;
              event.preventDefault();
              openingDialog.current = false;
            }}
          >
            <Menu.Group>
              <Menu.SubMenu>
                <Menu.SubTrigger>
                  <ViewsIcon />
                  Switch view
                  <ArrowRight2Icon className="ml-auto h-3.5 w-auto" />
                </Menu.SubTrigger>
                <Menu.SubItems className="max-h-80 w-64 overflow-y-auto">
                  <Menu.Group>
                    {query.isPending ? (
                      <Text className="px-2 py-2" color="muted" role="status">
                        Loading views...
                      </Text>
                    ) : null}
                    {query.isError ? (
                      <>
                        <Text className="px-2 py-2" role="alert">
                          Views could not be loaded.
                        </Text>
                        <Menu.Item
                          onSelect={(event) => {
                            event.preventDefault();
                            void query.refetch();
                          }}
                        >
                          Try again
                        </Menu.Item>
                      </>
                    ) : (
                      presets.map((preset) => (
                        <Menu.Item
                          active={preset.id === view.id}
                          key={preset.id}
                          onSelect={() => {
                            onSelect(preset);
                          }}
                        >
                          <SavedViewIcon
                            configuration={(preset as SavedView).configuration}
                          />
                          <span className="truncate">{preset.name}</span>
                        </Menu.Item>
                      ))
                    )}
                    {query.hasNextPage ? (
                      <Menu.Item
                        disabled={query.isFetchingNextPage}
                        onSelect={(event) => {
                          event.preventDefault();
                          void query.fetchNextPage();
                        }}
                      >
                        {query.isFetchingNextPage ? "Loading..." : "Load more"}
                      </Menu.Item>
                    ) : null}
                  </Menu.Group>
                </Menu.SubItems>
              </Menu.SubMenu>
              <Menu.Item asChild>
                <Link href={withWorkspace("/views")}>
                  <ViewsIcon />
                  Browse all views
                </Link>
              </Menu.Item>
              {view.canEdit ? (
                <>
                  <Menu.Separator />
                  <Menu.Item
                    onSelect={() => {
                      openingDialog.current = true;
                      openDialogAfterMenuClose(() => {
                        setRenaming(true);
                      });
                    }}
                  >
                    <EditIcon />
                    Rename view
                  </Menu.Item>
                  <Menu.Item
                    disabled={archive.isPending}
                    onSelect={() => void archiveView()}
                  >
                    <ArchiveIcon />
                    Archive view
                  </Menu.Item>
                </>
              ) : null}
            </Menu.Group>
          </Menu.Items>
        </Menu>
      </Flex>
      {renaming ? (
        <PresetNameDialog
          initialName={view.name}
          onOpenChange={setRenaming}
          onReturnFocus={() => triggerRef.current?.focus()}
          onSave={(name) => rename.mutateAsync({ id: view.id, name })}
          open
          title="Rename view"
        />
      ) : null}
    </>
  );
};
