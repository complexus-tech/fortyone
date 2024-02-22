"use client";
import Link from "next/link";
import { Box, Flex, Button, Text, Avatar, DatePicker } from "ui";
import {
  PriorityIcon,
  StatusesMenu,
  PrioritiesMenu,
  IssueStatusIcon,
  IssueContextMenu,
  AssigneesMenu,
} from "@/components/ui";
import type { Issue } from "@/types/issue";
import { CalendarIcon, CalendarPlusIcon, TagsIcon } from "icons";

export const Card = ({ issue }: { issue: Issue }) => {
  return (
    <IssueContextMenu>
      <Box className="w-[340px] cursor-pointer select-none rounded-lg border border-gray-100/80 bg-white p-4 backdrop-blur transition duration-200 ease-linear hover:bg-white/50 dark:border-dark-100/70 dark:bg-dark-200/50 dark:hover:bg-dark-200/90">
        <Flex align="center" className="mb-2" gap={2} justify="between">
          <Link className="flex-1" href="/projects/web/issues/test-123-issue">
            <Text className="text-[0.9rem]" color="muted" fontWeight="medium">
              COMP-123
            </Text>
          </Link>
          <AssigneesMenu>
            <AssigneesMenu.Trigger>
              <button className="block" type="button">
                <Avatar
                  name="Joseph Mukorivo"
                  size="xs"
                  src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                />
              </button>
            </AssigneesMenu.Trigger>
            <AssigneesMenu.Items />
          </AssigneesMenu>
        </Flex>
        <Link className="flex-1" href="/projects/web/issues/test-123-issue">
          <Text className="mb-2.5 line-clamp-2">{issue.title}</Text>
        </Link>
        <Flex gap={1} wrap>
          <StatusesMenu>
            <StatusesMenu.Trigger>
              <Button
                className="bg-white dark:border-dark-100 dark:bg-dark-300/50"
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
                className="bg-white dark:border-dark-100 dark:bg-dark-300/50"
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
                className="bg-white px-2 text-sm dark:border-dark-100 dark:bg-dark-300/50"
                color="tertiary"
                leftIcon={<CalendarPlusIcon className="h-4 w-auto" />}
                size="xs"
                type="button"
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
                className="bg-white px-2 text-sm dark:border-dark-100 dark:bg-dark-300/50"
                color="tertiary"
                leftIcon={<CalendarIcon className="h-4 w-auto" />}
                size="xs"
                variant="outline"
              >
                Sep 21
              </Button>
            </DatePicker.Trigger>
            <DatePicker.Calendar />
          </DatePicker>
          <Button
            className="bg-white dark:border-dark-100 dark:bg-dark-300/50"
            color="tertiary"
            leftIcon={<TagsIcon className="h-4 w-auto" />}
            size="xs"
            type="button"
            variant="outline"
          >
            3 labels
          </Button>
        </Flex>
      </Box>
    </IssueContextMenu>
  );
};
