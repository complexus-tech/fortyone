"use client";
import { Button } from "ui";
import { ArrowDownIcon, FilterIcon } from "icons";
import { useState } from "react";
import { useHotkeys } from "react-hotkeys-hook";
import type { StoriesFilter } from "./stories-filter-types";
import type { StoriesFilterField } from "./stories-filter-bar/types";
import {
  getActiveStoriesFilterCount,
  hasActiveStoriesFilters,
} from "./stories-filter-utils";
import { StoriesFilterMenu } from "./stories-filter-bar";

type StoriesFilterButtonProps = {
  filters: StoriesFilter;
  setFilters: (v: StoriesFilter) => void;
  resetFilters: () => void;
  iconOnly?: boolean;
  hiddenFields?: readonly StoriesFilterField[];
};

export const StoriesFilterButton = ({
  filters,
  setFilters,
  iconOnly = false,
  hiddenFields = [],
}: StoriesFilterButtonProps) => {
  const [open, setOpen] = useState(false);
  const visibleFilters = { ...filters };
  hiddenFields.forEach((field) => {
    if (field === "assignedToMe" || field === "createdByMe") {
      visibleFilters[field] = false;
      return;
    }

    visibleFilters[field] = null;
  });
  const filtersCount = getActiveStoriesFilterCount(visibleFilters);
  const getButtonLabel = () => {
    if (filtersCount) {
      return `${filtersCount} filter${filtersCount > 1 ? "s" : ""} applied`;
    }
    return "Filters";
  };

  useHotkeys("v+f", (e) => {
    e.preventDefault();
    setOpen((current) => !current);
  });

  return (
    <StoriesFilterMenu
      align="end"
      filters={filters}
      hiddenFields={hiddenFields}
      onOpenChange={setOpen}
      open={open}
      setFilters={setFilters}
    >
      <Button
        aria-label={getButtonLabel()}
        className="relative"
        color="tertiary"
        leftIcon={<FilterIcon className="h-4 w-auto" />}
        rightIcon={
          iconOnly ? undefined : <ArrowDownIcon className="h-3.5 w-auto" />
        }
        size="sm"
        variant="outline"
      >
        {hasActiveStoriesFilters(visibleFilters) ? (
          <span
            aria-hidden="true"
            className="bg-primary absolute -top-0.5 -right-0.5 h-2.5 w-2.5 rounded-full"
          >
            <span className="bg-primary absolute inset-0 animate-ping rounded-full opacity-75" />
          </span>
        ) : null}
        {iconOnly ? null : (
          <span className="hidden md:inline">{getButtonLabel()}</span>
        )}
      </Button>
    </StoriesFilterMenu>
  );
};
