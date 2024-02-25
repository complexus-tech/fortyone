import { Avatar, Badge, Box, Flex, ProgressBar, Text } from "ui";
import { CalendarIcon, SprintsIcon } from "icons";
import Link from "next/link";
import { cn } from "lib";
import { IssueStatusIcon, PriorityIcon, RowWrapper } from "@/components/ui";

export type SprintCardProps = {
  name: string;
  description: string;
};

export const SprintCard = ({ name, description }: SprintCardProps) => {
  const recentIssues = [
    { id: 1, title: "Issue with the login page" },
    { id: 2, title: "Issue with the login page" },
    { id: 3, title: "Issue with the login page" },
  ];
  return (
    <Link href="/projects/web/sprints/1">
      <RowWrapper className="block py-6">
        <Flex align="start" justify="between">
          <Box>
            <Text
              className="mb-2 flex items-center gap-1"
              fontSize="2xl"
              fontWeight="medium"
            >
              <SprintsIcon className="relative -left-[2px] h-6 w-auto" />
              {name}
            </Text>
            <Text className="max-w-lg" color="muted">
              {description || "-"}
            </Text>
          </Box>
          <Flex align="center" gap={1}>
            <Text>Lead:</Text>
            <Avatar name="Joseph Mukorivo" size="sm" />
            <Badge className="mx-4" color="tertiary" rounded="sm">
              5 days left
            </Badge>
            <CalendarIcon className="h-5 w-auto" /> Mar 15 - Mar 30
          </Flex>
        </Flex>
        <Box className="mt-3 grid grid-cols-3 divide-x divide-gray-100/70 dark:divide-dark-200">
          <Box className="pr-5">
            <Flex className="mb-2" justify="between">
              <Text fontSize="lg" fontWeight="medium">
                Progress
              </Text>
              <Text fontSize="lg" fontWeight="medium">
                75%
              </Text>
            </Flex>
            <ProgressBar progress={75} />
            <Box className="mt-2 grid gap-2">
              <Text
                className="flex items-center justify-between gap-1"
                color="muted"
              >
                <span className="inline-block size-2.5 rounded-full bg-primary" />
                Not started
                <span className="mx-2 h-[1px] flex-1 border-b border-dashed border-gray-100 dark:border-dark-200" />
                <span className="font-medium">6 &bull; 100%</span>
              </Text>
              <Text
                className="flex items-center justify-between gap-1"
                color="muted"
              >
                <span className="inline-block size-2.5 rounded-full bg-warning" />
                In progress
                <span className="mx-2 h-[1px] flex-1 border-b border-dashed border-gray-100 dark:border-dark-200" />
                <span className="font-medium">6 &bull; 100%</span>
              </Text>
              <Text
                className="flex items-center justify-between gap-1"
                color="muted"
              >
                <span className="inline-block size-2.5 rounded-full bg-success" />
                Completed
                <span className="mx-2 h-[1px] flex-1 border-b border-dashed border-gray-100 dark:border-dark-200" />
                <span className="font-medium">6 &bull; 100%</span>
              </Text>
            </Box>
          </Box>
          <Box className="px-5">
            <Text className="mb-2" fontSize="lg" fontWeight="medium">
              Burndown chart
            </Text>
            <Box className="h-28 border-b border-l border-gray-100/70 dark:border-dark-200" />
          </Box>
          <Box className="pl-5">
            <Text className="mb-2" fontSize="lg" fontWeight="medium">
              Recent issues
            </Text>
            {recentIssues.map(({ id, title }, idx) => (
              <RowWrapper
                className={cn("px-0 py-2", {
                  "border-b-0": idx === recentIssues.length - 1,
                })}
                key={id}
              >
                <Flex align="center" gap={1}>
                  <IssueStatusIcon className="h-[1.1rem] w-auto" />
                  <Text
                    className="shrink-0 text-[0.95rem]"
                    color="muted"
                    fontWeight="medium"
                  >
                    WEB-{id}
                  </Text>
                  <Text className="mr-1 line-clamp-1">
                    {title} Lorem ipsum dolor sit amet.
                  </Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <PriorityIcon className="h-[1.1rem] w-auto" priority="High" />
                  <Avatar name="Joseph Mukorivo" size="xs" />
                </Flex>
              </RowWrapper>
            ))}
          </Box>
        </Box>
      </RowWrapper>
    </Link>
  );
};
