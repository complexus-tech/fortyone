"use client";

import type { RefObject } from "react";
import { useRef, useState } from "react";
import {
  ArchiveIcon,
  ArrowDownIcon,
  MoreHorizontalIcon,
  EditIcon,
} from "icons";
import { Box, Button, Flex, Menu, Popover, Text } from "ui";
import { toast } from "sonner";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";
import { usePresetMutations, useWorkPresets } from "./hooks";
import { PresetNameDialog } from "./name-dialog";
import type { Preset, PresetKind } from "./types";

export const PresetPicker = ({
  teamId,
  kind,
  label,
  onSelect,
  onSaveCurrent,
  onCreate,
  disabled = false,
  hideWhenEmpty = false,
  browseHref,
  triggerRef,
}: {
  teamId: string;
  kind: PresetKind;
  label: string;
  onSelect: (preset: Preset) => void;
  onSaveCurrent?: () => void;
  onCreate?: () => void;
  disabled?: boolean;
  hideWhenEmpty?: boolean;
  browseHref?: string;
  triggerRef?: RefObject<HTMLButtonElement | null>;
}) => {
  const [open, setOpen] = useState(false);
  const [renaming, setRenaming] = useState<Preset | null>(null);
  const pickerTrigger = useRef<HTMLButtonElement | null>(null);
  const openingDialog = useRef(false);
  const query = useWorkPresets(teamId, kind, open || hideWhenEmpty);
  const { rename, archive } = usePresetMutations(teamId, kind);
  const presets = query.data?.pages.flatMap((page) => page.items) ?? [];
  const archiveItem = async (id: string) => {
    try {
      await archive.mutateAsync(id);
      toast.success("Preset archived");
    } catch (error) {
      toast.error("Could not archive preset", {
        description:
          error instanceof Error ? error.message : "Please try again.",
      });
    }
  };
  if (hideWhenEmpty && (query.isPending || query.isError || !presets.length))
    return null;
  return (
    <>
      <Popover
        onOpenChange={(nextOpen) => {
          if (nextOpen) openingDialog.current = false;
          setOpen(nextOpen);
        }}
        open={open}
      >
        <Popover.Trigger asChild>
          <Button
            className="max-w-48 min-w-0"
            color="tertiary"
            data-view-menu-trigger={kind === "view" ? "" : undefined}
            disabled={disabled || !teamId}
            ref={triggerRef ?? pickerTrigger}
            rightIcon={<ArrowDownIcon className="h-3.5 w-auto" />}
            size="sm"
            variant="outline"
          >
            <span className="min-w-0 truncate">{label}</span>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="end"
          className="w-80 p-3"
          onCloseAutoFocus={(event) => {
            if (openingDialog.current) event.preventDefault();
          }}
        >
          <Text className="mb-2 px-2" fontWeight="medium">
            {kind === "view" ? "Saved views" : "Task templates"}
          </Text>
          {query.isError ? (
            <Box className="p-2">
              <Text role="alert">Could not load presets.</Text>
              <Button
                color="tertiary"
                onClick={() => void query.refetch()}
                variant="naked"
              >
                Try again
              </Button>
            </Box>
          ) : null}
          {!query.isError && query.isPending ? (
            <Text className="p-2" color="muted">
              Loading...
            </Text>
          ) : null}
          {!query.isError && !query.isPending && presets.length > 0 ? (
            <Box className="max-h-72 overflow-y-auto">
              {presets.map((preset) => (
                <Flex align="center" className="gap-1" key={preset.id}>
                  <Button
                    className="min-w-0 flex-1 justify-start text-left"
                    color="tertiary"
                    onClick={() => {
                      onSelect(preset);
                      setOpen(false);
                    }}
                    variant="naked"
                  >
                    <span className="truncate">{preset.name}</span>
                    <span className="text-text-muted ml-auto">
                      {preset.visibility === "team" ? "Team" : "Private"}
                    </span>
                  </Button>
                  {preset.canEdit ? (
                    <Menu>
                      <Menu.Button>
                        <Button
                          aria-label={`Manage ${preset.name}`}
                          asIcon
                          color="tertiary"
                          variant="naked"
                        >
                          <MoreHorizontalIcon className="h-4 w-auto" />
                        </Button>
                      </Menu.Button>
                      <Menu.Items align="end">
                        <Menu.Item
                          onSelect={() => {
                            openingDialog.current = true;
                            setOpen(false);
                            openDialogAfterMenuClose(() => {
                              setRenaming(preset);
                            });
                          }}
                        >
                          <EditIcon />
                          Rename
                        </Menu.Item>
                        <Menu.Item
                          disabled={archive.isPending}
                          onSelect={() => void archiveItem(preset.id)}
                        >
                          <ArchiveIcon />
                          Archive
                        </Menu.Item>
                      </Menu.Items>
                    </Menu>
                  ) : null}
                </Flex>
              ))}
            </Box>
          ) : null}
          {!query.isError && !query.isPending && presets.length === 0 ? (
            <Text className="p-2" color="muted">
              {kind === "view"
                ? "Save your filters and layout to revisit them or share with your team."
                : "Save a task as a template to reuse its description and properties."}
            </Text>
          ) : null}
          {query.hasNextPage ? (
            <Button
              className="mt-2 w-full"
              color="tertiary"
              loading={query.isFetchingNextPage}
              onClick={() => void query.fetchNextPage()}
              variant="naked"
            >
              Load more
            </Button>
          ) : null}
          {onSaveCurrent ? (
            <Button
              className="border-border mt-3 w-full border-t pt-3"
              color="tertiary"
              onClick={() => {
                openingDialog.current = true;
                setOpen(false);
                openDialogAfterMenuClose(onSaveCurrent);
              }}
              variant="naked"
            >
              Save current view
            </Button>
          ) : null}
          {onCreate ? (
            <Button
              className="mt-2 w-full justify-start"
              color="tertiary"
              onClick={() => {
                openingDialog.current = true;
                setOpen(false);
                openDialogAfterMenuClose(onCreate);
              }}
              variant="naked"
            >
              Create view
            </Button>
          ) : null}
          {browseHref ? (
            <Button
              className="mt-2 w-full justify-start"
              color="tertiary"
              href={browseHref}
              onClick={() => {
                setOpen(false);
              }}
              variant="naked"
            >
              Browse all views
            </Button>
          ) : null}
        </Popover.Content>
      </Popover>
      {renaming ? (
        <PresetNameDialog
          initialName={renaming.name}
          onOpenChange={(value) => {
            if (!value) setRenaming(null);
          }}
          onReturnFocus={() => (triggerRef ?? pickerTrigger).current?.focus()}
          onSave={(name) => rename.mutateAsync({ id: renaming.id, name })}
          open
          title={`Rename ${kind}`}
        />
      ) : null}
    </>
  );
};
