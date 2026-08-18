"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { addDays, format, formatISO, startOfDay } from "date-fns";
import { cn } from "lib";
import {
  ArrowDownIcon,
  CalendarIcon,
  CheckIcon,
  DeleteIcon,
  EditIcon,
  ObjectiveIcon,
  PlusIcon,
} from "icons";
import { usePathname } from "next/navigation";
import {
  Avatar,
  Box,
  Button,
  Checkbox,
  CircleProgressBar,
  Container,
  ContextMenu,
  DatePicker,
  Flex,
  Input,
  Popover,
  ProgressBar,
  Text,
  Tooltip,
} from "ui";
import {
  AssigneesMenu,
  ConfirmDialog,
  RowWrapper,
  TeamColor,
} from "@/components/ui";
import { ContextMenuItem } from "@/components/ui/story/context-menu-item";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { useLocalStorage, useTerminology, useUserRole } from "@/hooks";
import { useDeleteKeyResultMutation } from "@/modules/objectives/hooks/use-delete-key-result-mutation";
import { useUpdateKeyResultMutation } from "@/modules/objectives/hooks/use-update-key-result-mutation";
import { NewKeyResultButton } from "@/modules/objectives/stories/overview/new-key-result";
import { UpdateKeyResultDialog } from "@/modules/objectives/stories/overview/update-key-result-dialog";
import type { Member } from "@/types";
import { hexToRgba } from "@/utils";
import type { KeyResultWithTeam } from "../types";
import type { ObjectiveKeyResultGroup } from "../utils";
import {
  formatKeyResultValue,
  getKeyResultProgress,
  getKeyResultReference,
} from "../utils";

const getDisplayName = (member?: Member) =>
  member?.fullName.trim() || member?.username || "No lead";

const KeyResultDateRange = ({
  canEdit,
  keyResult,
}: {
  canEdit: boolean;
  keyResult: KeyResultWithTeam;
}) => {
  const { getTermDisplay } = useTerminology();
  const { mutate: updateKeyResult } = useUpdateKeyResultMutation();
  const startDate = new Date(keyResult.startDate);
  const endDate = new Date(keyResult.endDate);
  const today = startOfDay(new Date());
  const normalizedEndDate = startOfDay(endDate);
  const isOverdue = normalizedEndDate < today;
  const isDueSoon =
    normalizedEndDate >= today && normalizedEndDate <= addDays(today, 7);
  const endDateColor = {
    "text-primary dark:text-primary": isOverdue,
    "text-warning dark:text-warning": isDueSoon,
  };

  const updateDate = (data: { startDate?: string; endDate?: string }) => {
    updateKeyResult({
      keyResultId: keyResult.id,
      objectiveId: keyResult.objectiveId,
      data,
      silent: true,
    });
  };

  return (
    <Flex align="center" className="shrink-0" gap={1}>
      <DatePicker>
        <Tooltip title={`Starts ${format(startDate, "MMM d, yyyy")}`}>
          <span>
            <DatePicker.Trigger>
              <Button
                className="px-2"
                color="tertiary"
                disabled={!canEdit}
                rounded="md"
                size="xs"
                type="button"
                variant="outline"
              >
                <CalendarIcon className="h-4" />
                {format(startDate, "MMM d")}
              </Button>
            </DatePicker.Trigger>
          </span>
        </Tooltip>
        <DatePicker.Calendar
          onDayClick={(day) => {
            updateDate({
              startDate: formatISO(day, { representation: "date" }),
            });
          }}
          selected={startDate}
        />
      </DatePicker>
      <Text className="text-text-muted px-0.5">–</Text>
      <DatePicker>
        <Tooltip
          className="py-3"
          title={
            <Flex align="start" gap={2}>
              <CalendarIcon
                className={cn("relative top-[2.5px] h-5 w-auto", endDateColor)}
              />
              <Box>
                {getDueDateMessage(endDate, getTermDisplay("keyResultTerm"))}
              </Box>
            </Flex>
          }
        >
          <span>
            <DatePicker.Trigger>
              <Button
                className={cn("px-2", endDateColor)}
                color="tertiary"
                disabled={!canEdit}
                rounded="md"
                size="xs"
                type="button"
                variant="outline"
              >
                <CalendarIcon className={cn("h-4", endDateColor)} />
                {format(endDate, "MMM d")}
              </Button>
            </DatePicker.Trigger>
          </span>
        </Tooltip>
        <DatePicker.Calendar
          onDayClick={(day) => {
            updateDate({
              endDate: formatISO(day, { representation: "date" }),
            });
          }}
          selected={endDate}
        />
      </DatePicker>
    </Flex>
  );
};

const InlineProgressEditor = ({
  keyResult,
}: {
  keyResult: KeyResultWithTeam;
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const [value, setValue] = useState(String(keyResult.currentValue));
  const { mutate: updateKeyResult, isPending } = useUpdateKeyResultMutation();
  const progress = getKeyResultProgress(keyResult);

  const handleOpenChange = (open: boolean) => {
    setIsOpen(open);
    if (open) setValue(String(keyResult.currentValue));
  };

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const currentValue = Number(value);

    if (!Number.isFinite(currentValue)) return;

    updateKeyResult(
      {
        keyResultId: keyResult.id,
        objectiveId: keyResult.objectiveId,
        data: { currentValue },
        silent: true,
      },
      {
        onSuccess: () => {
          setIsOpen(false);
        },
      },
    );
  };

  const updateBooleanProgress = (currentValue: number) => {
    updateKeyResult(
      {
        keyResultId: keyResult.id,
        objectiveId: keyResult.objectiveId,
        data: { currentValue },
        silent: true,
      },
      {
        onSuccess: () => {
          setIsOpen(false);
        },
      },
    );
  };

  return (
    <Popover onOpenChange={handleOpenChange} open={isOpen}>
      <Tooltip title="Update progress">
        <span>
          <Popover.Trigger asChild>
            <button
              className="hover:bg-state-hover flex size-7 items-center justify-center rounded transition-colors disabled:cursor-not-allowed disabled:opacity-50"
              disabled={isPending}
              type="button"
            >
              <CircleProgressBar
                progress={progress}
                size={22}
                strokeWidth={3}
              />
            </button>
          </Popover.Trigger>
        </span>
      </Tooltip>
      <Popover.Content align="end" className="w-72">
        <Box className="p-3">
          <Text className="mb-1" fontWeight="medium">
            Update progress
          </Text>
          <Text className="mb-3" color="muted" fontSize="sm">
            Current value:{" "}
            {formatKeyResultValue(
              keyResult.currentValue,
              keyResult.measurementType,
            )}{" "}
            of{" "}
            {formatKeyResultValue(
              keyResult.targetValue,
              keyResult.measurementType,
            )}
          </Text>
          {keyResult.measurementType === "boolean" ? (
            <Flex gap={2}>
              <Button
                className="flex-1"
                color="tertiary"
                disabled={isPending}
                onClick={() => {
                  updateBooleanProgress(0);
                }}
                size="sm"
                variant={keyResult.currentValue === 0 ? "solid" : "outline"}
              >
                Incomplete
              </Button>
              <Button
                className="flex-1"
                disabled={isPending}
                onClick={() => {
                  updateBooleanProgress(1);
                }}
                size="sm"
              >
                Complete
              </Button>
            </Flex>
          ) : (
            <form onSubmit={handleSubmit}>
              <Flex align="end" gap={2}>
                <Box className="flex-1">
                  <Text className="mb-1.5" color="muted" fontSize="sm">
                    Current value
                  </Text>
                  <Input
                    autoFocus
                    max={
                      keyResult.measurementType === "percentage"
                        ? 100
                        : undefined
                    }
                    min={
                      keyResult.measurementType === "percentage" ? 0 : undefined
                    }
                    onChange={(event) => {
                      setValue(event.target.value);
                    }}
                    step="any"
                    type="number"
                    value={value}
                  />
                </Box>
                <Button loading={isPending} size="sm" type="submit">
                  Save
                </Button>
              </Flex>
            </form>
          )}
        </Box>
      </Popover.Content>
    </Popover>
  );
};

const KeyResultLeadMenu = ({
  keyResult,
  lead,
}: {
  keyResult: KeyResultWithTeam;
  lead?: Member;
}) => {
  const { mutate: updateKeyResult } = useUpdateKeyResultMutation();

  return (
    <AssigneesMenu>
      <Tooltip title={lead ? `Lead: ${getDisplayName(lead)}` : "Assign lead"}>
        <span>
          <AssigneesMenu.Trigger>
            <button
              aria-label={
                lead
                  ? `Change lead from ${getDisplayName(lead)}`
                  : "Assign lead"
              }
              className="flex"
              type="button"
            >
              <Avatar
                name={lead ? getDisplayName(lead) : undefined}
                size="sm"
                src={lead?.avatarUrl}
              />
            </button>
          </AssigneesMenu.Trigger>
        </span>
      </Tooltip>
      <AssigneesMenu.Items
        assigneeId={keyResult.lead}
        onAssigneeSelected={(leadUser) => {
          updateKeyResult({
            keyResultId: keyResult.id,
            objectiveId: keyResult.objectiveId,
            data: { lead: leadUser },
            silent: true,
          });
        }}
        placeholder="Assign lead..."
      />
    </AssigneesMenu>
  );
};

const KeyResultRow = ({
  keyResult,
  memberById,
  selected,
  setSelected,
}: {
  keyResult: KeyResultWithTeam;
  memberById: ReadonlyMap<string, Member>;
  selected: boolean;
  setSelected: (selected: boolean) => void;
}) => {
  const { getTermDisplay } = useTerminology();
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  const [isUpdateOpen, setIsUpdateOpen] = useState(false);
  const [updateMode, setUpdateMode] = useState<"progress" | "other">("other");
  const { userRole } = useUserRole();
  const canEdit = userRole !== "guest";
  const { mutate: deleteKeyResult } = useDeleteKeyResultMutation();
  const lead = keyResult.lead ? memberById.get(keyResult.lead) : undefined;
  const reference = getKeyResultReference(
    keyResult.teamCode,
    keyResult.sequenceId,
  );
  const contributors = keyResult.contributors
    .map((id) => memberById.get(id))
    .filter((member): member is Member => Boolean(member));
  const openEditor = (mode: "progress" | "other") => {
    if (!canEdit) return;
    setUpdateMode(mode);
    setIsUpdateOpen(true);
  };

  return (
    <>
      <ContextMenu>
        <ContextMenu.Trigger>
          <Box>
            <RowWrapper
              className="@container cursor-pointer gap-4"
              onClick={(event) => {
                const target = event.target as HTMLElement;
                if (
                  target.closest(
                    "button, a, input, [role='menuitem'], [data-radix-popper-content-wrapper]",
                  )
                ) {
                  return;
                }
                openEditor("other");
              }}
              onKeyDown={(event) => {
                if (
                  event.key === "Enter" &&
                  event.target === event.currentTarget
                ) {
                  openEditor("other");
                }
              }}
            >
              <Flex
                align="center"
                className="relative min-w-0 flex-1 select-none"
                gap={2}
              >
                <Checkbox
                  checked={selected}
                  className="shrink-0 rounded md:absolute md:-left-[1.6rem]"
                  onCheckedChange={(checked) => {
                    setSelected(Boolean(checked));
                  }}
                  onClick={(event) => {
                    event.stopPropagation();
                  }}
                />
                {reference ? (
                  <Text className="text-text-muted min-w-[6ch] shrink-0 text-[0.95rem] whitespace-nowrap">
                    {reference}
                  </Text>
                ) : null}
                <Text
                  className="min-w-0 truncate whitespace-nowrap"
                  fontWeight="medium"
                >
                  {keyResult.name}
                </Text>
              </Flex>
              <Flex align="center" className="shrink-0" gap={3}>
                <Box className="hidden md:block">
                  <KeyResultDateRange canEdit={canEdit} keyResult={keyResult} />
                </Box>
                {canEdit ? (
                  <InlineProgressEditor keyResult={keyResult} />
                ) : (
                  <CircleProgressBar
                    progress={getKeyResultProgress(keyResult)}
                    size={22}
                    strokeWidth={3}
                  />
                )}
                {canEdit ? (
                  <KeyResultLeadMenu keyResult={keyResult} lead={lead} />
                ) : (
                  <Tooltip title={getDisplayName(lead)}>
                    <span>
                      <Avatar
                        name={lead ? getDisplayName(lead) : undefined}
                        size="sm"
                        src={lead?.avatarUrl}
                      />
                    </span>
                  </Tooltip>
                )}
                {contributors.length > 0 ? (
                  <Tooltip title={contributors.map(getDisplayName).join(", ")}>
                    <Flex className="-space-x-1.5">
                      {contributors.slice(0, 2).map((contributor) => (
                        <Avatar
                          className="ring-surface ring-1"
                          key={contributor.id}
                          name={getDisplayName(contributor)}
                          size="xs"
                          src={contributor.avatarUrl}
                        />
                      ))}
                      {contributors.length > 2 ? (
                        <span className="bg-surface-muted ring-surface flex size-5 items-center justify-center rounded-full text-xs ring-1">
                          +{contributors.length - 2}
                        </span>
                      ) : null}
                    </Flex>
                  </Tooltip>
                ) : null}
              </Flex>
            </RowWrapper>
          </Box>
        </ContextMenu.Trigger>
        <ContextMenu.Items className="w-56">
          <ContextMenu.Group>
            <ContextMenuItem
              disabled={!canEdit}
              icon={<EditIcon />}
              label="Edit..."
              onSelect={() => {
                openEditor("other");
              }}
            />
            <ContextMenuItem
              disabled={!canEdit}
              icon={<CheckIcon />}
              label="Update progress..."
              onSelect={() => {
                openEditor("progress");
              }}
            />
          </ContextMenu.Group>
          <ContextMenu.Separator />
          <ContextMenu.Group>
            <ContextMenuItem
              disabled={!canEdit}
              icon={<DeleteIcon className="text-danger dark:text-danger" />}
              label="Delete"
              onSelect={() => {
                setIsDeleteOpen(true);
              }}
            />
          </ContextMenu.Group>
        </ContextMenu.Items>
      </ContextMenu>
      <ConfirmDialog
        confirmText="Yes, Delete"
        description={`Are you sure you want to delete this ${getTermDisplay(
          "keyResultTerm",
        )}? This action cannot be undone.`}
        isOpen={isDeleteOpen}
        onClose={() => {
          setIsDeleteOpen(false);
        }}
        onConfirm={() => {
          deleteKeyResult({
            keyResultId: keyResult.id,
            objectiveId: keyResult.objectiveId,
          });
        }}
        title={`Delete ${getTermDisplay("keyResultTerm", {
          capitalize: true,
        })}`}
      />
      <UpdateKeyResultDialog
        isOpen={isUpdateOpen}
        keyResult={keyResult}
        onOpenChange={setIsUpdateOpen}
        updateMode={updateMode}
      />
    </>
  );
};

const ObjectiveGroup = ({
  group,
  memberById,
  selectedKeyResultIds,
  setSelectedKeyResultIds,
  teamColorById,
}: {
  group: ObjectiveKeyResultGroup;
  memberById: ReadonlyMap<string, Member>;
  selectedKeyResultIds: ReadonlySet<string>;
  setSelectedKeyResultIds: (ids: Set<string>) => void;
  teamColorById: ReadonlyMap<string, string>;
}) => {
  const pathname = usePathname();
  const { getTermDisplay } = useTerminology();
  const keyResultLabel = getTermDisplay("keyResultTerm");
  const [isCollapsed, setIsCollapsed] = useLocalStorage(
    `${pathname}:key-results:${group.objectiveId}`,
    false,
  );
  const groupKeyResultIds = group.keyResults.map((keyResult) => keyResult.id);
  const isGroupSelected =
    groupKeyResultIds.length > 0 &&
    groupKeyResultIds.every((id) => selectedKeyResultIds.has(id));

  return (
    <Box>
      <Container className="border-border bg-surface-muted/85 sticky top-0 z-1 border-b-[0.5px] py-[0.4rem] backdrop-blur select-none">
        <Flex align="center" justify="between">
          <Flex align="center" className="relative min-w-0 gap-1.5">
            <Checkbox
              checked={isGroupSelected}
              className="absolute -left-[1.6rem] hidden rounded md:inline"
              onCheckedChange={(checked) => {
                const nextSelected = new Set(selectedKeyResultIds);
                for (const id of groupKeyResultIds) {
                  if (checked) {
                    nextSelected.add(id);
                  } else {
                    nextSelected.delete(id);
                  }
                }
                setSelectedKeyResultIds(nextSelected);
              }}
            />
            <Button
              className="min-w-0"
              color="tertiary"
              leftIcon={<ObjectiveIcon className="h-[1.1rem] shrink-0" />}
              onClick={() => {
                setIsCollapsed(!isCollapsed);
              }}
              rightIcon={
                <ArrowDownIcon
                  className={cn(
                    "text-text-muted h-4 w-auto shrink-0 transition",
                    {
                      "-rotate-90": isCollapsed,
                    },
                  )}
                  strokeWidth={1}
                />
              }
              size="sm"
              variant="naked"
            >
              <Text className="truncate" fontWeight="medium">
                {group.objectiveName}
              </Text>
            </Button>
            <Button
              className="pointer-events-none gap-1 pr-2"
              color="tertiary"
              leftIcon={<TeamColor color={teamColorById.get(group.teamId)} />}
              rounded="md"
              size="xs"
              style={{
                backgroundColor: hexToRgba(
                  teamColorById.get(group.teamId),
                  0.1,
                ),
                borderColor: hexToRgba(teamColorById.get(group.teamId), 0.2),
              }}
              tabIndex={-1}
              variant="outline"
            >
              {group.teamName}
            </Button>
            <Text className="shrink-0 whitespace-nowrap" color="muted">
              {group.keyResults.length}{" "}
              {getTermDisplay("keyResultTerm", {
                variant: group.keyResults.length === 1 ? "singular" : "plural",
              })}
            </Text>
          </Flex>
          <Flex align="center" className="shrink-0" gap={2}>
            <Flex
              align="center"
              className="hidden w-24 shrink-0 md:flex"
              gap={2}
            >
              <ProgressBar className="w-12" progress={group.averageProgress} />
              <Text color="muted">{group.averageProgress}%</Text>
            </Flex>
            <NewKeyResultButton
              aria-label={`Add ${keyResultLabel} to ${group.objectiveName}`}
              asIcon
              iconOnly
              leftIcon={
                <PlusIcon className="text-foreground h-[1.1rem] w-auto" />
              }
              objectiveId={group.objectiveId}
              size="sm"
              variant="outline"
            />
          </Flex>
        </Flex>
      </Container>
      {!isCollapsed
        ? group.keyResults.map((keyResult) => (
            <KeyResultRow
              key={keyResult.id}
              keyResult={keyResult}
              memberById={memberById}
              selected={selectedKeyResultIds.has(keyResult.id)}
              setSelected={(selected) => {
                const nextSelected = new Set(selectedKeyResultIds);
                if (selected) {
                  nextSelected.add(keyResult.id);
                } else {
                  nextSelected.delete(keyResult.id);
                }
                setSelectedKeyResultIds(nextSelected);
              }}
            />
          ))
        : null}
      {!isCollapsed ? (
        <RowWrapper className="grid h-12 py-0 md:grid-cols-2">
          <Text className="min-w-0 truncate whitespace-nowrap" color="muted">
            Showing{" "}
            <span className="font-semibold">{group.keyResults.length}</span>{" "}
            {getTermDisplay("keyResultTerm", {
              variant: group.keyResults.length === 1 ? "singular" : "plural",
            })}{" "}
            for <span className="font-semibold">{group.objectiveName}</span>
          </Text>
        </RowWrapper>
      ) : null}
    </Box>
  );
};

export const KeyResultsList = ({
  groups,
  memberById,
  selectedKeyResultIds,
  setSelectedKeyResultIds,
  teamColorById,
}: {
  groups: ObjectiveKeyResultGroup[];
  memberById: ReadonlyMap<string, Member>;
  selectedKeyResultIds: ReadonlySet<string>;
  setSelectedKeyResultIds: (ids: Set<string>) => void;
  teamColorById: ReadonlyMap<string, string>;
}) => (
  <Box>
    {groups.map((group) => (
      <ObjectiveGroup
        group={group}
        key={group.objectiveId}
        memberById={memberById}
        selectedKeyResultIds={selectedKeyResultIds}
        setSelectedKeyResultIds={setSelectedKeyResultIds}
        teamColorById={teamColorById}
      />
    ))}
  </Box>
);
