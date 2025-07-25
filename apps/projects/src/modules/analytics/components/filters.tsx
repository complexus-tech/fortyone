"use client";
import { useState } from "react";
import {
  Box,
  Button,
  Command,
  Divider,
  Flex,
  Popover,
  Text,
  DatePicker,
} from "ui";
import {
  CalendarIcon,
  CheckIcon,
  TeamIcon,
  SprintsIcon,
  ObjectiveIcon,
  CloseIcon,
  ArrowDown2Icon,
} from "icons";
import { format, subDays, startOfDay, endOfDay } from "date-fns";
import { cn } from "lib";
import { useTeams } from "@/modules/teams/hooks/teams";
import { useTeamSprints } from "@/modules/sprints/hooks/team-sprints";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";

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

type FilterButtonProps = {
  label: string;
  icon: React.ReactNode;
  text: string;
  popover: React.ReactNode;
  isActive?: boolean;
};

const FilterButton = ({ label, icon, text, popover }: FilterButtonProps) => {
  const [isOpen, setIsOpen] = useState(false);

  return (
    <Box>
      <Text className="mb-1 text-[0.95rem]" color="muted">
        {label}
      </Text>
      <Popover onOpenChange={setIsOpen} open={isOpen}>
        <Popover.Trigger asChild>
          <Button
            className="min-w-28 justify-between rounded-[0.7rem] md:h-[2.3rem]"
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4" strokeWidth={3} />}
            variant="outline"
          >
            <Flex align="center" gap={2}>
              {icon}
              <span>{text}</span>
            </Flex>
          </Button>
        </Popover.Trigger>
        <Popover.Content
          align="start"
          className="w-[23rem] bg-opacity-80 pb-2.5 dark:bg-opacity-80"
        >
          {popover}
        </Popover.Content>
      </Popover>
    </Box>
  );
};

const DateRangeSelector = () => {
  const [customStartDate, setCustomStartDate] = useState<Date | undefined>();
  const [customEndDate, setCustomEndDate] = useState<Date | undefined>();

  const handlePresetSelect = (preset: DatePreset) => {
    const { startDate: presetStart, endDate: presetEnd } = preset.getDates();
    setCustomStartDate(presetStart);
    setCustomEndDate(presetEnd);
  };

  const handleCustomDateChange = (start?: Date, end?: Date) => {
    setCustomStartDate(start);
    setCustomEndDate(end);
  };

  const clearDates = () => {
    setCustomStartDate(undefined);
    setCustomEndDate(undefined);
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

const TeamsSelector = () => {
  const { data: teams = [] } = useTeams();
  const [selectedTeams, setSelectedTeams] = useState<string[]>([]);
  const [query, setQuery] = useState("");

  const toggleTeam = (teamId: string) => {
    setSelectedTeams((prev) =>
      prev.includes(teamId)
        ? prev.filter((id) => id !== teamId)
        : [...prev, teamId],
    );
  };

  const filteredTeams = teams.filter((team) =>
    team.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search teams..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setSelectedTeams([]);
          }}
        >
          <Flex align="center" gap={2}>
            <TeamIcon className="h-4 w-auto" />
            <Text>All teams</Text>
          </Flex>
          {selectedTeams.length === 0 && <CheckIcon className="h-4 w-auto" />}
        </Command.Item>
        {filteredTeams.map((team) => (
          <Command.Item
            className="justify-between"
            key={team.id}
            onSelect={() => {
              toggleTeam(team.id);
            }}
          >
            <Flex align="center" gap={2}>
              <Box
                className="size-3 rounded"
                style={{ backgroundColor: team.color }}
              />
              <Text>{team.name}</Text>
            </Flex>
            {selectedTeams.includes(team.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

const SprintsSelector = () => {
  const { data: sprints = [] } = useTeamSprints("");
  const [selectedSprints, setSelectedSprints] = useState<string[]>([]);
  const [query, setQuery] = useState("");

  const toggleSprint = (sprintId: string) => {
    setSelectedSprints((prev) =>
      prev.includes(sprintId)
        ? prev.filter((id) => id !== sprintId)
        : [...prev, sprintId],
    );
  };

  const filteredSprints = sprints.filter((sprint) =>
    sprint.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search sprints..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setSelectedSprints([]);
          }}
        >
          <Flex align="center" gap={2}>
            <SprintsIcon className="h-4 w-auto" />
            <Text>All sprints</Text>
          </Flex>
          {selectedSprints.length === 0 && <CheckIcon className="h-4 w-auto" />}
        </Command.Item>
        {filteredSprints.map((sprint) => (
          <Command.Item
            className="justify-between"
            key={sprint.id}
            onSelect={() => {
              toggleSprint(sprint.id);
            }}
          >
            <Flex align="center" gap={2}>
              <SprintsIcon className="h-4 w-auto" />
              <Text>
                {sprint.name}
                <Text as="span" color="muted">
                  {" "}
                  {format(new Date(sprint.startDate), "MMM d")} -{" "}
                  {format(new Date(sprint.endDate), "MMM d")}
                </Text>
              </Text>
            </Flex>
            {selectedSprints.includes(sprint.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

const ObjectivesSelector = () => {
  const { data: objectives = [] } = useObjectives();
  const [selectedObjectives, setSelectedObjectives] = useState<string[]>([]);
  const [query, setQuery] = useState("");

  const toggleObjective = (objectiveId: string) => {
    setSelectedObjectives((prev) =>
      prev.includes(objectiveId)
        ? prev.filter((id) => id !== objectiveId)
        : [...prev, objectiveId],
    );
  };

  const filteredObjectives = objectives.filter((objective) =>
    objective.name.toLowerCase().includes(query.toLowerCase()),
  );

  return (
    <Command>
      <Command.Input
        onValueChange={setQuery}
        placeholder="Search objectives..."
        value={query}
      />
      <Divider className="my-2" />
      <Command.Group>
        <Command.Item
          className="justify-between"
          onSelect={() => {
            setSelectedObjectives([]);
          }}
        >
          <Flex align="center" gap={2}>
            <ObjectiveIcon className="h-4 w-auto" />
            <Text>All objectives</Text>
          </Flex>
          {selectedObjectives.length === 0 && (
            <CheckIcon className="h-4 w-auto" />
          )}
        </Command.Item>
        {filteredObjectives.map((objective) => (
          <Command.Item
            className="justify-between"
            key={objective.id}
            onSelect={() => {
              toggleObjective(objective.id);
            }}
          >
            <Flex align="center" gap={2}>
              <ObjectiveIcon className="h-4 w-auto" />
              <Text>{objective.name}</Text>
            </Flex>
            {selectedObjectives.includes(objective.id) && (
              <CheckIcon className="h-4 w-auto" />
            )}
          </Command.Item>
        ))}
      </Command.Group>
    </Command>
  );
};

export const Filters = () => {
  // Mock state for demonstration - in real implementation, this would be managed by context or props
  const [dateRange, _setDateRange] = useState<{
    startDate?: string;
    endDate?: string;
  }>({});
  const [selectedTeams, _setSelectedTeams] = useState<string[]>([]);
  const [selectedSprints, _setSelectedSprints] = useState<string[]>([]);
  const [selectedObjectives, _setSelectedObjectives] = useState<string[]>([]);

  const { data: teams = [] } = useTeams();
  const { data: sprints = [] } = useTeamSprints("");
  const { data: objectives = [] } = useObjectives();

  const getDateRangeText = () => {
    if (!dateRange.startDate || !dateRange.endDate) return "Select dates";
    const start = format(new Date(dateRange.startDate), "MMM d");
    const end = format(new Date(dateRange.endDate), "MMM d");
    return `${start} → ${end}`;
  };

  const getTeamsText = () => {
    if (selectedTeams.length === 0) return "All";
    if (selectedTeams.length === 1) {
      const team = teams.find((t) => t.id === selectedTeams[0]);
      return team?.name || "Unknown";
    }
    return `${selectedTeams.length} teams`;
  };

  const getSprintsText = () => {
    if (selectedSprints.length === 0) return "All";
    if (selectedSprints.length === 1) {
      const sprint = sprints.find((s) => s.id === selectedSprints[0]);
      return sprint?.name || "Unknown";
    }
    return `${selectedSprints.length} sprints`;
  };

  const getObjectivesText = () => {
    if (selectedObjectives.length === 0) return "All";
    if (selectedObjectives.length === 1) {
      const objective = objectives.find((o) => o.id === selectedObjectives[0]);
      return objective?.name || "Unknown";
    }
    return `${selectedObjectives.length} objectives`;
  };

  return (
    <Flex align="end" gap={4}>
      <FilterButton
        icon={<CalendarIcon className="h-4 w-auto" />}
        isActive={Boolean(dateRange.startDate && dateRange.endDate)}
        label="Date Range"
        popover={<DateRangeSelector />}
        text={getDateRangeText()}
      />
      <FilterButton
        icon={<TeamIcon className="h-4 w-auto" />}
        isActive={selectedTeams.length > 0}
        label="Teams"
        popover={<TeamsSelector />}
        text={getTeamsText()}
      />
      <FilterButton
        icon={<SprintsIcon className="h-4 w-auto" />}
        isActive={selectedSprints.length > 0}
        label="Sprints"
        popover={<SprintsSelector />}
        text={getSprintsText()}
      />
      <FilterButton
        icon={<ObjectiveIcon className="h-4 w-auto" />}
        isActive={selectedObjectives.length > 0}
        label="Objectives"
        popover={<ObjectivesSelector />}
        text={getObjectivesText()}
      />
    </Flex>
  );
};
