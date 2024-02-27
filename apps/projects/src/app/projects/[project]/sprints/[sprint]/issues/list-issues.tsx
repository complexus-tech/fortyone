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
  PreferencesIcon,
  IssuesIcon,
  SprintsIcon,
  MoreVerticalIcon,
  LinkIcon,
  DeleteIcon,
  StarIcon,
  EditIcon,
} from "icons";
import {
  AreaChart,
  Area,
  XAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from "recharts";
import type { Issue, IssueStatus } from "@/types/issue";
import { useLocalStorage } from "@/hooks";
import type { IssuesLayout } from "@/components/ui";
import {
  IssuesBoard,
  LayoutSwitcher,
  SideDetailsSwitch,
  BoardDividedPanel,
  RowWrapper,
  IssueStatusIcon,
  PriorityIcon,
} from "@/components/ui";
import { HeaderContainer } from "@/components/layout";

export const ListIssues = ({
  issues,
  statuses,
}: {
  issues: Issue[];
  statuses: IssueStatus[];
}) => {
  const [layout, setLayout] = useLocalStorage<IssuesLayout>(
    "project:sprints:layout",
    "kanban",
  );
  const [isExpanded, setIsExpanded] = useLocalStorage<boolean>(
    "project:sprints:isExpanded",
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
              name: "Web design",
              icon: "🚀",
            },
            {
              name: "Sprints",
              icon: <SprintsIcon className="h-4 w-auto" />,
            },
            {
              name: "Issues",
              icon: <IssuesIcon className="h-[1.1rem] w-auto" />,
            },
          ]}
        />
        <Flex align="center" gap={2}>
          <LayoutSwitcher layout={layout} setLayout={setLayout} />
          <Button
            color="tertiary"
            leftIcon={<PreferencesIcon className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <span className="text-gray-200 dark:text-dark-100">|</span>
          <SideDetailsSwitch
            isExpanded={isExpanded}
            setIsExpanded={setIsExpanded}
          />
        </Flex>
      </HeaderContainer>
      <BoardDividedPanel autoSaveId="my-issues:divided-panel">
        <BoardDividedPanel.MainPanel>
          <IssuesBoard issues={issues} layout={layout} statuses={statuses} />
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
              <Text className="my-6" color="muted">
                Lorem, ipsum dolor sit amet consectetur adipisicing elit. Nihil,
                hic?
              </Text>
              <Flex align="center" gap={2} justify="between">
                <Flex align="center" gap={2}>
                  <Badge color="success" rounded="sm">
                    Current
                  </Badge>
                  <Badge color="tertiary" rounded="sm">
                    21 Feb - 21 Mar
                  </Badge>
                </Flex>
                <Button
                  color="tertiary"
                  rightIcon={<Avatar name="Joseph Mukorivo" size="xs" />}
                  size="sm"
                  variant="naked"
                >
                  josemukorivo
                </Button>
              </Flex>
              <Flex
                align="center"
                className="mb-2 mt-3"
                gap={2}
                justify="between"
              >
                <Text>Sprint Progress</Text>
                <Text>75%</Text>
              </Flex>
              <ProgressBar progress={75} />
            </Box>
            <Divider className="mb-6 mt-8" />
            <Box className="px-6">
              <Text fontSize="lg" fontWeight="medium">
                Burndown chart
              </Text>
              <Text className="mb-4" color="muted">
                Lorem ipsum dolor sit amet consectetur adipisicing elit. Nihil,
                hic?
              </Text>
              <ResponsiveContainer height={120} width="100%">
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
              <Text className="mt-4" fontSize="lg" fontWeight="medium">
                Velocity chart
              </Text>
              <Text className="mb-4" color="muted">
                Lorem ipsum dolor sit amet consectetur adipisicing elit. Nihil,
                hic?
              </Text>
              <ResponsiveContainer height={120} width="100%">
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
                        <IssueStatusIcon />
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
          </Box>
        </BoardDividedPanel.SideBar>
      </BoardDividedPanel>
    </>
  );
};
