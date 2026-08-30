"use client";

import type { FormEvent } from "react";
import { useState } from "react";
import { addDays, format, formatISO, startOfDay } from "date-fns";
import { cn } from "lib";
import { CalendarIcon, CheckIcon, DeleteIcon, EditIcon } from "icons";
import {
  Avatar,
  Box,
  Button,
  Checkbox,
  CircleProgressBar,
  ContextMenu,
  DatePicker,
  Flex,
  Input,
  Popover,
  Text,
  Tooltip,
} from "ui";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { ContextMenuItem } from "@/components/ui/story/context-menu-item";
import { AssigneesMenu } from "@/components/ui/story/assignees-menu";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import { RowWrapper } from "@/components/ui/row-wrapper";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useUserRole } from "@/hooks/role";
import {
  UpdateKeyResultDialog,
  useDeleteKeyResultMutation,
  useUpdateKeyResultMutation,
} from "@/modules/objectives/public/key-results";
import type { KeyResultWithTeam } from "../types";
import {
  formatKeyResultValue,
  getKeyResultProgress,
  getKeyResultReference,
} from "../utils";
import type { KeyResultsMember } from "./key-results-member";

const getDisplayName = (member?: KeyResultsMember) =>
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
  lead?: KeyResultsMember;
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

export const KeyResultRow = ({
  keyResult,
  memberById,
  selected,
  setSelected,
}: {
  keyResult: KeyResultWithTeam;
  memberById: ReadonlyMap<string, KeyResultsMember>;
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
    .filter((member): member is KeyResultsMember => Boolean(member));
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
