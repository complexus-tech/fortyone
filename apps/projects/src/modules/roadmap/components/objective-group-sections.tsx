"use client";

import {
  ArrowDownIcon,
  MoreHorizontalIcon,
  ObjectiveIcon,
  PlusIcon,
} from "icons";
import { cn } from "lib";
import {
  Avatar,
  Box,
  Button,
  Checkbox,
  Flex,
  Popover,
  Text,
  Tooltip,
} from "ui";
import { ObjectiveStatusIcon } from "@/components/ui/objective-status-icon";
import { PriorityIcon } from "@/components/ui/priority-icon";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import type {
  ObjectiveGroup,
  ObjectiveViewOptions,
} from "../objective-board-utils";

const ObjectiveGroupIdentity = ({
  group,
  groupBy,
}: {
  group: ObjectiveGroup;
  groupBy: ObjectiveViewOptions["groupBy"];
}) => {
  if (groupBy === "status") {
    return (
      <>
        <ObjectiveStatusIcon statusId={group.status?.id} />
        <Text className="max-w-[20ch] truncate">
          {group.status?.name ?? "Unknown status"}
        </Text>
      </>
    );
  }

  if (groupBy === "priority") {
    return (
      <>
        <PriorityIcon priority={group.priority} />
        <Text>{group.priority ?? "No Priority"}</Text>
      </>
    );
  }

  return (
    <>
      <Avatar
        name={group.member?.fullName || group.member?.username}
        size="xs"
        src={group.member?.avatarUrl}
      />
      <Text className="max-w-[20ch] truncate">
        {group.member?.username ?? "Unassigned"}
      </Text>
    </>
  );
};

export const ObjectiveGroupHeader = ({
  group,
  groupBy,
  onCreateObjective,
  collapsible = false,
  isCollapsed = false,
  onToggle,
  onHide,
  selectedObjectives,
  setSelectedObjectives,
}: {
  group: ObjectiveGroup;
  groupBy: ObjectiveViewOptions["groupBy"];
  onCreateObjective: () => void;
  collapsible?: boolean;
  isCollapsed?: boolean;
  onToggle?: () => void;
  onHide?: () => void;
  selectedObjectives?: string[];
  setSelectedObjectives?: (objectiveIds: string[]) => void;
}) => {
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const groupedObjectiveIds = group.objectives.map(({ id }) => id);
  const groupedObjectiveIdSet = new Set(groupedObjectiveIds);
  const selectedObjectiveIdSet = new Set(selectedObjectives ?? []);

  return (
    <Flex align="center" gap={2} justify="between">
      <Flex align="center" className="relative min-w-0" gap={2}>
        {selectedObjectives && setSelectedObjectives ? (
          <Checkbox
            checked={
              groupedObjectiveIds.length > 0 &&
              groupedObjectiveIds.every((id) => selectedObjectiveIdSet.has(id))
            }
            className="absolute -left-[1.6rem] hidden rounded md:inline"
            disabled={userRole === "guest"}
            onCheckedChange={(checked) => {
              if (checked) {
                setSelectedObjectives(
                  Array.from(
                    new Set([...selectedObjectives, ...groupedObjectiveIds]),
                  ),
                );
              } else {
                setSelectedObjectives(
                  selectedObjectives.filter(
                    (id) => !groupedObjectiveIdSet.has(id),
                  ),
                );
              }
            }}
          />
        ) : null}
        <button
          className="focus-visible:ring-primary flex min-w-0 items-center gap-2 rounded-sm outline-none focus-visible:ring-1 disabled:cursor-default"
          disabled={!collapsible}
          onClick={onToggle}
          type="button"
        >
          <ObjectiveGroupIdentity group={group} groupBy={groupBy} />
          {collapsible ? (
            <ArrowDownIcon
              className={cn("text-text-muted h-4 w-auto transition", {
                "-rotate-90": isCollapsed,
              })}
              strokeWidth={1}
            />
          ) : null}
        </button>
        <Tooltip
          title={`Total ${getTermDisplay("objectiveTerm", { variant: "plural" })}`}
        >
          <span>
            <ObjectiveIcon className="text-text-muted h-5 w-auto" />
          </span>
        </Tooltip>
        <Text color="muted">
          {group.objectives.length}{" "}
          {getTermDisplay("objectiveTerm", {
            variant: group.objectives.length === 1 ? "singular" : "plural",
          })}
        </Text>
      </Flex>
      <Flex align="center" gap={1}>
        {onHide ? (
          <Popover>
            <Popover.Trigger asChild>
              <Button
                aria-label="Column options"
                color="tertiary"
                size="sm"
                variant="naked"
              >
                <MoreHorizontalIcon
                  className="h-[1.15rem] w-auto"
                  strokeWidth={4}
                />
              </Button>
            </Popover.Trigger>
            <Popover.Content align="end" className="w-44 p-1.5">
              <Button
                className="justify-start px-2"
                color="tertiary"
                fullWidth
                onClick={onHide}
                size="sm"
                variant="naked"
              >
                Hide column
              </Button>
            </Popover.Content>
          </Popover>
        ) : null}
        <Button
          aria-label={`New ${getTermDisplay("objectiveTerm")}`}
          color="tertiary"
          disabled={userRole === "guest"}
          onClick={onCreateObjective}
          size="sm"
          variant="naked"
        >
          <PlusIcon className="h-[1.2rem] w-auto" />
        </Button>
      </Flex>
    </Flex>
  );
};

export const HiddenObjectiveGroups = ({
  groups,
  groupBy,
  onShow,
}: {
  groups: ObjectiveGroup[];
  groupBy: ObjectiveViewOptions["groupBy"];
  onShow: (groupKey: string) => void;
}) => {
  if (groups.length === 0) return null;

  return (
    <Box className="w-[340px] shrink-0">
      <Flex direction="column" gap={3}>
        {groups.map((group) => (
          <Box
            className="border-border bg-surface hover:bg-surface-elevated group dark:border-border/70 flex min-h-14 cursor-pointer items-center justify-between rounded-xl border-[0.5px] px-4 transition duration-200 ease-linear select-none"
            key={group.key}
            onClick={() => {
              onShow(group.key);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter" || event.key === " ") {
                event.preventDefault();
                onShow(group.key);
              }
            }}
            role="button"
            tabIndex={0}
          >
            <Flex align="center" className="min-w-0" gap={2}>
              <ObjectiveGroupIdentity group={group} groupBy={groupBy} />
            </Flex>
            <Flex align="center" className="shrink-0" gap={1}>
              <Popover>
                <Popover.Trigger asChild>
                  <Button
                    aria-label="Hidden column options"
                    className="opacity-0 transition group-hover:opacity-100 focus-visible:opacity-100"
                    color="tertiary"
                    onClick={(event) => {
                      event.stopPropagation();
                    }}
                    size="sm"
                    variant="naked"
                  >
                    <MoreHorizontalIcon
                      className="h-[1.15rem] w-auto"
                      strokeWidth={4}
                    />
                  </Button>
                </Popover.Trigger>
                <Popover.Content align="end" className="w-40 p-1.5">
                  <Button
                    className="justify-start px-2"
                    color="tertiary"
                    fullWidth
                    onClick={() => {
                      onShow(group.key);
                    }}
                    size="sm"
                    variant="naked"
                  >
                    Show column
                  </Button>
                </Popover.Content>
              </Popover>
              <Text color="muted">{group.objectives.length}</Text>
            </Flex>
          </Box>
        ))}
      </Flex>
    </Box>
  );
};
