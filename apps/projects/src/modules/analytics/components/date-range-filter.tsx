"use client";
import { useState } from "react";
import { Box, Button, Divider, Flex, Popover, Text, DatePicker } from "ui";
import { CalendarIcon, CloseIcon } from "icons";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
import { cn } from "lib";

type DatePreset = {
  label: string;
  value: string;
  getDates: () => { startDate: Date; endDate: Date };
};

const datePresets: DatePreset[] = [
  {
    label: "Last 7 days",
    value: "7d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 7)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "Last 30 days",
    value: "30d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 30)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "Last 90 days",
    value: "90d",
    getDates: () => ({
      startDate: startOfDay(subDays(new Date(), 90)),
      endDate: endOfDay(new Date()),
    }),
  },
  {
    label: "This month",
    value: "month",
    getDates: () => {
      const now = new Date();
      const startOfMonth = new Date(now.getFullYear(), now.getMonth(), 1);
      return {
        startDate: startOfDay(startOfMonth),
        endDate: endOfDay(now),
      };
    },
  },
];

type DateRangeFilterProps = {
  startDate?: string;
  endDate?: string;
  onDateChange: (startDate?: string, endDate?: string) => void;
};

const DateRangeSelector = ({
  startDate,
  endDate,
  onDateChange,
}: DateRangeFilterProps) => {
  const [customStartDate, setCustomStartDate] = useState<Date | undefined>(
    startDate ? new Date(startDate) : undefined,
  );
  const [customEndDate, setCustomEndDate] = useState<Date | undefined>(
    endDate ? new Date(endDate) : undefined,
  );

  const handlePresetSelect = (preset: DatePreset) => {
    const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
    onDateChange(presetStart.toISOString(), presetEnd.toISOString());
  };

  const handleCustomDateChange = (start?: Date, end?: Date) => {
    setCustomStartDate(start);
    setCustomEndDate(end);
    onDateChange(start?.toISOString(), end?.toISOString());
  };

  const clearDates = () => {
    setCustomStartDate(undefined);
    setCustomEndDate(undefined);
    onDateChange(undefined, undefined);
  };

  const getCurrentPreset = () => {
    if (!startDate || !endDate) return null;

    const start = new Date(startDate);
    const end = new Date(endDate);

    return datePresets.find((preset) => {
      const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
      return (
        Math.abs(start.getTime() - presetStart.getTime()) <
          24 * 60 * 60 * 1000 &&
        Math.abs(end.getTime() - presetEnd.getTime()) < 24 * 60 * 60 * 1000
      );
    });
  };

  const currentPreset = getCurrentPreset();

  return (
    <Box>
      <Text className="mb-3" fontWeight="medium">
        Quick Presets
      </Text>
      <Flex className="mb-4" gap={2} wrap>
        {datePresets.map((preset) => (
          <Button
            className={cn({
              "ring-2 dark:ring-white/30":
                currentPreset?.value === preset.value,
            })}
            color="tertiary"
            key={preset.value}
            onClick={() => {
              handlePresetSelect(preset);
            }}
            size="sm"
            variant={
              currentPreset?.value === preset.value ? "solid" : "outline"
            }
          >
            {preset.label}
          </Button>
        ))}
      </Flex>

      <Divider className="my-4" />

      <Text className="mb-2" fontWeight="medium">
        Custom Range
      </Text>
      <Flex gap={2} wrap>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="gap-2 px-2"
              color="tertiary"
              leftIcon={<CalendarIcon className="h-4 w-auto" />}
              rightIcon={
                customStartDate ? (
                  <CloseIcon
                    aria-label="Clear start date"
                    className="h-4 w-auto"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCustomDateChange(undefined, customEndDate);
                    }}
                    role="button"
                  />
                ) : null
              }
              size="sm"
              variant="outline"
            >
              {customStartDate
                ? format(customStartDate, "MMM d, yyyy")
                : "Start date"}
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            onDayClick={(date: Date) => {
              handleCustomDateChange(date, customEndDate);
            }}
            selected={customStartDate}
          />
        </DatePicker>

        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="gap-2 px-2"
              color="tertiary"
              leftIcon={<CalendarIcon className="h-4 w-auto" />}
              rightIcon={
                customEndDate ? (
                  <CloseIcon
                    aria-label="Clear end date"
                    className="h-4 w-auto"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleCustomDateChange(customStartDate, undefined);
                    }}
                    role="button"
                  />
                ) : null
              }
              size="sm"
              variant="outline"
            >
              {customEndDate
                ? format(customEndDate, "MMM d, yyyy")
                : "End date"}
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            fromDate={customStartDate}
            onDayClick={(date: Date) => {
              handleCustomDateChange(customStartDate, date);
            }}
            selected={customEndDate}
          />
        </DatePicker>
        {customStartDate || customEndDate ? (
          <Button
            color="tertiary"
            onClick={clearDates}
            size="sm"
            variant="outline"
          >
            Clear
          </Button>
        ) : null}
      </Flex>
    </Box>
  );
};

export const DateRangeFilter = ({
  startDate,
  endDate,
  onDateChange,
}: DateRangeFilterProps) => {
  const [isOpen, setIsOpen] = useState(false);

  const getCurrentPreset = () => {
    if (!startDate || !endDate) return null;

    const start = new Date(startDate);
    const end = new Date(endDate);

    return datePresets.find((preset) => {
      const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
      return (
        Math.abs(start.getTime() - presetStart.getTime()) <
          24 * 60 * 60 * 1000 &&
        Math.abs(end.getTime() - presetEnd.getTime()) < 24 * 60 * 60 * 1000
      );
    });
  };

  const currentPreset = getCurrentPreset();
  const hasCustomDates = startDate && endDate && !currentPreset;

  const getButtonText = () => {
    if (currentPreset) {
      return currentPreset.label;
    }
    if (hasCustomDates) {
      const start = format(new Date(startDate), "MMM d");
      const end = format(new Date(endDate), "MMM d");
      return `${start} - ${end}`;
    }
    return "Date Range";
  };

  return (
    <Popover onOpenChange={setIsOpen} open={isOpen}>
      <Popover.Trigger asChild>
        <Button
          color="tertiary"
          leftIcon={<CalendarIcon className="h-4 w-auto" />}
          size="sm"
          variant={startDate && endDate ? "solid" : "outline"}
        >
          {getButtonText()}
        </Button>
      </Popover.Trigger>

      <Popover.Content
        align="end"
        className="mr-0 w-96 bg-opacity-80 px-4 py-3 dark:bg-opacity-80"
      >
        <DateRangeSelector
          endDate={endDate}
          onDateChange={onDateChange}
          startDate={startDate}
        />
      </Popover.Content>
    </Popover>
  );
};
