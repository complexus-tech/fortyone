"use client";
import { parseAsIsoDateTime, useQueryStates } from "nuqs";
import { Box, Button, DatePicker, Divider, Flex, Text } from "ui";
import { CalendarIcon } from "icons";
import { format, subMonths } from "date-fns";
import { cn } from "lib";
import { FilterButton } from "./filter-button";
import { datePresets, getDefaultDateRange, type DatePreset } from "./types";

const DateRangeSelector = () => {
  const defaultDates = getDefaultDateRange();
  const [filters, setFilters] = useQueryStates({
    startDate: parseAsIsoDateTime.withDefault(defaultDates.startDate),
    endDate: parseAsIsoDateTime.withDefault(defaultDates.endDate),
  });

  const handlePresetSelect = (preset: DatePreset) => {
    const { startDate, endDate } = preset.getDates();
    setFilters({ startDate, endDate });
  };

  const handleDateChange = (start?: Date, end?: Date) => {
    setFilters({
      startDate: start ?? filters.startDate,
      endDate: end ?? filters.endDate,
    });
  };

  const currentPreset = datePresets.find((preset) => {
    const { startDate, endDate } = preset.getDates();
    return (
      filters.startDate.getTime() === startDate.getTime() &&
      filters.endDate.getTime() === endDate.getTime()
    );
  });

  return (
    <Box className="px-4 py-1">
      <Text className="mb-1" color="muted" fontWeight="medium">
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

      <Divider className="my-2" />

      <Text className="mb-1" color="muted" fontWeight="medium">
        Custom Range
      </Text>
      <Flex gap={2} wrap>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="gap-2 px-2"
              color="tertiary"
              leftIcon={<CalendarIcon className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            >
              {format(filters.startDate, "MMM d, yyyy")}
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            fromDate={subMonths(new Date(), 6)}
            onDayClick={(date: Date) => {
              handleDateChange(date, filters.endDate);
            }}
            selected={filters.startDate}
          />
        </DatePicker>

        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="gap-2 px-2"
              color="tertiary"
              leftIcon={<CalendarIcon className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            >
              {format(filters.endDate, "MMM d, yyyy")}
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar
            fromDate={filters.startDate}
            onDayClick={(date: Date) => {
              handleDateChange(filters.startDate, date);
            }}
            selected={filters.endDate}
            toDate={new Date()}
          />
        </DatePicker>
      </Flex>
    </Box>
  );
};

export const DateRangeFilter = () => {
  const defaultDates = getDefaultDateRange();
  const [filters] = useQueryStates({
    startDate: parseAsIsoDateTime.withDefault(defaultDates.startDate),
    endDate: parseAsIsoDateTime.withDefault(defaultDates.endDate),
  });

  return (
    <FilterButton
      icon={<CalendarIcon className="h-4 w-auto" />}
      isActive
      label="Date Range"
      popover={<DateRangeSelector />}
      text={`${format(filters.startDate, "MMM d")} - ${format(filters.endDate, "MMM d")}`}
    />
  );
};
