"use client";
import type { ReactElement, ReactNode } from "react";
import { useState } from "react";
import { useParams } from "next/navigation";
import { Box, Flex, Menu, Text } from "ui";
import { ChevronRightIcon } from "icons";
import { useTerminology } from "@/hooks/use-terminology-display";
import { buildFilterOptions } from "./stories-filter-bar/filter-options";
import { TitleFilterDialog } from "./stories-filter-bar/filter-chip";
import {
  EMPTY_FILTER_FIELDS,
  getEditorContentClassName,
} from "./stories-filter-bar/filter-model";
import type {
  StoriesFilterEditorProps,
  StoriesFilterField,
} from "./stories-filter-bar/types";

export const StoriesFilterMenu = ({
  children,
  filters,
  setFilters,
  hiddenFields = EMPTY_FILTER_FIELDS,
  align = "start",
  open,
  onOpenChange,
  renderEditor,
}: StoriesFilterEditorProps & {
  children: ReactElement;
  renderEditor: (field: StoriesFilterField) => ReactNode;
  hiddenFields?: readonly StoriesFilterField[];
  align?: "start" | "end";
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
}) => {
  const { teamId } = useParams<{ teamId?: string }>();
  const { getTermDisplay } = useTerminology();
  const [titleDialogOpen, setTitleDialogOpen] = useState(false);
  const [query, setQuery] = useState("");
  const filterOptions = buildFilterOptions({
    filters,
    getTermDisplay,
    hasRouteTeam: Boolean(teamId),
    hiddenFields: new Set(hiddenFields),
  });
  return (
    <>
      <Menu
        onOpenChange={(nextOpen) => {
          if (!nextOpen) setQuery("");
          onOpenChange?.(nextOpen);
        }}
        open={open}
      >
        <Menu.Button>{children}</Menu.Button>
        <Menu.Items align={align} className="w-80 py-1">
          <Box className="px-4 py-2">
            <Menu.Input
              autoFocus
              onChange={(event) => {
                setQuery(event.target.value);
              }}
              onKeyDown={(event) => {
                if (event.key !== "Escape" && event.key !== "Tab")
                  event.stopPropagation();
              }}
              placeholder="Add filter..."
              value={query}
            />
          </Box>
          <Menu.Separator className="my-0" />
          <Menu.Group className="max-h-[min(40rem,calc(var(--radix-dropdown-menu-content-available-height)-4rem))] overflow-y-auto px-1 py-1.5">
            {filterOptions
              .filter((option) =>
                option.label.toLowerCase().includes(query.trim().toLowerCase()),
              )
              .map((option) => {
                const value = filters[option.field];
                const isActive = Array.isArray(value)
                  ? value.length > 0
                  : Boolean(value);
                if (option.field === "contentContains") {
                  return (
                    <Menu.Item
                      active={isActive}
                      className="justify-between gap-4"
                      key={option.field}
                      onSelect={() => {
                        setTitleDialogOpen(true);
                      }}
                    >
                      <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                        <span className="text-text-secondary flex h-6 w-6 shrink-0 items-center">
                          {option.icon}
                        </span>
                        <Text className="truncate">{option.label}</Text>
                      </Box>
                    </Menu.Item>
                  );
                }

                return (
                  <Menu.SubMenu key={option.field}>
                    <Menu.SubTrigger
                      active={isActive}
                      className="justify-between gap-4"
                    >
                      <Box className="grid min-w-0 flex-1 grid-cols-[24px_minmax(0,1fr)] items-center">
                        <span className="text-text-secondary flex h-6 w-6 shrink-0 items-center">
                          {option.icon}
                        </span>
                        <Text className="truncate">{option.label}</Text>
                      </Box>
                      <Flex align="center" className="shrink-0" gap={1}>
                        <ChevronRightIcon
                          className="text-text-muted h-3.5 w-auto"
                          strokeWidth={2.8}
                        />
                      </Flex>
                    </Menu.SubTrigger>
                    <Menu.SubItems
                      alignOffset={-6}
                      className={getEditorContentClassName(option.field)}
                      sideOffset={8}
                    >
                      {renderEditor(option.field)}
                    </Menu.SubItems>
                  </Menu.SubMenu>
                );
              })}
          </Menu.Group>
        </Menu.Items>
      </Menu>
      {titleDialogOpen ? (
        <TitleFilterDialog
          filters={filters}
          key={filters.contentContains ?? ""}
          onOpenChange={setTitleDialogOpen}
          open={titleDialogOpen}
          setFilters={setFilters}
        />
      ) : null}
    </>
  );
};
