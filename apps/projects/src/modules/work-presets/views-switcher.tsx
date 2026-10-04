"use client";

import { useState } from "react";
import { CheckIcon, ChevronRightIcon, ViewsIcon } from "icons";
import {
  Box,
  BreadCrumbs,
  Button,
  Command,
  Flex,
  Popover,
  Skeleton,
  Text,
} from "ui";
import type { RefObject } from "react";
import { useSession } from "@/lib/auth/client";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { SavedViewIcon } from "@/shared/views/icons";
import { useViewsDirectory } from "./use-views-directory";
import { ViewsDirectoryStatus } from "./views-directory-status";
import type { SavedView } from "./resolve-views";

const ViewsPickerContent = ({
  selected,
  onSelect,
}: {
  selected: SavedView | null;
  onSelect: (view: SavedView) => void;
}) => {
  const [search, setSearch] = useState("");
  const directory = useViewsDirectory();
  const term = search.trim().toLocaleLowerCase();
  const matches = directory.views.filter(({ view, team }) =>
    `${view.name} ${team.name}`.toLocaleLowerCase().includes(term),
  );
  let emptyText = "No matching views loaded yet.";
  if (directory.complete)
    emptyText = search ? "No matching views." : "No saved views yet.";
  return (
    <Command label="Search views" shouldFilter={false}>
      <Box className="border-border border-b px-2 pb-2">
        <Command.Input
          aria-label="Search views"
          autoFocus
          onValueChange={setSearch}
          placeholder="Find a view..."
          value={search}
        />
      </Box>
      <Command.List
        className="mt-1 max-h-80 w-full overflow-y-auto border-0 bg-transparent p-1 shadow-none dark:bg-transparent"
        label="Saved views"
      >
        {matches.map(({ view, team }) => (
          <Command.Item
            key={view.id}
            onSelect={() => {
              onSelect(view);
            }}
            value={view.id}
          >
            <SavedViewIcon
              className="h-4.5 w-auto shrink-0"
              configuration={view.configuration}
            />
            <span className="min-w-0 flex-1 truncate">{view.name}</span>
            <Text className="max-w-28 shrink-0 truncate" color="muted">
              {team.name}
            </Text>
            {selected?.id === view.id ? (
              <CheckIcon aria-hidden className="h-4 w-auto shrink-0" />
            ) : null}
          </Command.Item>
        ))}
        {!matches.length &&
        !directory.pending &&
        !directory.teamsError &&
        !directory.failedFeeds.length ? (
          <Text className="px-2 py-3" color="muted">
            {emptyText}
          </Text>
        ) : null}
      </Command.List>
      <Box className="px-3">
        <ViewsDirectoryStatus directory={directory} />
      </Box>
    </Command>
  );
};

export const ViewsSwitcher = ({
  selected,
  onSelect,
  loading = false,
  triggerRef,
}: {
  selected: SavedView | null;
  onSelect: (view: SavedView) => void;
  loading?: boolean;
  triggerRef?: RefObject<HTMLButtonElement | null>;
}) => {
  const [open, setOpen] = useState(false);
  const { data: session } = useSession();
  const { workspaceSlug } = useWorkspacePath();
  return (
    <Popover onOpenChange={setOpen} open={open}>
      <BreadCrumbs
        breadCrumbs={[
          { name: "Views", icon: <ViewsIcon />, className: "shrink-0" },
          {
            name:
              selected?.name ?? (loading ? "Loading view..." : "Choose view"),
            className: "min-w-0",
            content: (
              <Box
                aria-label={
                  selected?.name ??
                  (loading ? "Loading view..." : "Choose view")
                }
                as="h1"
                className="min-w-0"
              >
                <Popover.Trigger asChild>
                  <Button
                    aria-label={`Switch view${selected ? `: ${selected.name}` : ""}`}
                    className="max-w-72 min-w-0 justify-start px-1.5"
                    color="tertiary"
                    data-view-menu-trigger
                    ref={triggerRef}
                    rightIcon={
                      <ChevronRightIcon
                        className="h-3.5 w-auto shrink-0 rotate-90"
                        strokeWidth={3.5}
                      />
                    }
                    size="sm"
                    variant="naked"
                  >
                    {loading ? (
                      <Flex align="center" aria-hidden as="span" gap={2}>
                        <Skeleton as="span" className="size-4.5 rounded-full" />
                        <Skeleton as="span" className="h-4 w-32" />
                      </Flex>
                    ) : (
                      <>
                        {selected ? (
                          <SavedViewIcon
                            className="h-4.5 w-auto shrink-0"
                            configuration={selected.configuration}
                          />
                        ) : (
                          <ViewsIcon className="h-4.5 w-auto shrink-0" />
                        )}
                        <Text
                          as="span"
                          className="min-w-0 truncate"
                          title={selected?.name}
                        >
                          {selected?.name ?? "Choose view"}
                        </Text>
                      </>
                    )}
                  </Button>
                </Popover.Trigger>
              </Box>
            ),
          },
        ]}
        className="min-w-0"
      />
      <Popover.Content
        align="start"
        className="w-88 max-w-[var(--radix-popover-content-available-width)] px-1"
      >
        <ViewsPickerContent
          key={`${workspaceSlug}:${session?.user.id ?? ""}`}
          onSelect={(view) => {
            onSelect(view);
            setOpen(false);
          }}
          selected={selected}
        />
      </Popover.Content>
    </Popover>
  );
};
