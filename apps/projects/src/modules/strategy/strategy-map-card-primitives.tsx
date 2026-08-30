"use client";

import type { ReactNode } from "react";
import { HelpIcon } from "icons";
import { cn } from "lib";
import { Box, Flex, Text, Tooltip } from "ui";

export const getStrategyDescriptionPreview = (description: string | null) => {
  if (!description) return "";
  if (!/<\/?[a-z][\s\S]*>/i.test(description)) return description.trim();

  return description
    .replace(/<[^>]*>/g, " ")
    .replace(/&nbsp;/g, " ")
    .replace(/&amp;/g, "&")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .replace(/&#39;/g, "'")
    .replace(/&quot;/g, '"')
    .replace(/\s+/g, " ")
    .trim();
};

const getProgressFillClassName = (progress: number) => {
  if (progress < 25) return "bg-danger";
  if (progress < 50) return "bg-warning";
  if (progress < 75) return "bg-info";
  return "bg-success";
};

export const StrategyProgressBar = ({
  className,
  progress,
}: {
  className?: string;
  progress: number;
}) => (
  <Flex align="center" className={cn("gap-2.5", className)}>
    <div
      aria-label={`Progress ${progress}%`}
      aria-valuemax={100}
      aria-valuemin={0}
      aria-valuenow={progress}
      className="bg-border h-1.5 min-w-0 flex-1 overflow-hidden rounded-full"
      role="progressbar"
    >
      <div
        className={cn(
          "h-full rounded-full",
          getProgressFillClassName(progress),
        )}
        style={{ width: `${progress}%` }}
      />
    </div>
    <Text className="shrink-0 text-sm tabular-nums" color="muted">
      {progress}%
    </Text>
  </Flex>
);

export const cardClasses = cn(
  "border-border-strong/65 bg-white shadow-shadow dark:border-foreground/20 dark:bg-accent/70",
  "rounded-xl border-2 shadow-lg backdrop-blur",
  "transition-[border-color,box-shadow,background-color] duration-150",
  "hover:border-foreground/35 hover:bg-surface-elevated hover:shadow-xl dark:hover:border-foreground/45 dark:hover:bg-accent/70",
  "group-data-[dragging=true]/node:border-foreground/65 group-data-[dragging=true]/node:shadow-2xl",
);

export const objectivePropertyControlClasses =
  "dark:border-foreground/15 dark:bg-state-hover! dark:hover:bg-state-active!";

export const NodeEyebrow = ({
  children,
  className,
}: {
  children: ReactNode;
  className?: string;
}) => (
  <Text
    className={cn(
      "text-[0.68rem] font-medium tracking-[0.12em] uppercase",
      className,
    )}
    color="muted"
  >
    {children}
  </Text>
);

export const ContextMenuLabel = ({
  children,
  icon,
}: {
  children: ReactNode;
  icon: ReactNode;
}) => (
  <Flex align="center" className="w-full gap-2">
    <span className="grid h-4 w-4 place-items-center text-current">{icon}</span>
    <Text>{children}</Text>
  </Flex>
);

export const StrategyConceptInfo = ({
  description,
  label,
}: {
  description: string;
  label: string;
}) => (
  <Tooltip
    className="border-border-strong dark:border-border-strong dark:bg-surface-elevated max-w-72 py-2.5"
    delayDuration={250}
    title={description}
  >
    <button
      aria-label={`About ${label}`}
      className="text-text-muted hover:text-foreground inline-grid shrink-0 place-items-center rounded-sm transition-colors"
      data-no-drag
      onClick={(event) => {
        event.stopPropagation();
      }}
      type="button"
    >
      <HelpIcon className="h-4 w-4 text-current" />
    </button>
  </Tooltip>
);

export const Metric = ({
  label,
  value,
}: {
  label: string;
  value: string | number;
}) => (
  <Box className="min-w-0 flex-1 text-center">
    <Text className="text-xl tabular-nums" fontWeight="semibold">
      {value}
    </Text>
    <Text
      className="mt-0.5 truncate text-[0.75rem] tracking-[0.1em] uppercase"
      color="muted"
    >
      {label}
    </Text>
  </Box>
);
