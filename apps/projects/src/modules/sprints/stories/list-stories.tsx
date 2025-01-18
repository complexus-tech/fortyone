"use client";
import {
  Avatar,
  Badge,
  Box,
  Button,
  Divider,
  Flex,
  Menu,
  ProgressBar,
  Tabs,
  Text,
} from "ui";
import {
  SprintsIcon,
  MoreVerticalIcon,
  LinkIcon,
  DeleteIcon,
  StarIcon,
  EditIcon,
  DocsIcon,
} from "icons";
import {
  AreaChart,
  Area,
  XAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type { StoriesLayout } from "@/components/ui";
import {
  BoardDividedPanel,
  RowWrapper,
  StoryStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story } from "@/modules/stories/types";
import { Header } from "./header";
import { SprintStoriesProvider } from "./provider";
import { AllStories } from "./all-stories";

export const ListSprintStories = ({ stories }: { stories: Story[] }) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:sprints:layout",
    "list",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "objective:sprints:isExpanded",
    false,
  );

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
    <SprintStoriesProvider>
      <Header
        allStories={stories.length}
        isExpanded={isExpanded}
        layout={layout}
        setIsExpanded={setIsExpanded}
        setLayout={setLayout}
      />
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <AllStories layout={layout} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Box className="py-6">
            <Box className="px-6">
              <Flex align="center" justify="between">
                <Text className="flex items-center gap-1.5" fontSize="lg">
                  <SprintsIcon className="relative -top-px h-[1.4rem] w-auto" />
                  Sprint 1
                </Text>
                <Menu>
                  <Menu.Button>
                    <Button
                      className="aspect-square"
                      color="tertiary"
                      leftIcon={<MoreVerticalIcon className="h-5 w-auto" />}
                      size="sm"
                      variant="outline"
                    >
                      <span className="sr-only">More options</span>
                    </Button>
                  </Menu.Button>
                  <Menu.Items align="end" className="w-64">
                    <Menu.Group className="px-4">
                      <Text className="mt-1">Manage Sprint</Text>
                    </Menu.Group>
                    <Menu.Separator className="mb-2" />
                    <Menu.Group>
                      <Menu.Item>
                        <EditIcon className="h-[1.1rem] w-auto" />
                        Edit sprint
                      </Menu.Item>
                      <Menu.Item>
                        <LinkIcon className="h-5 w-auto" />
                        Copy link
                      </Menu.Item>
                      <Menu.Item>
                        <StarIcon className="h-[1.15rem] w-auto" />
                        Favorite
                      </Menu.Item>
                      <Menu.Item>
                        <DeleteIcon className="h-5 w-auto" />
                        Delete sprint
                      </Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              </Flex>
              <Flex align="center" className="my-4" gap={2}>
                <Badge>Current</Badge>
                <Badge color="tertiary">21 Feb - 21 Mar</Badge>
              </Flex>
              <Flex
                align="center"
                className="mb-2 mt-3"
                gap={2}
                justify="between"
              >
                <Text>Sprint Progress</Text>
                <Text>40%</Text>
              </Flex>
              <ProgressBar className="h-1" progress={40} />
            </Box>
            <Divider className="mb-8 mt-6" />
            <Box className="px-6">
              <Tabs defaultValue="assignees">
                <Tabs.List className="mx-0 mb-1">
                  <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
                  <Tabs.Tab value="labels">Labels</Tabs.Tab>
                  <Tabs.Tab value="status">Status</Tabs.Tab>
                  <Tabs.Tab value="priority">Priority</Tabs.Tab>
                </Tabs.List>
                <Tabs.Panel value="assignees">
                  {new Array(4).fill(1).map((_, idx) => (
                    <RowWrapper className="px-1 py-2 md:px-1" key={idx}>
                      <Flex align="center" gap={2}>
                        <Avatar name="Joseph Mukorivo" size="xs" />
                        <Text color="muted">josemukorivo</Text>
                      </Flex>
                      <Flex align="center" gap={2}>
                        <ProgressBar className="w-20" progress={25} />
                        <Text color="muted">25% of 4</Text>
                      </Flex>
                    </RowWrapper>
                  ))}
                </Tabs.Panel>
                <Tabs.Panel value="labels">
                  {new Array(4).fill(1).map((_, idx) => (
                    <RowWrapper className="px-1 py-2 md:px-1" key={idx}>
                      <Flex align="center" gap={2}>
                        <span className="block size-2 rounded-full bg-primary" />
                        <Text color="muted">Feature</Text>
                      </Flex>
                      <Flex align="center" gap={2}>
                        <ProgressBar className="w-20" progress={25} />
                        <Text color="muted">25% of 4</Text>
                      </Flex>
                    </RowWrapper>
                  ))}
                </Tabs.Panel>
                <Tabs.Panel value="status">
                  {new Array(4).fill(1).map((_, idx) => (
                    <RowWrapper className="px-1 py-2 md:px-1" key={idx}>
                      <Flex align="center" gap={2}>
                        <StoryStatusIcon />
                        <Text color="muted">Backlog</Text>
                      </Flex>
                      <Flex align="center" gap={2}>
                        <ProgressBar className="w-20" progress={25} />
                        <Text color="muted">25% of 4</Text>
                      </Flex>
                    </RowWrapper>
                  ))}
                </Tabs.Panel>
                <Tabs.Panel value="priority">
                  {new Array(4).fill(1).map((_, idx) => (
                    <RowWrapper className="px-1 py-2 md:px-1" key={idx}>
                      <Flex align="center" gap={2}>
                        <PriorityIcon priority="High" />
                        <Text color="muted">High</Text>
                      </Flex>
                      <Flex align="center" gap={2}>
                        <ProgressBar className="w-20" progress={25} />
                        <Text color="muted">25% of 4</Text>
                      </Flex>
                    </RowWrapper>
                  ))}
                </Tabs.Panel>
              </Tabs>
            </Box>
            <Divider className="mb-6 mt-8" />
            <Box className="px-6">
              <Text as="h3" fontSize="lg" fontWeight="medium">
                Burndown chart
              </Text>
              <Text className="mb-4" color="muted">
                Burndown chart shows the amount of work remaining in the sprint.
              </Text>
              {/* <ResponsiveContainer height={150} width="100%">
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
                  <defs>
                    <linearGradient id="colorUv" x1="0" x2="0" y1="0" y2="1">
                      <stop offset="5%" stopColor="#eab308" stopOpacity={0.6} />
                      <stop offset="95%" stopColor="#eab308" stopOpacity={0} />
                    </linearGradient>
                  </defs>
                  <defs>
                    <linearGradient id="colorAmt" x1="0" x2="0" y1="0" y2="1">
                      <stop offset="5%" stopColor="#002F61" stopOpacity={0.6} />
                      <stop offset="95%" stopColor="#002F61" stopOpacity={0} />
                    </linearGradient>
                  </defs>

                  <XAxis dataKey="name" fontSize="0.9rem" />
                  <Tooltip />
                  <Area
                    dataKey="uv"
                    fill="url(#colorUv)"
                    fillOpacity={1}
                    stackId="1"
                    stroke="#eab308"
                    strokeDasharray="5 5"
                    type="monotone"
                  />
                  <Area
                    dataKey="amt"
                    fill="url(#colorAmt)"
                    fillOpacity={1}
                    stackId="1"
                    stroke="#002F61"
                    type="monotone"
                  />
                </AreaChart>
              </ResponsiveContainer> */}
            </Box>
            <Divider className="mb-6 mt-8" />
            <Box className="px-6">
              <Text as="h3" className="mb-3" fontSize="lg" fontWeight="medium">
                Documents
              </Text>
              <Flex
                align="center"
                className="rounded-xl bg-gray-50/80 px-4 py-10 dark:bg-dark-200/20"
                direction="column"
                justify="center"
              >
                <DocsIcon className="h-20 w-auto rotate-12" strokeWidth={1} />
                <Text className="mt-4" color="muted">
                  No documents for this sprint
                </Text>
              </Flex>
            </Box>
          </Box>
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </SprintStoriesProvider>
  );
};
