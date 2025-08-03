"use client";
import { parseAsIsoDateTime, useQueryStates } from "nuqs";
import { Box, Button, DatePicker, Divider, Flex, Text } from "ui";
import { CalendarIcon, CloseIcon } from "icons";
import { format } from "date-fns";
import { cn } from "lib";
import { FilterButton } from "./filter-button";
import { datePresets, getDefaultDateRange, type DatePreset } from "./types";

const DateRangeSelector = () => {
  // Default to last 30 days (matching backend default)
  const defaultDates = getDefaultDateRange();

  const [filters, setFilters] = useQueryStates({
    startDate: parseAsIsoDateTime.withDefault(defaultDates.startDate),
    endDate: parseAsIsoDateTime.withDefault(defaultDates.endDate),
  });

  const customStartDate = filters.startDate || undefined;
  const customEndDate = filters.endDate || undefined;

  const handlePresetSelect = (preset: DatePreset) => {
    const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
    setFilters({
      startDate: presetStart,
      endDate: presetEnd,
    });
  };

  const handleCustomDateChange = (start?: Date, end?: Date) => {
    setFilters({
      startDate: start !== undefined ? start : filters.startDate,
      endDate: end !== undefined ? end : filters.endDate,
    });
  };

  const clearDates = () => {
    setFilters({
      startDate: null,
      endDate: null,
    });
  };

  const getCurrentPreset = () => {
    if (!customStartDate || !customEndDate) return null;

    return datePresets.find((preset) => {
      const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
      return (
        Math.abs(customStartDate.getTime() - presetStart.getTime()) <
          24 * 60 * 60 * 1000 &&
        Math.abs(customEndDate.getTime() - presetEnd.getTime()) <
          24 * 60 * 60 * 1000
      );
    });
  };

  const currentPreset = getCurrentPreset();

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
              rightIcon={
                <CloseIcon
                  aria-label="Clear start date"
                  className="h-4 w-auto"
                  onClick={(e) => {
                    e.stopPropagation();
                    handleCustomDateChange(undefined, customEndDate);
                  }}
                  role="button"
                />
              }
              size="sm"
              variant="outline"
            >
              {format(customStartDate, "MMM d, yyyy")}
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
        <Button
          color="tertiary"
          leftIcon={<CloseIcon className="h-4 w-auto" />}
          onClick={clearDates}
          size="sm"
          variant="outline"
        >
          Clear
        </Button>
      </Flex>
    </Box>
  );
};

export const DateRangeFilter = () => {
  // Default to last 30 days (matching backend default)
  const defaultDates = getDefaultDateRange();

  const [filters] = useQueryStates({
    startDate: parseAsIsoDateTime.withDefault(defaultDates.startDate),
    endDate: parseAsIsoDateTime.withDefault(defaultDates.endDate),
  });

  const getDateRangeText = () => {
    const start = format(filters.startDate, "MMM d");
    const end = format(filters.endDate, "MMM d");
    return `${start} — ${end}`;
  };

  return (
    <FilterButton
      icon={<CalendarIcon className="h-4 w-auto" />}
      isActive
      label="Date Range"
      popover={<DateRangeSelector />}
      text={getDateRangeText()}
    />
  );
};
