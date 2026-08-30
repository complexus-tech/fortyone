import type { ComponentProps } from "react";
import { format, formatISO } from "date-fns";
import { CalendarIcon, CloseIcon } from "icons";
import { Avatar, Button, ColorPicker, DatePicker, Flex, Tooltip } from "ui";
import { ObjectiveStatusesMenu } from "../objective-statuses-menu";
import { PriorityIcon } from "../priority-icon";
import { StoryStatusIcon } from "../story-status-icon";
import { AssigneesMenu } from "../story/assignees-menu";
import { PrioritiesMenu } from "../story/priorities-menu";

type ObjectivePriority = NonNullable<
  ComponentProps<typeof PriorityIcon>["priority"]
>;

type ObjectiveLead = {
  avatarUrl: string | null;
  fullName: string;
  username: string;
};

type ObjectiveStatusOption = {
  id: string;
  name: string;
};

export const NewObjectiveDialogControls = ({
  color,
  endDate,
  lead,
  leadUserId,
  onColorChange,
  onEndDateChange,
  onLeadChange,
  onPriorityChange,
  onStartDateChange,
  onStatusChange,
  priority,
  startDate,
  statusId,
  statuses,
}: {
  color?: string;
  endDate?: string | null;
  lead?: ObjectiveLead;
  leadUserId?: string | null;
  onColorChange: (color: string) => void;
  onEndDateChange: (endDate: string | null) => void;
  onLeadChange: (leadUserId: string | null) => void;
  onPriorityChange: (priority: ObjectivePriority) => void;
  onStartDateChange: (startDate: string | null) => void;
  onStatusChange: (statusId: string) => void;
  priority?: ObjectivePriority;
  startDate?: string | null;
  statusId?: string;
  statuses: ObjectiveStatusOption[];
}) => (
  <Flex align="center" className="mt-4 gap-1.5" wrap>
    <Tooltip title="Objective color">
      <span>
        <ColorPicker onChange={onColorChange} value={color} />
      </span>
    </Tooltip>
    <ObjectiveStatusesMenu>
      <ObjectiveStatusesMenu.Trigger>
        <Button
          color="tertiary"
          leftIcon={<StoryStatusIcon statusId={statusId} />}
          size="sm"
          type="button"
          variant="outline"
        >
          {statuses.find((status) => status.id === statusId)?.name}
        </Button>
      </ObjectiveStatusesMenu.Trigger>
      <ObjectiveStatusesMenu.Items
        setStatusId={onStatusChange}
        statusId={statusId}
      />
    </ObjectiveStatusesMenu>
    <PrioritiesMenu>
      <PrioritiesMenu.Trigger>
        <Button
          color="tertiary"
          leftIcon={<PriorityIcon priority={priority} />}
          size="sm"
          type="button"
          variant="outline"
        >
          {priority ?? "No Priority"}
        </Button>
      </PrioritiesMenu.Trigger>
      <PrioritiesMenu.Items
        priority={priority}
        setPriority={onPriorityChange}
      />
    </PrioritiesMenu>
    <DatePicker>
      <DatePicker.Trigger>
        <Button
          className="px-2"
          color="tertiary"
          leftIcon={<CalendarIcon className="h-4 w-auto" />}
          rightIcon={
            startDate ? (
              <CloseIcon
                aria-label="Remove date"
                className="h-4 w-auto"
                onClick={(event) => {
                  event.stopPropagation();
                  onStartDateChange(null);
                }}
                role="button"
              />
            ) : null
          }
          size="sm"
          variant="outline"
        >
          {startDate
            ? format(new Date(startDate), "MMM d, yyyy")
            : "Start date"}
        </Button>
      </DatePicker.Trigger>
      <DatePicker.Calendar
        onDayClick={(date) => {
          onStartDateChange(formatISO(date, { representation: "date" }));
        }}
      />
    </DatePicker>
    <DatePicker>
      <DatePicker.Trigger>
        <Button
          className="px-2"
          color="tertiary"
          leftIcon={<CalendarIcon className="h-4 w-auto" />}
          rightIcon={
            endDate ? (
              <CloseIcon
                aria-label="Remove date"
                className="h-4 w-auto"
                onClick={(event) => {
                  event.stopPropagation();
                  onEndDateChange(null);
                }}
                role="button"
              />
            ) : null
          }
          size="sm"
          variant="outline"
        >
          {endDate ? format(new Date(endDate), "MMM d, yyyy") : "Deadline"}
        </Button>
      </DatePicker.Trigger>
      <DatePicker.Calendar
        fromDate={startDate ? new Date(startDate) : undefined}
        onDayClick={(date) => {
          onEndDateChange(formatISO(date, { representation: "date" }));
        }}
      />
    </DatePicker>
    <AssigneesMenu>
      <AssigneesMenu.Trigger>
        <Button
          className="gap-1.5 px-2"
          color="tertiary"
          leftIcon={
            <Avatar
              color="tertiary"
              name={lead?.fullName}
              size="xs"
              src={lead?.avatarUrl}
            />
          }
          size="sm"
          variant="outline"
        >
          <span className="relative -top-px inline-block max-w-[12ch] truncate">
            {lead?.username || "Lead"}
          </span>
        </Button>
      </AssigneesMenu.Trigger>
      <AssigneesMenu.Items
        assigneeId={leadUserId || null}
        onAssigneeSelected={onLeadChange}
      />
    </AssigneesMenu>
  </Flex>
);
