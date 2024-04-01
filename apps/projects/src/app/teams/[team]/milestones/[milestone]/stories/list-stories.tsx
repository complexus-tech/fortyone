"use client";
import {
  Avatar,
  Badge,
  Box,
  BreadCrumbs,
  Button,
  Divider,
  Flex,
  Menu,
  ProgressBar,
  Tabs,
  Text,
} from "ui";
import {
  MilestonesIcon,
  MoreVerticalIcon,
  LinkIcon,
  DeleteIcon,
  StarIcon,
  EditIcon,
  StoryIcon,
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
import { HeaderContainer } from "@/components/shared";
import type { StoriesLayout } from "@/components/ui";
import {
  LayoutSwitcher,
  StoriesViewOptionsButton,
  SideDetailsSwitch,
  BoardDividedPanel,
  StoriesBoard,
  RowWrapper,
  StoryStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { useLocalStorage } from "@/hooks";
import type { Story, StoryStatus } from "@/types/story";

export const ListStories = ({
  stories,
  statuses,
}: {
  stories: Story[];
  statuses: StoryStatus[];
}) => {
  const [layout, setLayout] = useLocalStorage<StoriesLayout>(
    "objective:milestones:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "objective:milestones:isExpanded",
    true,
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
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Engineering",
              icon: "🚀",
            },
            {
              name: "Milestones",
              icon: <MilestonesIcon className="h-4 w-auto" />,
            },
            {
              name: "Stories",
              icon: <StoryIcon className="h-[1.1rem] w-auto" strokeWidth={2} />,
            },
          ]}
        />
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <StoriesViewOptionsButton />
          <span className="text-gray-200 dark:text-dark-100">|</span>
          <SideDetailsSwitch
            isExpanded={isExpanded}
            setIsExpanded={setIsExpanded}
          />
        </Flex>
      </HeaderContainer>
      <BoardDividedPanel autoSaveId="my-stories:divided-panel">
        <BoardDividedPanel.MainPanel>
          <StoriesBoard layout={layout} statuses={statuses} stories={stories} />
        </BoardDividedPanel.MainPanel>
        <BoardDividedPanel.SideBar isExpanded={isExpanded}>
          <Box className="py-6">
            <Box className="px-6">
              <Flex align="center" justify="between">
                <Text className="flex items-center gap-1.5" fontSize="lg">
                  <MilestonesIcon className="relative -top-px h-[1.4rem] w-auto" />
                  Milestone 1
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
                    <Menu.Group>
                      <Text className="mt-1" color="muted">
                        Manage sprint
                      </Text>
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
                <Badge rounded="sm">Current</Badge>
                <Badge color="tertiary" rounded="sm">
                  21 Feb - 21 Mar
                </Badge>
              </Flex>
              <Flex
                align="center"
                className="mb-2 mt-3"
                gap={2}
                justify="between"
              >
                <Text>Milestone Progress</Text>
                <Text>75%</Text>
              </Flex>
              <ProgressBar className="h-1.5" progress={75} />
            </Box>
            <Divider className="mb-6 mt-8" />
            <Box className="px-6">
              <Text as="h3" fontSize="lg" fontWeight="medium">
                Burndown chart
              </Text>
              <Text className="mb-4" color="muted">
                Burndown chart shows the amount of work remaining in the sprint.
              </Text>
              <ResponsiveContainer height={150} width="100%">
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
                    fill="#002F61"
                    stackId="1"
                    stroke="#002F61"
                    type="monotone"
                  />
                  <Area
                    dataKey="amt"
                    fill="#eab308"
                    stackId="1"
                    stroke="#eab308"
                    type="monotone"
                  />
                </AreaChart>
              </ResponsiveContainer>
            </Box>
            <Divider className="my-8" />
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
                    <RowWrapper className="px-1 py-2" key={idx}>
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
                    <RowWrapper className="px-1 py-2" key={idx}>
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
                    <RowWrapper className="px-1 py-2" key={idx}>
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
                    <RowWrapper className="px-1 py-2" key={idx}>
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
                  No documents for this milestone
                </Text>
              </Flex>
            </Box>
          </Box>
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
