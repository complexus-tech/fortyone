"use client";
import {
  Box,
  BreadCrumbs,
  Button,
  Container,
  Flex,
  Tabs,
  Text,
  Avatar,
  Menu,
} from "ui";
import {
  ArrowUpRight,
  Calendar,
  ChevronDown,
  Columns3,
  SlidersHorizontal,
} from "lucide-react";
import HeatMap from "@uiw/react-heat-map";
import { cn } from "lib";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import { NewIssueButton, RowWrapper } from "@/components/ui";
import { IssueContextMenu } from "@/components/ui/issue/context-menu";
import { StatusesMenu } from "@/components/ui/issue/statuses-menu";
import { PrioritiesMenu } from "@/components/ui/issue/priorities-menu";

type ActivityProps = {
  id: number;
  user: string;
  action: string;
  prevValue: string;
  newValue: string;
  timestamp: string;
};
const Activity = ({
  user,
  action,
  prevValue,
  newValue,
  timestamp,
}: ActivityProps) => (
  <Flex align="center" className="z-[1]" gap={1}>
    <Box className="flex aspect-square items-center rounded-full bg-white p-[0.3rem] dark:bg-dark-200/80">
      <Avatar
        name="Joseph Mukorivo"
        size="sm"
        src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
      />
    </Box>
    <Text className="ml-2" fontWeight="medium">
      {user}
    </Text>
    <Text color="muted">{action}</Text>
    <Text fontWeight="medium">{prevValue}</Text>
    <Text color="muted">to</Text>
    <Text fontWeight="medium">{newValue}</Text>
    <Text color="muted">·</Text>
    <Text color="muted">{timestamp}</Text>
  </Flex>
);

export default function Page(): JSX.Element {
  const activites: ActivityProps[] = [
    {
      id: 1,
      user: "josemukorivo",
      action: "changed status from",
      prevValue: "Todo",
      newValue: "In Progress",
      timestamp: "23 hours ago",
    },
    {
      id: 2,
      user: "janedoe",
      action: "created task",
      prevValue: "Issue 1",
      newValue: "Todo",
      timestamp: "2 days ago",
    },
    {
      id: 3,
      user: "johnsmith",
      action: "assigned task to",
      prevValue: "jackdoe",
      newValue: "janedoe",
      timestamp: "1 week ago",
    },
    {
      id: 4,
      user: "johndoe",
      action: "changed status from",
      prevValue: "In Progress",
      newValue: "Done",
      timestamp: "1 hour ago",
    },
    {
      id: 5,
      user: "josemukorivo",
      action: "changed status from",
      prevValue: "Todo",
      newValue: "In Progress",
      timestamp: "23 hours ago",
    },
    {
      id: 6,
      user: "janedoe",
      action: "created task",
      prevValue: "Issue 1",
      newValue: "Todo",
      timestamp: "2 days ago",
    },
    {
      id: 7,
      user: "johnsmith",
      action: "assigned task to",
      prevValue: "jackdoe",
      newValue: "janedoe",
      timestamp: "1 week ago",
    },
  ];

  function generateContributions(count: number) {
    const contributions = [];
    const daysInMonth = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31];

    let currentCount = 0;
    const currentDate = new Date("2023-01-01");

    for (let i = 0; i < 12; i++) {
      const monthDays = daysInMonth[i];
      for (let j = 1; j <= monthDays; j++) {
        const year = currentDate.getFullYear();
        const month = String(currentDate.getMonth() + 1).padStart(2, "0");
        const day = String(currentDate.getDate()).padStart(2, "0");
        const dateString = `${year}/${month}/${day}`;
        const randomCount = Math.floor(Math.random() * 10) + 1; // Random count between 1 and 10
        contributions.push({ date: dateString, count: randomCount });
        currentCount++;
        if (currentCount >= count) {
          return contributions;
        }
        currentDate.setDate(currentDate.getDate() + 1);
      }
    }

    return contributions;
  }

  return (
    <>
      <HeaderContainer className="justify-between">
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Dashboard",
              icon: <Columns3 className="h-5 w-auto" />,
            },
          ]}
        />
        <Flex gap={2}>
          <Button
            color="tertiary"
            leftIcon={<SlidersHorizontal className="h-4 w-auto" />}
            size="sm"
            variant="outline"
          >
            Display
          </Button>
          <NewIssueButton />
        </Flex>
      </HeaderContainer>
      <BodyContainer>
        <Container className="py-4">
          <Text as="h2" fontSize="3xl" fontWeight="medium">
            Good afternoon, Joseph.
          </Text>
          <Text color="muted">
            Here&rsquo;s what&rsquo;s happening with your projects today.
          </Text>

          <Box className="my-4 grid grid-cols-5 gap-4">
            {new Array(5).fill(0).map((_, i) => (
              <Box
                className="rounded-lg border border-gray-50 px-4 py-4 shadow dark:border-dark-200 dark:bg-dark-200/30"
                key={i}
              >
                <Flex align="center" justify="between">
                  <Text className="mb-2" color="muted">
                    Issues closed
                  </Text>
                  <ArrowUpRight className="h-5 w-auto text-primary" />
                </Flex>
                <Text className="mb-2" fontSize="2xl" fontWeight="semibold">
                  27
                </Text>
                <Text color="muted">+5% from last month</Text>
              </Box>
            ))}
          </Box>

          <Text fontSize="lg" fontWeight="medium">
            {generateContributions(365).length} contributions in 2023
          </Text>
          <Box className="mt-1 rounded-lg border border-gray-50 p-4 shadow dark:border-dark-200 dark:bg-dark-200/30">
            <HeatMap
              className="w-full"
              legendCellSize={20}
              rectProps={{
                rx: 10,
              }}
              rectSize={15}
              startDate={new Date("2023/01/01")}
              value={generateContributions(365)}
              weekLabels={["", "Mon", "", "Wed", "", "Fri", ""]}
            />
            <Text className="ml-2.5 mt-4" color="muted" fontSize="sm">
              Light color represents less contributions and dark color
              represents more contributions.
            </Text>
          </Box>

          <Box className="my-4 grid grid-cols-2 gap-4">
            <Box className="h-[30rem] rounded-lg border border-gray-50 p-4 shadow dark:border-dark-200 dark:bg-dark-200/30">
              <Flex align="center" className="mb-3" justify="between">
                <Text fontSize="lg">Assigned to me</Text>
                <Menu>
                  <Menu.Button>
                    <Button
                      color="tertiary"
                      rightIcon={<ChevronDown className="h-5 w-auto" />}
                      size="sm"
                      variant="outline"
                    >
                      Due this week
                    </Button>
                  </Menu.Button>
                  <Menu.Items align="end">
                    <Menu.Group>
                      <Menu.Item>Last week</Menu.Item>
                      <Menu.Item>Last month</Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              </Flex>

              <Tabs defaultValue="open">
                <Tabs.List className="mx-0">
                  <Tabs.Tab value="open">Open</Tabs.Tab>
                  <Tabs.Tab value="closed">Closed</Tabs.Tab>
                </Tabs.List>
                <Tabs.Panel value="open">
                  <Box className="mt-4 border-t border-gray-50 dark:border-dark-200">
                    {new Array(7).fill(0).map((_, i) => (
                      <IssueContextMenu key={i}>
                        <RowWrapper
                          className={cn("px-1", {
                            "border-b-0": i === 7 - 1,
                          })}
                        >
                          <Flex
                            align="center"
                            className="relative select-none"
                            gap={2}
                          >
                            <PrioritiesMenu priority="No Priority" />
                            <Flex align="center" gap={2}>
                              <Text
                                className="w-[55px] truncate"
                                color="muted"
                                fontWeight="medium"
                              >
                                COM-12
                              </Text>
                              <StatusesMenu isSearchEnabled status="Backlog" />
                              <Text className="overflow-hidden text-ellipsis whitespace-nowrap pl-2 hover:opacity-90">
                                Design a new homepage
                              </Text>
                            </Flex>
                          </Flex>
                          <Flex align="center" gap={3}>
                            <Text
                              className="flex items-center gap-1"
                              color="muted"
                            >
                              Sep 27
                              <Calendar className="h-4 w-auto" />
                            </Text>
                          </Flex>
                        </RowWrapper>
                      </IssueContextMenu>
                    ))}
                  </Box>
                </Tabs.Panel>
                <Tabs.Panel value="closed">
                  <Text className="pt-1">Closed</Text>
                </Tabs.Panel>
              </Tabs>
            </Box>

            <Box className="h-[30rem] rounded-lg border border-gray-50 p-4 shadow dark:border-dark-200 dark:bg-dark-200/30">
              <Flex align="center" className="mb-5" justify="between">
                <Text fontSize="lg">Recent activities</Text>
                <Menu>
                  <Menu.Button>
                    <Button
                      color="tertiary"
                      rightIcon={<ChevronDown className="h-5 w-auto" />}
                      size="sm"
                      variant="outline"
                    >
                      Due this week
                    </Button>
                  </Menu.Button>
                  <Menu.Items align="end">
                    <Menu.Group>
                      <Menu.Item>Last week</Menu.Item>
                      <Menu.Item>Last month</Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              </Flex>

              <Flex className="relative" direction="column" gap={5}>
                <Box className="pointer-events-none absolute left-4 top-0 z-0 h-full border-l-[1.5px] border-gray-100 dark:border-dark-200" />
                {activites.map((activity) => (
                  <Activity key={activity.id} {...activity} />
                ))}
              </Flex>
            </Box>
          </Box>
        </Container>
      </BodyContainer>
    </>
  );
}
