import { Avatar, Badge, Box, Flex, ProgressBar, Text } from "ui";
import { CalendarIcon, SprintsIcon } from "icons";
import Link from "next/link";
import { cn } from "lib";
import {
  AreaChart,
  Area,
  XAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import { StoryStatusIcon, PriorityIcon, RowWrapper } from "@/components/ui";

export type SprintCardProps = {
  name: string;
  description: string;
};

export const SprintCard = ({ name, description }: SprintCardProps) => {
  const recentStories = [
    { id: 1, title: "Story with the login page" },
    { id: 2, title: "Story with the login page" },
    { id: 3, title: "Story with the login page" },
  ];

  const data = [
    {
      name: "Feb 1",
      uv: 4000,
      pv: 2400,
      amt: 2400,
    },
    {
      name: "Feb 4",
      uv: 3000,
      pv: 1398,
      amt: 2210,
    },
    {
      name: "Feb 8",
      uv: 2000,
      pv: 9800,
      amt: 2290,
    },
    {
      name: "Feb 12",
      uv: 2780,
      pv: 3908,
      amt: 2000,
    },
    {
      name: "Feb 20",
      uv: 1890,
      pv: 4800,
      amt: 2181,
    },
  ];
  return (
    <Link href="/objectives/web/sprints/1/stories">
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
            <ProgressBar className="h-1.5" progress={75} />
            <Box className="mt-2.5 grid gap-2.5">
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
            <ResponsiveContainer height={100} width="100%">
              <AreaChart
                data={data}
                margin={{
                  top: 0,
                  right: 0,
                  left: -13,
                  bottom: -12,
                }}
              >
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" fontSize="0.9rem" />
                <Tooltip />
                <Area
                  dataKey="uv"
                  fill="#8884d8"
                  stackId="1"
                  stroke="#8884d8"
                  type="monotone"
                />
                <Area
                  dataKey="pv"
                  fill="#82ca9d"
                  stackId="1"
                  stroke="#82ca9d"
                  type="monotone"
                />
                <Area
                  dataKey="amt"
                  fill="#ffc658"
                  stackId="1"
                  stroke="#ffc658"
                  type="monotone"
                />
              </AreaChart>
            </ResponsiveContainer>
          </Box>
          <Box className="pl-5">
            <Text className="mb-2" fontSize="lg" fontWeight="medium">
              Recent stories
            </Text>
            {recentStories.map(({ id, title }, idx) => (
              <RowWrapper
                className={cn("px-0 py-2", {
                  "border-b-0": idx === recentStories.length - 1,
                })}
                key={id}
              >
                <Flex align="center" gap={1}>
                  <StoryStatusIcon className="h-[1.1rem] w-auto" />
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
