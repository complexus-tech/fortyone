"use client";
import { Box, Button, Container, Divider, Text, DatePicker, Avatar } from "ui";
import { useState, type ReactNode } from "react";
import { addDays } from "date-fns";
import type { DateRange } from "react-day-picker";
import { CalendarIcon } from "icons";
import {
  PrioritiesMenu,
  StatusesMenu,
  AssigneesMenu,
  ModulesMenu,
  SprintsMenu,
  StoryStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { Labels } from "@/components/ui/story/labels";
import { AddLinks, OptionsHeader } from ".";
import { DetailedStory } from "../../types";

const Option = ({ label, value }: { label: string; value: ReactNode }) => {
  return (
    <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
      <Text
        className="flex items-center gap-1 truncate"
        color="muted"
        fontWeight="medium"
      >
        {label}
      </Text>
      {value}
    </Box>
  );
};

export const Options = ({ story }: { story: DetailedStory }) => {
  const { priority, startDate, endDate } = story;
  const [date, setDate] = useState<DateRange | undefined>({
    from: new Date(2022, 0, 20),
    to: addDays(new Date(2022, 0, 20), 20),
  });
  return (
    <Box className="h-full overflow-y-auto bg-gradient-to-br from-white via-gray-50/50 to-gray-50 pb-6 dark:from-dark-200/50 dark:to-dark">
      <OptionsHeader />

      <Container className="px-8 pt-4 text-gray-300/90">
        <Text className="mb-5" fontWeight="semibold">
          Properties
        </Text>
        <Option
          label="Status"
          value={
            <StatusesMenu>
              <StatusesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<StoryStatusIcon />}
                  type="button"
                  variant="naked"
                >
                  Backlog
                </Button>
              </StatusesMenu.Trigger>
              <StatusesMenu.Items />
            </StatusesMenu>
          }
        />
        <Option
          label="Priority"
          value={
            <PrioritiesMenu>
              <PrioritiesMenu.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<PriorityIcon priority={priority} />}
                  type="button"
                  variant="naked"
                >
                  {priority}
                </Button>
              </PrioritiesMenu.Trigger>
              <PrioritiesMenu.Items />
            </PrioritiesMenu>
          }
        />
        <Option
          label="Assignee"
          value={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  className="font-medium"
                  color="tertiary"
                  leftIcon={
                    <Avatar
                      name="Joseph Mukorivo"
                      size="xs"
                      src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                    />
                  }
                  type="button"
                  variant="naked"
                >
                  josemukorivo
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items />
            </AssigneesMenu>
          }
        />
        <Option label="Labels" value={<Labels />} />
        <Option
          label="Start date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                  variant="naked"
                >
                  Sep 27, 2024
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar
                defaultMonth={date?.from}
                initialFocus
                mode="range"
                numberOfMonths={2}
                onSelect={setDate}
                selected={date}
              />
            </DatePicker>
          }
        />
        <Option
          label="Due date"
          value={
            <DatePicker>
              <DatePicker.Trigger>
                <Button
                  color="tertiary"
                  leftIcon={<CalendarIcon className="h-[1.15rem] w-auto" />}
                  variant="naked"
                >
                  Sep 27, 2024
                </Button>
              </DatePicker.Trigger>
              <DatePicker.Calendar />
            </DatePicker>
          }
        />
        <Option label="Sprint" value={<SprintsMenu />} />
        <Option label="Module" value={<ModulesMenu />} />
        <Option
          label="Parent"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocking"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Blocked by"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />
        <Option
          label="Related to"
          value={
            <Button color="tertiary" variant="naked">
              None
            </Button>
          }
        />

        <Divider className="my-4" />
        <AddLinks />
      </Container>
    </Box>
  );
};
