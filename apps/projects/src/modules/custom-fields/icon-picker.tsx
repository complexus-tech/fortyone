"use client";

import type { KeyboardEvent } from "react";
import { useEffect, useId, useRef, useState } from "react";
import { SearchIcon } from "icons";
import { Box, Button, commandFilter, Input, Popover, Text, Tooltip } from "ui";
import type { CustomFieldIconKey, CustomFieldType } from "./types";
import { CustomFieldIcon, customFieldIconOptions } from "./icons";

const GRID_COLUMNS = 8;
const GRID_OFFSETS: Partial<Record<string, number>> = {
  ArrowDown: GRID_COLUMNS,
  ArrowUp: -GRID_COLUMNS,
  ArrowRight: 1,
  ArrowLeft: -1,
};

type IconChoice = {
  value: CustomFieldIconKey | "automatic";
  label: string;
  keywords: readonly string[];
};

export const CustomFieldIconPicker = ({
  value,
  type,
  onChange,
  disabled = false,
}: {
  value: CustomFieldIconKey | null;
  type: CustomFieldType;
  onChange: (icon: CustomFieldIconKey | null) => void;
  disabled?: boolean;
}) => {
  const [open, setOpen] = useState(false);
  const [search, setSearch] = useState("");
  const [activeIcon, setActiveIcon] = useState<IconChoice["value"]>(
    value ?? "automatic",
  );
  const listId = useId();
  const activeOptionRef = useRef<HTMLButtonElement>(null);
  const selected = customFieldIconOptions.find(
    (option) => option.value === value,
  );
  const choices: IconChoice[] = [
    {
      value: "automatic",
      label: "Automatic",
      keywords: ["default", "reset", type],
    },
    ...customFieldIconOptions,
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

  const choose = (icon: CustomFieldIconKey | null) => {
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
    "h-auto w-full min-w-0 rounded-md p-0 ring-ring aria-selected:bg-state-hover focus-visible:ring-2 data-[chosen=true]:bg-state-selected data-[chosen=true]:ring-1";

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
            aria-label={`Field icon: ${selected?.label ?? "Automatic"}`}
            asIcon
            className="h-[2.8rem] w-[2.8rem]"
            color="tertiary"
            disabled={disabled}
            type="button"
            variant="outline"
          >
            <CustomFieldIcon
              className="h-[1.15rem] w-[1.15rem] shrink-0"
              field={{ type, icon: value }}
            />
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          aria-label="Choose field icon"
          className="flex max-h-[var(--radix-popover-content-available-height)] w-80 max-w-[calc(100vw-2rem)] flex-col overflow-hidden p-0"
        >
          <Box className="min-h-0 flex-1">
            <Input
              aria-activedescendant={
                activeChoice ? optionId(activeChoice.value) : undefined
              }
              aria-autocomplete="list"
              aria-controls={listId}
              aria-expanded={open}
              aria-label="Search field icons"
              autoFocus
              className="h-11 rounded-none border-0 bg-transparent px-3 text-base focus-visible:ring-inset dark:bg-transparent"
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
              aria-label="Field icons"
              className="grid max-h-80 min-h-0 grid-cols-8 gap-1 overflow-y-auto p-2"
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
                    <CustomFieldIcon
                      className="h-[1.15rem] w-[1.15rem] shrink-0"
                      field={{
                        type,
                        icon:
                          choice.value === "automatic" ? null : choice.value,
                      }}
                    />
                  </Button>
                </Tooltip>
              ))}
              {!visibleChoices.length && (
                <Text
                  className="col-span-8 px-3 py-4 text-center"
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
