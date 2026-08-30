"use client";

import { addDays, format, formatISO } from "date-fns";
import { Box, Button, DatePicker, Flex, Text, Tooltip } from "ui";
import { CalendarIcon } from "icons";
import { cn } from "lib";
import { PropertyOption as Option } from "@/components/ui/property-option";
import { getDueDateMessage } from "@/components/ui/story/due-date-tooltip";
import type { DetailedStory } from "../../types";
import { useOptionHotkey } from "./use-option-hotkey";

type DateOptionsProps = {
  disabled: boolean;
  endDate: DetailedStory["endDate"];
  isCompact: boolean;
  isNotifications: boolean;
  onUpdate: (story: Partial<DetailedStory>) => void;
  startDate: DetailedStory["startDate"];
  storyTerm: string;
};

export const DateOptions = ({
  disabled,
  endDate,
  isCompact,
  isNotifications,
  onUpdate,
  startDate,
  storyTerm,
}: DateOptionsProps) => {
  const startDateButtonRef = useOptionHotkey("b", !disabled);
  const dueDateButtonRef = useOptionHotkey("d", !disabled);

  return (
    <>
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Start date"
        value={
          <DatePicker>
            <DatePicker.Trigger>
              <Button
                color="tertiary"
                disabled={disabled}
                leftIcon={
                  <CalendarIcon
                    className={cn("h-[1.15rem] w-auto", {
                      "text-text-muted": !startDate,
                    })}
                  />
                }
                ref={startDateButtonRef}
                size="sm"
                variant={isCompact ? "solid" : "naked"}
              >
                {startDate ? (
                  format(new Date(startDate), "MMM d, yyyy")
                ) : (
                  <Text color="muted">Add start date</Text>
                )}
              </Button>
            </DatePicker.Trigger>
            <DatePicker.Calendar
              onDayClick={(day) => {
                onUpdate({
                  startDate: formatISO(day, { representation: "date" }),
                });
              }}
              selected={startDate ? new Date(startDate) : undefined}
            />
          </DatePicker>
        }
      />
      <Option
        isCompact={isCompact}
        isNotifications={isNotifications}
        label="Deadline"
        value={
          <DatePicker>
            <Tooltip
              className="py-3"
              hidden={!endDate}
              title={
                <Flex align="start" gap={2}>
                  <CalendarIcon
                    className={cn("relative top-[2.5px] h-5 w-auto", {
                      "text-primary dark:text-primary":
                        new Date(endDate!) < new Date(),
                      "text-warning dark:text-warning":
                        new Date(endDate!) <= addDays(new Date(), 7) &&
                        new Date(endDate!) >= new Date(),
                    })}
                  />
                  <Box>{getDueDateMessage(new Date(endDate!), storyTerm)}</Box>
                </Flex>
              }
            >
              <span>
                <DatePicker.Trigger>
                  <Button
                    className={cn({
                      "text-primary dark:text-primary":
                        endDate && new Date(endDate) < new Date(),
                      "text-warning dark:text-warning":
                        endDate &&
                        new Date(endDate) <= addDays(new Date(), 7) &&
                        new Date(endDate) >= new Date(),
                      "text-text-muted": !endDate,
                    })}
                    color="tertiary"
                    disabled={disabled}
                    leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                    ref={dueDateButtonRef}
                    size="sm"
                    variant={isCompact ? "solid" : "naked"}
                  >
                    {endDate ? (
                      format(new Date(endDate), "MMM d, yyyy")
                    ) : (
                      <Text color="muted">Add deadline</Text>
                    )}
                  </Button>
                </DatePicker.Trigger>
              </span>
            </Tooltip>
            <DatePicker.Calendar
              onDayClick={(day) => {
                onUpdate({
                  endDate: formatISO(day, { representation: "date" }),
                });
              }}
              selected={endDate ? new Date(endDate) : undefined}
            />
          </DatePicker>
        }
      />
    </>
  );
};
