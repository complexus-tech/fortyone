import { useRef, type Dispatch, type ReactNode } from "react";
import { addDays, format, formatISO } from "date-fns";
import {
  CalendarIcon,
  CloseIcon,
  EstimateIcon,
  ObjectiveIcon,
  PlusIcon,
  SprintsIcon,
  TagsIcon,
  Time02Icon,
  TimeScheduleIcon,
} from "icons";
import { cn } from "lib";
import { Avatar, Box, Button, DatePicker, Flex } from "ui";
import { getNewStoryAutoSchedulingEnabled } from "@/lib/auto-scheduling";
import { formatEstimate, type EstimateScheme } from "@/lib/estimate";
import { formatTimeNeeded } from "@/lib/time-needed";
import { AssigneesMenu } from "./story/assignees-menu";
import { AutoSchedulingMenu } from "./story/auto-scheduling-menu";
import { EstimateMenu } from "./story/estimate-menu";
import { LabelsMenu } from "./story/labels-menu";
import { ObjectiveKeyResultMenu } from "./story/objective-key-result-menu";
import { PrioritiesMenu } from "./story/priorities-menu";
import { SprintsMenu } from "./story/sprints-menu";
import { StatusesMenu } from "./story/statuses-menu";
import { TimeNeededMenu } from "./story/time-needed-menu";
import { PriorityIcon } from "./priority-icon";
import {
  getDeadlineForSprintSelection,
  toDateOnly,
  type DeadlineSource,
  type NewStoryDialogForm,
  type StoryFormAction,
} from "./new-story-dialog-form";
import { StoryStatusIcon } from "./story-status-icon";

type SelectedAssignee = {
  avatarUrl?: string | null;
  fullName?: string;
  username?: string;
};

type SelectedLabel = {
  color: string;
  id: string;
  name: string;
};

export const NewStoryDialogFields = ({
  canUseBackgroundMaya,
  children,
  currentTeamId,
  deadlineSourceRef,
  dispatch,
  estimateScheme,
  isMayaAssigned,
  mayaAssigneeId,
  member,
  objectiveTerm,
  selectedLabels,
  showObjectives,
  showSprints,
  sprintName,
  sprintTerm,
  storyForm,
  strategyLinkLabel,
  teamStatuses,
}: {
  canUseBackgroundMaya: boolean;
  children?: ReactNode;
  currentTeamId?: string;
  deadlineSourceRef: { current: DeadlineSource };
  dispatch: Dispatch<StoryFormAction>;
  estimateScheme: EstimateScheme;
  isMayaAssigned: boolean;
  mayaAssigneeId?: string | null;
  member?: SelectedAssignee | null;
  objectiveTerm: string;
  selectedLabels: SelectedLabel[];
  showObjectives: boolean;
  showSprints: boolean;
  sprintName?: string;
  sprintTerm: string;
  storyForm: NewStoryDialogForm;
  strategyLinkLabel?: string;
  teamStatuses: { id: string; name: string }[];
}) => {
  const assigneeButtonRef = useRef<HTMLButtonElement>(null);
  const timeNeededButtonRef = useRef<HTMLButtonElement>(null);
  const selectedLabelIds = storyForm.labelIds ?? [];

  return (
    <Flex align="center" className="mt-4 gap-1.5" wrap>
      <StatusesMenu>
        <StatusesMenu.Trigger>
          <Button
            className="dark:bg-surface-elevated/90"
            color="tertiary"
            leftIcon={
              <StoryStatusIcon
                className="size-4 shrink-0"
                statusId={storyForm.statusId}
              />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {
              teamStatuses.find((status) => status.id === storyForm.statusId)
                ?.name
            }
          </Button>
        </StatusesMenu.Trigger>
        <StatusesMenu.Items
          setStatusId={(statusId) => {
            dispatch({
              type: "SET_FIELD",
              field: "statusId",
              value: statusId,
            });
          }}
          statusId={storyForm.statusId}
          teamId={currentTeamId ?? ""}
        />
      </StatusesMenu>
      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            className="dark:bg-surface-elevated/90"
            color="tertiary"
            leftIcon={
              <PriorityIcon className="h-4" priority={storyForm.priority} />
            }
            size="sm"
            type="button"
            variant="outline"
          >
            {storyForm.priority}
          </Button>
        </PrioritiesMenu.Trigger>
        <PrioritiesMenu.Items
          priority={storyForm.priority}
          setPriority={(priority) => {
            dispatch({
              type: "SET_FIELD",
              field: "priority",
              value: priority,
            });
          }}
        />
      </PrioritiesMenu>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            className="dark:bg-surface-elevated/90 px-2"
            color="tertiary"
            leftIcon={<CalendarIcon className="h-4.5 w-auto" />}
            rightIcon={
              storyForm.startDate ? (
                <CloseIcon
                  aria-label="Remove date"
                  className="h-4 w-auto"
                  onClick={() => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "startDate",
                      value: null,
                    });
                  }}
                  role="button"
                />
              ) : null
            }
            size="sm"
            variant="outline"
          >
            {storyForm.startDate
              ? format(new Date(storyForm.startDate), "MMM d, yyyy")
              : "Start date"}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(date) => {
            dispatch({
              type: "SET_FIELD",
              field: "startDate",
              value: formatISO(date, { representation: "date" }),
            });
          }}
        />
      </DatePicker>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            className={cn("dark:bg-surface-elevated/90 px-2", {
              "text-primary dark:text-primary": storyForm.endDate
                ? new Date(storyForm.endDate) < new Date()
                : false,
              "text-warning dark:text-warning": storyForm.endDate
                ? new Date(storyForm.endDate) <= addDays(new Date(), 7) &&
                  new Date(storyForm.endDate) >= new Date()
                : false,
            })}
            color="tertiary"
            leftIcon={<CalendarIcon className="h-4.5 w-auto" />}
            rightIcon={
              storyForm.endDate ? (
                <CloseIcon
                  aria-label="Remove date"
                  className="h-4 w-auto"
                  onClick={() => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "endDate",
                      value: null,
                    });
                    deadlineSourceRef.current = "cleared";
                  }}
                  role="button"
                />
              ) : null
            }
            size="sm"
            variant="outline"
          >
            {storyForm.endDate
              ? format(new Date(storyForm.endDate), "MMM d, yyyy")
              : "Deadline"}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          fromDate={
            storyForm.startDate ? new Date(storyForm.startDate) : undefined
          }
          onDayClick={(date) => {
            dispatch({
              type: "SET_FIELD",
              field: "endDate",
              value: formatISO(date, { representation: "date" }),
            });
            deadlineSourceRef.current = "manual";
          }}
        />
      </DatePicker>
      <Box className="order-8">
        <AssigneesMenu>
          <AssigneesMenu.Trigger>
            <Button
              className="dark:bg-surface-elevated/90 gap-1.5 px-2"
              color="tertiary"
              leftIcon={
                <Avatar
                  name={member?.fullName}
                  size="xs"
                  src={member?.avatarUrl}
                />
              }
              ref={assigneeButtonRef}
              size="sm"
              variant="outline"
            >
              <span className="relative -top-px inline-block max-w-[12ch] truncate">
                {member?.username || "Assignee"}
              </span>
            </Button>
          </AssigneesMenu.Trigger>
          <AssigneesMenu.Items
            assigneeId={storyForm.assigneeId}
            onAssigneeSelected={(assigneeId) => {
              dispatch({
                type: "PATCH_FORM",
                payload: {
                  assigneeId,
                  autoSchedulingEnabled: getNewStoryAutoSchedulingEnabled({
                    currentEnabled: Boolean(storyForm.autoSchedulingEnabled),
                    mayaAssigneeId,
                    selectedAssigneeId: assigneeId,
                  }),
                },
              });
            }}
            teamId={currentTeamId}
          />
        </AssigneesMenu>
      </Box>
      <Box className="order-9">
        <EstimateMenu>
          <EstimateMenu.Trigger>
            <Button
              className={cn("dark:bg-surface-elevated/90 gap-1.5 px-2", {
                "text-text-muted": !storyForm.estimateValue,
              })}
              color="tertiary"
              leftIcon={
                <EstimateIcon
                  className={cn("h-4.5 w-auto", {
                    "text-text-muted": !storyForm.estimateValue,
                  })}
                />
              }
              size="sm"
              type="button"
              variant="outline"
            >
              {storyForm.estimateValue
                ? formatEstimate(
                    estimateScheme,
                    storyForm.estimateValue,
                    "full",
                  )
                : "Complexity"}
            </Button>
          </EstimateMenu.Trigger>
          <EstimateMenu.Items
            estimateScheme={estimateScheme}
            estimateValue={storyForm.estimateValue}
            setEstimateValue={(estimateValue) => {
              dispatch({
                type: "SET_FIELD",
                field: "estimateValue",
                value: estimateValue,
              });
            }}
          />
        </EstimateMenu>
      </Box>
      <Box className="order-10">
        <TimeNeededMenu>
          <TimeNeededMenu.Trigger>
            <Button
              className={cn("dark:bg-surface-elevated/90 gap-1.5 px-2", {
                "text-text-muted": !storyForm.estimatedDurationMinutes,
              })}
              color="tertiary"
              leftIcon={
                <Time02Icon
                  className={cn("h-4.5 w-auto", {
                    "text-text-muted": !storyForm.estimatedDurationMinutes,
                  })}
                />
              }
              ref={timeNeededButtonRef}
              size="sm"
              type="button"
              variant="outline"
            >
              {storyForm.estimatedDurationMinutes
                ? formatTimeNeeded(storyForm.estimatedDurationMinutes, "full")
                : "Time needed"}
            </Button>
          </TimeNeededMenu.Trigger>
          <TimeNeededMenu.Items
            estimatedDurationMinutes={storyForm.estimatedDurationMinutes}
            minimumFocusBlockMinutes={storyForm.minimumFocusBlockMinutes}
            setTimeNeeded={({
              estimatedDurationMinutes,
              minimumFocusBlockMinutes,
            }) => {
              dispatch({
                type: "SET_FIELD",
                field: "estimatedDurationMinutes",
                value: estimatedDurationMinutes,
              });
              dispatch({
                type: "SET_FIELD",
                field: "minimumFocusBlockMinutes",
                value: minimumFocusBlockMinutes,
              });
            }}
          />
        </TimeNeededMenu>
      </Box>
      <Box className="order-11">
        <AutoSchedulingMenu>
          <AutoSchedulingMenu.Trigger>
            <Button
              className="dark:bg-surface-elevated/90 gap-1.5 px-2"
              color="tertiary"
              disabled={!canUseBackgroundMaya || isMayaAssigned}
              leftIcon={<TimeScheduleIcon className="h-4.5 w-auto" />}
              size="sm"
              type="button"
              variant="outline"
            >
              Auto-scheduling: {storyForm.autoSchedulingEnabled ? "On" : "Off"}
            </Button>
          </AutoSchedulingMenu.Trigger>
          <AutoSchedulingMenu.Items
            autoSchedulingEnabled={Boolean(storyForm.autoSchedulingEnabled)}
            autoSchedulingLocked={Boolean(storyForm.autoSchedulingLocked)}
            setAutoSchedulingEnabled={(enabled) => {
              dispatch({
                type: "PATCH_FORM",
                payload: { autoSchedulingEnabled: enabled },
              });
            }}
          />
        </AutoSchedulingMenu>
      </Box>
      <Box className="order-5">
        {selectedLabels.length > 0 ? (
          <Flex align="center" className="gap-1" wrap>
            {selectedLabels.map((label) => (
              <LabelsMenu key={label.id}>
                <LabelsMenu.Trigger>
                  <Button
                    className="dark:bg-surface-elevated/90 gap-1.5 px-2.5"
                    color="tertiary"
                    leftIcon={
                      <TagsIcon
                        className="h-4.5 w-auto"
                        style={{ color: label.color }}
                      />
                    }
                    size="sm"
                    type="button"
                    variant="outline"
                  >
                    <span className="inline-block max-w-[12ch] truncate">
                      {label.name}
                    </span>
                  </Button>
                </LabelsMenu.Trigger>
                <LabelsMenu.Items
                  labelIds={selectedLabelIds}
                  setLabelIds={(labelIds) => {
                    dispatch({
                      type: "SET_FIELD",
                      field: "labelIds",
                      value: labelIds,
                    });
                  }}
                  teamId={currentTeamId ?? ""}
                />
              </LabelsMenu>
            ))}
            <LabelsMenu>
              <LabelsMenu.Trigger>
                <Button
                  asIcon
                  className="dark:bg-surface-elevated/90"
                  color="tertiary"
                  leftIcon={<PlusIcon />}
                  rounded="full"
                  size="sm"
                  title="Add labels"
                  type="button"
                  variant="outline"
                >
                  <span className="sr-only">Add labels</span>
                </Button>
              </LabelsMenu.Trigger>
              <LabelsMenu.Items
                labelIds={selectedLabelIds}
                setLabelIds={(labelIds) => {
                  dispatch({
                    type: "SET_FIELD",
                    field: "labelIds",
                    value: labelIds,
                  });
                }}
                teamId={currentTeamId ?? ""}
              />
            </LabelsMenu>
          </Flex>
        ) : (
          <LabelsMenu>
            <LabelsMenu.Trigger>
              <Button
                className="dark:bg-surface-elevated/90 gap-1.5 px-2"
                color="tertiary"
                leftIcon={<TagsIcon className="h-4.5 w-auto" />}
                size="sm"
                type="button"
                variant="outline"
              >
                Labels
              </Button>
            </LabelsMenu.Trigger>
            <LabelsMenu.Items
              labelIds={selectedLabelIds}
              setLabelIds={(labelIds) => {
                dispatch({
                  type: "SET_FIELD",
                  field: "labelIds",
                  value: labelIds,
                });
              }}
              teamId={currentTeamId ?? ""}
            />
          </LabelsMenu>
        )}
      </Box>
      <Box className="order-6">
        {showObjectives ? (
          <ObjectiveKeyResultMenu
            keyResultId={storyForm.keyResultId ?? null}
            objectiveId={storyForm.objectiveId ?? null}
            onChange={(selection) => {
              dispatch({
                type: "SET_FIELD",
                field: "objectiveId",
                value: selection.objectiveId,
              });
              dispatch({
                type: "SET_FIELD",
                field: "keyResultId",
                value: selection.keyResultId,
              });
            }}
            teamId={currentTeamId ?? ""}
          >
            <Button
              className="dark:bg-surface-elevated/90 gap-1 px-2"
              color="tertiary"
              leftIcon={<ObjectiveIcon />}
              size="sm"
              variant="outline"
            >
              <span className="inline-block max-w-[18ch] truncate">
                {strategyLinkLabel || objectiveTerm}
              </span>
            </Button>
          </ObjectiveKeyResultMenu>
        ) : null}
      </Box>
      <Box className="order-7">
        {showSprints ? (
          <SprintsMenu>
            <SprintsMenu.Trigger>
              <Button
                className="dark:bg-surface-elevated/90 gap-1 px-2"
                color="tertiary"
                leftIcon={<SprintsIcon />}
                size="sm"
                variant="outline"
              >
                <span className="inline-block max-w-[12ch] truncate">
                  {sprintName || sprintTerm}
                </span>
              </Button>
            </SprintsMenu.Trigger>
            <SprintsMenu.Items
              setSprintId={(sprintId, sprintEndDate) => {
                const nextDeadline = getDeadlineForSprintSelection({
                  currentEndDate: storyForm.endDate,
                  currentSource: deadlineSourceRef.current,
                  sprintEndDate: toDateOnly(sprintEndDate),
                });
                dispatch({
                  type: "SET_FIELD",
                  field: "sprintId",
                  value: sprintId,
                });
                dispatch({
                  type: "SET_FIELD",
                  field: "endDate",
                  value: nextDeadline.endDate,
                });
                deadlineSourceRef.current = nextDeadline.source;
              }}
              sprintId={storyForm.sprintId ?? undefined}
              teamId={currentTeamId}
            />
          </SprintsMenu>
        ) : null}
      </Box>
      {children}
    </Flex>
  );
};
