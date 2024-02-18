"use client";
import { Box, Flex, Button, Text, Avatar, DatePicker } from "ui";
import { CalendarCheck2, Tags, Calendar } from "lucide-react";
import {
  PriorityIcon,
  StatusesMenu,
  PrioritiesMenu,
  IssueStatusIcon,
} from "@/components/ui";
import type { Issue } from "@/types/issue";

export const Card = ({ issue }: { issue: Issue }) => {
  return (
    <Box
      className="w-[340px] cursor-pointer select-none rounded-lg border border-gray-100/80 bg-white p-4 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:border-dark-100/70 dark:bg-dark-200/50 dark:hover:bg-dark-200/90"
      draggable
    >
      <Flex align="center" className="mb-2" gap={2} justify="between">
        <Text className="text-[0.93rem]" color="muted" fontWeight="medium">
          COMP-123
        </Text>
        <Avatar
          name="Joseph Mukorivo"
          size="xs"
          src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
        />
      </Flex>
      <Text className="mb-2.5 line-clamp-2">{issue.title}</Text>
      <Flex gap={1} wrap>
        <StatusesMenu>
          <StatusesMenu.Trigger>
            <Button
              className="bg-white dark:bg-dark-300/50"
              color="tertiary"
              leftIcon={
                <IssueStatusIcon
                  className="relative left-0.5 h-4 w-auto"
                  status={issue.status}
                />
              }
              size="xs"
              type="button"
              variant="outline"
            >
              <span className="sr-only">{issue.status}</span>
            </Button>
          </StatusesMenu.Trigger>
          <StatusesMenu.Items status={issue.status} />
        </StatusesMenu>
        <PrioritiesMenu>
          <PrioritiesMenu.Trigger>
            <Button
              className="bg-white dark:bg-dark-300/50"
              color="tertiary"
              leftIcon={
                <PriorityIcon
                  className="relative left-0.5 h-4 w-auto"
                  priority={issue.priority}
                />
              }
              size="xs"
              type="button"
              variant="outline"
            >
              <span className="sr-only">{issue.priority}</span>
            </Button>
          </PrioritiesMenu.Trigger>
          <PrioritiesMenu.Items priority={issue.priority} />
        </PrioritiesMenu>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="bg-white px-2 text-sm dark:bg-dark-300/50"
              color="tertiary"
              leftIcon={<Calendar className="h-4 w-auto" />}
              size="xs"
              variant="outline"
            >
              Start
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar />
        </DatePicker>
        <DatePicker>
          <DatePicker.Trigger>
            <Button
              className="bg-white px-2 text-sm dark:bg-dark-300/50"
              color="tertiary"
              leftIcon={<CalendarCheck2 className="h-4 w-auto" />}
              size="xs"
              variant="outline"
            >
              Sep 21
            </Button>
          </DatePicker.Trigger>
          <DatePicker.Calendar />
        </DatePicker>
        <Button
          className="bg-white dark:bg-dark-300/50"
          color="tertiary"
          leftIcon={<Tags className="h-4" />}
          size="xs"
          type="button"
          variant="outline"
        >
          3 labels
        </Button>
      </Flex>
    </Box>
  );
};
