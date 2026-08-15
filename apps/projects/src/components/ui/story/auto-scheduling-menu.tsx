"use client";

import {
  createContext,
  useContext,
  useId,
  useState,
  type ReactNode,
} from "react";
import {
  FileUnlockedIcon,
  LockIcon,
  PauseIcon,
  Time02Icon,
  TimeScheduleIcon,
  UsersAddIcon,
} from "icons";
import { cn } from "lib";
import { Box, Button, Divider, Flex, Popover, Switch, Text, TimeAgo } from "ui";
import {
  canToggleAutoSchedulingLock,
  deriveAutoSchedulingStatus,
  getAutoSchedulingHelper,
  getAutoSchedulingLabel,
} from "@/lib/auto-scheduling";
import type { AutoSchedulingStatus } from "@/modules/stories/types";

export type AutoSchedulingPatch = {
  autoSchedulingEnabled?: boolean;
  autoSchedulingLocked?: boolean;
};

const AutoSchedulingContext = createContext<{
  setOpen: (open: boolean) => void;
}>({ setOpen: () => {} });

export const AutoSchedulingMenu = ({ children }: { children: ReactNode }) => {
  const [open, setOpen] = useState(false);

  return (
    <AutoSchedulingContext.Provider value={{ setOpen }}>
      <Popover onOpenChange={setOpen} open={open}>
        {children}
      </Popover>
    </AutoSchedulingContext.Provider>
  );
};

const Trigger = ({ children }: { children: ReactNode }) => (
  <Popover.Trigger asChild>{children}</Popover.Trigger>
);

const STATUS_DOT_CLASSES: Record<AutoSchedulingStatus, string> = {
  off: "bg-text-muted/60",
  needs_owner: "bg-warning",
  needs_time: "bg-warning",
  planning: "bg-info",
  scheduled: "bg-success",
  at_risk: "bg-warning",
  cannot_fit: "bg-danger",
  locked: "bg-primary",
};

export const AutoSchedulingStatusDot = ({
  className,
  status,
}: {
  className?: string;
  status: AutoSchedulingStatus;
}) => (
  <span
    aria-hidden
    className={cn(
      "inline-block size-2 shrink-0 rounded-full",
      STATUS_DOT_CLASSES[status],
      className,
    )}
  />
);

const Items = ({
  align = "center",
  assigneeId,
  autoSchedulingEnabled,
  autoSchedulingLocked,
  autoSchedulingRequired = false,
  autoSchedulingReason,
  autoSchedulingStatus,
  autoSchedulingUpdatedAt,
  canManage = true,
  estimatedDurationMinutes,
  onChange,
  onRequestOwner,
  onRequestTime,
}: {
  align?: "start" | "end" | "center";
  assigneeId?: string | null;
  autoSchedulingEnabled: boolean;
  autoSchedulingLocked: boolean;
  autoSchedulingRequired?: boolean;
  autoSchedulingReason?: string | null;
  autoSchedulingStatus?: AutoSchedulingStatus | null;
  autoSchedulingUpdatedAt?: string | null;
  canManage?: boolean;
  estimatedDurationMinutes?: number | null;
  onChange: (patch: AutoSchedulingPatch) => void;
  onRequestOwner?: () => void;
  onRequestTime?: () => void;
}) => {
  const { setOpen } = useContext(AutoSchedulingContext);
  const enabledSwitchId = useId();
  const lockedSwitchId = useId();
  const status = deriveAutoSchedulingStatus({
    assigneeId,
    autoSchedulingEnabled,
    autoSchedulingLocked,
    autoSchedulingStatus,
    estimatedDurationMinutes,
  });
  const helper = canManage
    ? getAutoSchedulingHelper(status, autoSchedulingReason)
    : "Auto-scheduling is available on a paid plan.";
  const canToggleLock = canToggleAutoSchedulingLock(
    status,
    autoSchedulingLocked,
  );
  const updatedAt = autoSchedulingUpdatedAt
    ? new Date(autoSchedulingUpdatedAt)
    : null;
  const hasValidUpdatedAt = updatedAt && !Number.isNaN(updatedAt.getTime());

  const requestOwner = () => {
    setOpen(false);
    onRequestOwner?.();
  };
  const requestTime = () => {
    setOpen(false);
    onRequestTime?.();
  };

  return (
    <Popover.Content align={align} className="w-80 p-3">
      <Flex align="start" className="gap-2.5">
        <Box className="bg-surface-muted mt-0.5 flex size-8 shrink-0 items-center justify-center rounded-lg">
          <TimeScheduleIcon aria-hidden className="text-text-muted h-4.5" />
        </Box>
        <Box className="min-w-0 flex-1">
          <Flex align="center" gap={2}>
            <Text fontWeight="medium">Auto-scheduling</Text>
            <Flex align="center" className="gap-1.5" wrap={false}>
              <AutoSchedulingStatusDot status={status} />
              <Text className="text-xs" color="muted">
                {getAutoSchedulingLabel(status)}
              </Text>
            </Flex>
          </Flex>
          <Text className="mt-1 text-xs leading-5" color="muted">
            {helper}
          </Text>
          {hasValidUpdatedAt ? (
            <Text className="mt-1 text-[0.7rem]" color="muted">
              Updated <TimeAgo timestamp={autoSchedulingUpdatedAt!} />
            </Text>
          ) : null}
        </Box>
      </Flex>

      <Divider className="my-3" />

      <label
        className={cn(
          "flex items-center justify-between gap-4",
          autoSchedulingRequired ? "cursor-not-allowed" : "cursor-pointer",
        )}
        htmlFor={enabledSwitchId}
      >
        <Box>
          <Text fontSize="md" fontWeight="medium">
            Let Maya schedule it
          </Text>
          <Text className="mt-0.5 text-xs" color="muted">
            {autoSchedulingRequired
              ? "Selecting Maya requires auto-scheduling so it can assign and plan the work."
              : "Uses the assignee's availability and time needed."}
          </Text>
        </Box>
        <Switch
          aria-label="Enable auto-scheduling"
          checked={autoSchedulingEnabled}
          disabled={
            !canManage || autoSchedulingLocked || autoSchedulingRequired
          }
          id={enabledSwitchId}
          onCheckedChange={(checked) => {
            onChange({ autoSchedulingEnabled: checked });
          }}
        />
      </label>

      {autoSchedulingEnabled || autoSchedulingLocked ? (
        <>
          <Divider className="my-3" />
          <label
            className={cn(
              "flex items-center justify-between gap-4",
              canToggleLock
                ? "cursor-pointer"
                : "cursor-not-allowed opacity-60",
            )}
            htmlFor={lockedSwitchId}
          >
            <Box>
              <Text fontSize="md" fontWeight="medium">
                Lock current blocks
              </Text>
              <Text className="mt-0.5 text-xs" color="muted">
                {canToggleLock
                  ? "Keep Maya's current times fixed."
                  : "Available after Maya schedules this work."}
              </Text>
            </Box>
            <Switch
              aria-label="Lock auto-scheduled blocks"
              checked={autoSchedulingLocked}
              disabled={!canManage || !canToggleLock}
              id={lockedSwitchId}
              onCheckedChange={(checked) => {
                onChange({ autoSchedulingLocked: checked });
              }}
            />
          </label>
        </>
      ) : null}

      {canManage && !autoSchedulingRequired && status === "off" ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<TimeScheduleIcon className="h-4" />}
          onClick={() => {
            onChange({ autoSchedulingEnabled: true });
          }}
          size="sm"
          type="button"
          variant="outline"
        >
          Resume auto-scheduling
        </Button>
      ) : null}
      {canManage && status === "needs_owner" && onRequestOwner ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<UsersAddIcon className="h-4" />}
          onClick={requestOwner}
          size="sm"
          type="button"
          variant="outline"
        >
          Choose assignee
        </Button>
      ) : null}
      {canManage &&
      (status === "needs_time" ||
        status === "at_risk" ||
        status === "cannot_fit") &&
      onRequestTime ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<Time02Icon className="h-4" />}
          onClick={requestTime}
          size="sm"
          type="button"
          variant="outline"
        >
          {status === "needs_time" ? "Add time needed" : "Review time needed"}
        </Button>
      ) : null}
      {canManage && !autoSchedulingRequired && status === "planning" ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<PauseIcon className="h-4" />}
          onClick={() => {
            onChange({ autoSchedulingEnabled: false });
          }}
          size="sm"
          type="button"
          variant="outline"
        >
          Pause planning
        </Button>
      ) : null}
      {canManage && status === "scheduled" ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<LockIcon className="h-4" />}
          onClick={() => {
            onChange({ autoSchedulingLocked: true });
          }}
          size="sm"
          type="button"
          variant="outline"
        >
          Lock schedule
        </Button>
      ) : null}
      {canManage && status === "locked" ? (
        <Button
          className="mt-3 w-full justify-center"
          color="tertiary"
          leftIcon={<FileUnlockedIcon className="h-4" />}
          onClick={() => {
            onChange({ autoSchedulingLocked: false });
          }}
          size="sm"
          type="button"
          variant="outline"
        >
          Unlock schedule
        </Button>
      ) : null}
      {canManage &&
      !autoSchedulingRequired &&
      autoSchedulingEnabled &&
      status !== "planning" &&
      status !== "locked" ? (
        <Button
          className="mx-auto mt-1.5 px-2"
          color="tertiary"
          leftIcon={<PauseIcon className="h-4" />}
          onClick={() => {
            onChange({ autoSchedulingEnabled: false });
          }}
          size="xs"
          type="button"
          variant="naked"
        >
          Pause auto-scheduling
        </Button>
      ) : null}
    </Popover.Content>
  );
};

AutoSchedulingMenu.Trigger = Trigger;
AutoSchedulingMenu.Items = Items;
