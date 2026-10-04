"use client";

import type { KeyboardEvent } from "react";
import { useEffect, useId, useRef, useState } from "react";
import { SearchIcon } from "icons";
import { Box, Button, commandFilter, Input, Popover, Text, Tooltip } from "ui";
import type { ViewIconKey } from "@/shared/views/metadata";
import { SavedViewIcon } from "@/shared/views/icons";
import { VIEW_ICON_OPTIONS } from "@/shared/views/icon-options";

const GRID_COLUMNS = 5;
const GRID_OFFSETS: Partial<Record<string, number>> = {
  ArrowDown: GRID_COLUMNS,
  ArrowUp: -GRID_COLUMNS,
  ArrowRight: 1,
  ArrowLeft: -1,
};

type IconChoice = {
  value: ViewIconKey | "automatic";
  label: string;
  keywords: readonly string[];
};

export const ViewIconPicker = ({
  value,
  onChange,
  disabled = false,
}: {
  value: ViewIconKey | null;
  onChange: (icon: ViewIconKey | null) => void;
  disabled?: boolean;
}) => {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [activeIcon, setActiveIcon] = useState<IconChoice["value"]>(
    value ?? "automatic",
  );
  const listId = useId();
  const activeOptionRef = useRef<HTMLButtonElement>(null);
  const selected = VIEW_ICON_OPTIONS.find((option) => option.value === value);
  const choices: IconChoice[] = [
    {
      value: "automatic",
      label: "Automatic",
      keywords: ["default", "reset"],
    },
    ...VIEW_ICON_OPTIONS.map((option) => ({ ...option, keywords: [] })),
  ];
  const query = search.trim();
  const visibleChoices = query
    ? choices
        .map((choice) => ({
          choice,
          score: commandFilter(choice.value, query, [
            choice.label,
            ...choice.keywords,
          ]),
        }))
        .filter(({ score }) => score > 0)
        .sort((left, right) => right.score - left.score)
        .map(({ choice }) => choice)
    : choices;
  const activeChoice =
    visibleChoices.find((choice) => choice.value === activeIcon) ??
    visibleChoices.at(0);
  const optionId = (icon: IconChoice["value"]) => `${listId}-${icon}`;

  useEffect(() => {
    if (open) activeOptionRef.current?.scrollIntoView({ block: "nearest" });
  }, [activeChoice?.value, open]);

  const choose = (icon: ViewIconKey | null) => {
    if (disabled) return;
    onChange(icon);
    setOpen(false);
  };

  const navigateGrid = (event: KeyboardEvent<HTMLInputElement>) => {
    if (
      event.altKey ||
      event.ctrlKey ||
      event.metaKey ||
      event.nativeEvent.isComposing
    )
      return;
    if (event.key === "Enter") {
      event.preventDefault();
      event.stopPropagation();
      if (activeChoice)
        choose(activeChoice.value === "automatic" ? null : activeChoice.value);
      return;
    }
    // Leave horizontal arrows available for editing a search query.
    if (
      search &&
      ["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)
    )
      return;
    const offset = GRID_OFFSETS[event.key];
    if (offset === undefined && !["Home", "End"].includes(event.key)) return;
    event.preventDefault();
    event.stopPropagation();
    if (!visibleChoices.length) return;
    const currentIndex = Math.max(
      0,
      visibleChoices.findIndex(
        (choice) => choice.value === activeChoice?.value,
      ),
    );
    let nextIndex = currentIndex;
    if (event.key === "Home") nextIndex = 0;
    else if (event.key === "End") nextIndex = visibleChoices.length - 1;
    else if (offset !== undefined)
      nextIndex = Math.min(
        visibleChoices.length - 1,
        Math.max(0, currentIndex + offset),
      );
    setActiveIcon(visibleChoices[nextIndex].value);
  };

  const itemClassName =
    "h-auto w-full min-w-0 rounded-md p-0 ring-ring aria-selected:bg-state-hover focus-visible:ring-2 data-[chosen=true]:bg-state-selected";

  return (
    <Box className="w-[2.8rem] shrink-0">
      <Popover
        onOpenChange={(nextOpen) => {
          if (nextOpen && disabled) return;
          if (nextOpen) {
            setSearch("");
            setActiveIcon(value ?? "automatic");
          }
          setOpen(nextOpen);
        }}
        open={open}
      >
        <Popover.Trigger asChild>
          <Button
            aria-label={`View icon: ${selected?.label ?? "Automatic"}`}
            asIcon
            className="h-[2.8rem] w-[2.8rem] md:h-[2.8rem]"
            color="tertiary"
            disabled={disabled}
            type="button"
            variant="outline"
          >
            <SavedViewIcon
              className="h-[1.15rem] w-[1.15rem] shrink-0"
              configuration={{ icon: value }}
            />
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          aria-label="Choose view icon"
          className="flex max-h-[var(--radix-popover-content-available-height)] w-80 max-w-[calc(100vw-2rem)] flex-col overflow-hidden p-2"
        >
          <Box className="flex min-h-0 flex-1 flex-col gap-2">
            <Input
              aria-activedescendant={
                activeChoice ? optionId(activeChoice.value) : undefined
              }
              aria-autocomplete="list"
              aria-controls={listId}
              aria-expanded={open}
              aria-label="Search view icons"
              autoFocus
              className="rounded-lg px-3 text-base"
              leftIcon={
                <SearchIcon className="h-[1.15rem] w-[1.15rem] opacity-60" />
              }
              onChange={(event) => {
                setSearch(event.target.value);
                setActiveIcon("automatic");
              }}
              onKeyDown={navigateGrid}
              placeholder="Search icons…"
              role="combobox"
              value={search}
            />
            <Box
              aria-label="View icons"
              className="grid min-h-0 flex-1 grid-cols-5 gap-1 overflow-y-auto"
              id={listId}
              role="listbox"
            >
              {visibleChoices.map((choice) => (
                <Tooltip key={choice.value} title={choice.label}>
                  <Button
                    aria-label={choice.label}
                    aria-selected={activeChoice?.value === choice.value}
                    asIcon
                    className={itemClassName}
                    color="tertiary"
                    data-chosen={(value ?? "automatic") === choice.value}
                    disabled={disabled}
                    id={optionId(choice.value)}
                    onClick={() => {
                      choose(
                        choice.value === "automatic" ? null : choice.value,
                      );
                    }}
                    onMouseDown={(event) => {
                      event.preventDefault();
                    }}
                    onPointerMove={() => {
                      setActiveIcon(choice.value);
                    }}
                    ref={
                      activeChoice?.value === choice.value
                        ? activeOptionRef
                        : undefined
                    }
                    role="option"
                    size="sm"
                    tabIndex={-1}
                    type="button"
                    variant="naked"
                  >
                    <SavedViewIcon
                      className="h-[1.15rem] w-[1.15rem] shrink-0"
                      configuration={{
                        icon:
                          choice.value === "automatic" ? null : choice.value,
                      }}
                    />
                  </Button>
                </Tooltip>
              ))}
              {!visibleChoices.length && (
                <Text
                  className="col-span-5 px-3 py-4 text-center"
                  role="status"
                >
                  No icons found.
                </Text>
              )}
            </Box>
          </Box>
        </Popover.Content>
      </Popover>
    </Box>
  );
};
