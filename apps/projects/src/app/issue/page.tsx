"use client";
import { TbMoodPlus } from "react-icons/tb";
import {
  Avatar,
  Badge,
  Box,
  BreadCrumbs,
  Button,
  Container,
  Divider,
  Flex,
  Tabs,
  Text,
  Tooltip,
  ResizablePanel,
} from "ui";
import {
  Bell,
  Link2,
  Clipboard,
  Trash2,
  ChevronUp,
  ChevronDown,
  Star,
  TimerReset,
  Calendar,
  Plus,
  Layers3,
  ShieldCheck,
  ShieldEllipsis,
  Network,
  Disc,
  BarChart,
  User,
  Tags,
  Paperclip,
  History,
  MessageSquareText,
  Hourglass,
  GitCompareArrows,
} from "lucide-react";
import { BodyContainer, HeaderContainer } from "@/components/shared";
import type { ActivityProps } from "@/components/ui";
import { IssueStatusIcon, PriorityIcon, Activity } from "@/components/ui";
import { Editor } from "./editor";

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
  ];

  return (
    <>
      <HeaderContainer>
        <Flex align="center" className="w-full" justify="between">
          <BreadCrumbs
            breadCrumbs={[
              { name: "Complexus" },
              { name: "Web design" },
              {
                name: "COM-12",
                icon: <Link2 className="h-5 w-auto" strokeWidth={2} />,
              },
            ]}
          />
          <Flex align="center" gap={2} justify="between">
            <Text className="mr-2">
              2 /{" "}
              <Text as="span" color="muted">
                8
              </Text>
            </Text>
            <Button color="tertiary" size="sm">
              <ChevronUp className="h-4 w-auto" />
            </Button>
            <Button className="mr-2" color="tertiary" disabled size="sm">
              <ChevronDown className="h-4 w-auto" />
            </Button>
            <Button
              className="px-3"
              color="tertiary"
              leftIcon={<Star className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            >
              Favourite
            </Button>
            <Button
              className="px-2"
              leftIcon={<Bell className="h-4 w-auto" />}
              size="sm"
            >
              Subscribe
            </Button>
          </Flex>
        </Flex>
      </HeaderContainer>
      <BodyContainer className="overflow-y-hidden">
        <ResizablePanel direction="horizontal">
          <ResizablePanel.Panel defaultSize={74}>
            <Box className="h-full overflow-y-auto border-r border-gray-50 pb-8 dark:border-dark-200">
              <Container className="pt-6">
                <Text className="mb-6" fontSize="3xl" fontWeight="medium">
                  Change the color of the button to red
                </Text>
                <Editor />

                <Flex className="mb-4 mt-6" gap={1}>
                  <Badge
                    color="tertiary"
                    rounded="lg"
                    size="lg"
                    variant="outline"
                  >
                    👍🏼 1
                  </Badge>
                  <Badge
                    color="tertiary"
                    rounded="lg"
                    size="lg"
                    variant="outline"
                  >
                    🇿🇼 3
                  </Badge>
                  <Badge
                    color="tertiary"
                    rounded="lg"
                    size="lg"
                    variant="outline"
                  >
                    👌 2
                  </Badge>
                  <Badge
                    className="px-2"
                    color="tertiary"
                    rounded="lg"
                    size="lg"
                    variant="outline"
                  >
                    <TbMoodPlus className="h-5 w-auto opacity-80" />
                  </Badge>
                </Flex>
                <Flex justify="end">
                  <Button
                    color="tertiary"
                    leftIcon={<Plus className="h-5 w-auto" />}
                    size="sm"
                    variant="naked"
                  >
                    Add sub issue
                  </Button>
                </Flex>

                <Divider className="my-4" />

                <Box>
                  <Text
                    as="h4"
                    className="flex items-center gap-1"
                    fontWeight="medium"
                  >
                    <Paperclip className="h-5 w-auto" />
                    Attachements
                  </Text>
                  <Box className="mb-4 mt-3 flex h-24 cursor-pointer items-center justify-center rounded-xl border-[1.5px] border-dashed border-gray-200 bg-gray-50/50 dark:border-dark-100 dark:bg-dark-200/30">
                    <Text color="muted">Click or drag files here</Text>
                  </Box>

                  <Box className="grid grid-cols-5 gap-4">
                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />

                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-50/70 dark:bg-dark-200/50" />
                  </Box>
                </Box>

                <Divider className="mb-6 mt-8" />

                <Box>
                  <Text
                    as="h4"
                    className="mb-6 flex items-center gap-1"
                    fontWeight="medium"
                  >
                    <History className="h-5 w-auto" />
                    Activity feed
                  </Text>

                  <Tabs defaultValue="all">
                    <Tabs.List className="mx-0 mb-5">
                      <Tabs.Tab
                        className="text-[0.95rem] font-medium"
                        leftIcon={
                          <History
                            className="h-[1.1rem] w-auto"
                            strokeWidth={2.2}
                          />
                        }
                        value="all"
                      >
                        All Activities
                      </Tabs.Tab>
                      <Tabs.Tab
                        className="text-[0.95rem] font-medium"
                        leftIcon={
                          <Hourglass className="h-4 w-auto" strokeWidth={2.8} />
                        }
                        value="updates"
                      >
                        Updates
                      </Tabs.Tab>
                      <Tabs.Tab
                        className="text-[0.95rem] font-medium"
                        leftIcon={
                          <MessageSquareText
                            className="h-[1.1rem] w-auto"
                            strokeWidth={2.2}
                          />
                        }
                        value="comments"
                      >
                        Comments
                      </Tabs.Tab>
                    </Tabs.List>
                    <Tabs.Panel value="all">
                      <Flex className="relative" direction="column" gap={4}>
                        <Box className="pointer-events-none absolute left-4 top-0 z-0 h-[95%] border-l-[1.5px] border-gray-100 dark:border-dark-200" />
                        {activites.map((activity) => (
                          <Activity key={activity.id} {...activity} />
                        ))}
                        <Flex align="start" className="relative z-[2]">
                          <Box className="pointer-events-none absolute bottom-0 left-4 h-[calc(100%-3rem)] w-1 bg-white dark:bg-dark" />
                          <Box className="z-[1] mt-4 flex aspect-square items-center bg-white p-[0.3rem] dark:bg-dark">
                            <Avatar name="Joseph Mukorivo" size="sm" />
                          </Box>
                          <Flex
                            className="ml-1 mt-2 min-h-[6rem] w-full rounded-lg border border-gray-50 px-4 py-4 text-[0.95rem] shadow-sm transition-shadow duration-200 ease-linear focus-within:shadow-lg dark:border-dark-200/80 dark:bg-dark-200/50 dark:shadow-dark-200/50"
                            direction="column"
                            gap={2}
                            justify="between"
                          >
                            <Box
                              className="text-gray-250 outline-none"
                              // contentEditable
                              spellCheck={false}
                            >
                              <Text color="muted">Leave a comment...</Text>
                            </Box>

                            <Flex gap={1} justify="end">
                              <Button
                                className="px-3"
                                color="tertiary"
                                leftIcon={<Paperclip className="h-4 w-auto" />}
                                size="sm"
                                variant="naked"
                              >
                                <span className="sr-only">Attach files</span>
                              </Button>
                              <Button
                                className="px-3"
                                color="tertiary"
                                size="sm"
                                variant="outline"
                              >
                                Comment
                              </Button>
                            </Flex>
                          </Flex>
                        </Flex>
                      </Flex>
                    </Tabs.Panel>
                    <Tabs.Panel value="updates">
                      <Flex className="relative" direction="column" gap={4}>
                        <Box className="pointer-events-none absolute left-4 top-0 z-0 h-full border-l-[1.5px] border-gray-100 dark:border-dark-200" />
                        {activites.map((activity) => (
                          <Activity key={activity.id} {...activity} />
                        ))}
                      </Flex>
                    </Tabs.Panel>
                    <Tabs.Panel value="comments">
                      <Text>No comments available</Text>
                    </Tabs.Panel>
                  </Tabs>
                </Box>
              </Container>
            </Box>
          </ResizablePanel.Panel>
          <ResizablePanel.Handle />
          <ResizablePanel.Panel defaultSize={26} maxSize={35} minSize={25}>
            <Box className="h-full overflow-y-auto bg-gray-50/20 pb-6 dark:bg-dark-200/40">
              <Box className="flex h-16 items-center border-b border-gray-50 dark:border-dark-200">
                <Container className="flex w-full items-center justify-between">
                  <Text color="muted" fontWeight="medium">
                    COMP-13
                  </Text>
                  <Flex gap={2}>
                    <Tooltip title="Copy issue link">
                      <Button
                        color="tertiary"
                        leftIcon={
                          <Link2 className="h-5 w-auto" strokeWidth={2.5} />
                        }
                        variant="naked"
                      >
                        <span className="sr-only">Copy issue link</span>
                      </Button>
                    </Tooltip>
                    <Tooltip title="Copy issue id">
                      <Button
                        color="tertiary"
                        leftIcon={<Clipboard className="h-5 w-auto" />}
                        variant="naked"
                      >
                        <span className="sr-only">Copy issue id</span>
                      </Button>
                    </Tooltip>
                    <Tooltip title="Delete issue">
                      <Button
                        color="danger"
                        leftIcon={<Trash2 className="h-5 w-auto" />}
                        variant="naked"
                      >
                        <span className="sr-only">Delete issue</span>
                      </Button>
                    </Tooltip>
                  </Flex>
                </Container>
              </Box>
              <Container className="pt-6 text-gray-300/90">
                <Text fontWeight="medium">Properties</Text>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Disc className="h-5 w-auto" />
                    Status
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="In Progress" />
                      <Text>In Progress</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <BarChart className="h-5 w-auto" />
                    Priority
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <PriorityIcon priority="High" />
                      <Text>High</Text>
                    </Box>
                  </Button>
                </Box>

                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <User className="h-5 w-auto" />
                    Assignee
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Done" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>

                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Tags className="h-5 w-auto" />
                    Labels
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Calendar className="h-5 w-auto" />
                    Start date
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Calendar className="h-5 w-auto" />
                    Due date
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <TimerReset className="h-5 w-auto" />
                    Sprint
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Layers3 className="h-5 w-auto" />
                    Modules
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <Network className="h-5 w-auto" />
                    Parent
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <ShieldCheck className="h-5 w-auto" />
                    Blocking
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <ShieldEllipsis className="h-5 w-auto" />
                    Blocked by
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>
                <Box className="my-3 grid grid-cols-[9rem_auto] items-center gap-3">
                  <Text
                    className="flex items-center gap-1 truncate"
                    color="muted"
                    fontWeight="medium"
                  >
                    <GitCompareArrows className="h-5 w-auto" />
                    Related to
                  </Text>
                  <Button className="px-4" color="tertiary" variant="naked">
                    <Box className="grid grid-cols-[1.5rem_auto] items-center gap-1">
                      <IssueStatusIcon status="Backlog" />
                      <Text>Backlog</Text>
                    </Box>
                  </Button>
                </Box>

                <Divider className="my-4" />
                <Flex align="center">
                  <Text fontWeight="medium">Links</Text>
                  <Button
                    className="ml-auto"
                    color="tertiary"
                    leftIcon={<Plus className="h-5 w-auto" strokeWidth={2} />}
                    size="sm"
                    variant="outline"
                  >
                    Add
                  </Button>
                </Flex>
              </Container>
            </Box>
          </ResizablePanel.Panel>
        </ResizablePanel>
      </BodyContainer>
    </>
  );
}
