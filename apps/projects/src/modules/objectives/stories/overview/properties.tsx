import { CalendarIcon } from "icons";
import { Text, Flex, Button, Avatar, DatePicker } from "ui";
import { cn } from "lib";
import { addDays, format } from "date-fns";
import {
  AssigneesMenu,
  PrioritiesMenu,
  PriorityIcon,
  StatusesMenu,
  StoryStatusIcon,
} from "@/components/ui";

export const Properties = () => {
  const endDate = new Date().toISOString();
  return (
    <Flex align="center" className="mt-6" gap={2} wrap>
      <Text color="muted" fontWeight="semibold">
        Properties:
      </Text>
      <StatusesMenu>
        <StatusesMenu.Trigger>
          <Button
            color="tertiary"
            leftIcon={<StoryStatusIcon />}
            size="sm"
            type="button"
            variant="naked"
          >
            Backlog
          </Button>
        </StatusesMenu.Trigger>
        <StatusesMenu.Items
          setStatusId={(_) => {}}
          // statusId={statusId}
        />
      </StatusesMenu>
      <PrioritiesMenu>
        <PrioritiesMenu.Trigger>
          <Button
            color="tertiary"
            leftIcon={<PriorityIcon />}
            size="sm"
            type="button"
            variant="naked"
          >
            No Priority
          </Button>
        </PrioritiesMenu.Trigger>
        <PrioritiesMenu.Items setPriority={(_) => {}} />
      </PrioritiesMenu>
      <AssigneesMenu>
        <AssigneesMenu.Trigger>
          <Button
            className={cn("font-medium", {
              "text-gray-200 dark:text-gray-300": false,
            })}
            color="tertiary"
            leftIcon={
              <Avatar
                className={cn({
                  "text-dark/80 dark:text-gray-200": false,
                })}
                size="xs"
              />
            }
            size="sm"
            type="button"
            variant="naked"
          >
            <Text as="span" color="muted">
              Assign lead
            </Text>
          </Button>
        </AssigneesMenu.Trigger>
        <AssigneesMenu.Items
          // assigneeId={assigneeId}
          onAssigneeSelected={(_) => {}}
        />
      </AssigneesMenu>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            color="tertiary"
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem] w-auto", {
                  "text-gray/80 dark:text-gray-300/80": false,
                })}
              />
            }
            size="sm"
            variant="naked"
          >
            {/* {startDate ? (
        format(new Date(), "MMM d, yyyy")
      ) : (
        <Text as="span" color="muted">Add start date</Text>
      )} */}
            <Text as="span" color="muted">
              Start date
            </Text>
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(_) => {
            // handleUpdate({
            //   startDate: formatISO(day),
            // });
          }}
          // selected={startDate ? new Date(startDate) : undefined}
        />
      </DatePicker>
      <DatePicker>
        <DatePicker.Trigger>
          <Button
            className={cn({
              "text-primary dark:text-primary": new Date(endDate) < new Date(),
              "text-warning dark:text-warning":
                new Date(endDate) <= addDays(new Date(), 7) &&
                new Date(endDate) >= new Date(),
              "text-gray/80 dark:text-gray-300/80": !endDate,
            })}
            color="tertiary"
            leftIcon={
              <CalendarIcon
                className={cn("h-[1.15rem]", {
                  "text-primary dark:text-primary":
                    new Date(endDate) < new Date(),
                  "text-warning dark:text-warning":
                    new Date(endDate) <= addDays(new Date(), 7) &&
                    new Date(endDate) >= new Date(),
                  "text-gray/80 dark:text-gray-300/80": !endDate,
                })}
              />
            }
            size="sm"
            variant="naked"
          >
            {endDate ? (
              format(new Date(endDate), "MMM d, yy")
            ) : (
              <Text color="muted">Target date</Text>
            )}
          </Button>
        </DatePicker.Trigger>
        <DatePicker.Calendar
          onDayClick={(_) => {
            // handleUpdate({
            //   endDate: formatISO(day),
            // });
          }}
          selected={endDate ? new Date(endDate) : undefined}
        />
      </DatePicker>
    </Flex>
  );
};
