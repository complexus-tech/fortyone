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
} from "lucide-react";
import { Panel, PanelGroup, PanelResizeHandle } from "react-resizable-panels";
import { BodyContainer, HeaderContainer } from "../../../components/shared";
import { IssueStatusIcon, PriorityIcon } from "../../../components/ui";

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
    <Box className="flex aspect-square items-center bg-white p-[0.3rem] dark:bg-dark">
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
        <PanelGroup direction="horizontal">
          <Panel defaultSize={74}>
            <Box className="h-full overflow-y-auto border-r border-gray-100 pb-8 dark:border-dark-200">
              <Container className="pt-6">
                <Text
                  as="h3"
                  className="flex items-center gap-2"
                  color="muted"
                  fontSize="lg"
                  fontWeight="medium"
                >
                  <Link2 className="h-5 w-auto" strokeWidth={2.5} />
                  COMP-13
                </Text>
                <Text className="mb-6 mt-3" fontSize="3xl">
                  Change the color of the button to red
                </Text>
                <Text className="leading-7" color="muted" fontSize="lg">
                  Change the color of the button to red. This will hold the
                  description of the issue. It will be a long one so we can test
                  the overflow of the container.
                  <br />
                  <br />
                  Use the color palette from the design system to change the
                  color of the button. The pallette is available in the design
                  system documentation. Use the color palette from the design
                  system to change the color of the button. The pallette is
                  available in the design system documentation.
                </Text>

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
                  <Box className="mb-4 mt-3 flex h-24 cursor-pointer items-center justify-center rounded-xl border-[1.5px] border-dashed bg-gray-100 dark:border-dark-100 dark:bg-dark-200/30">
                    <Text color="muted">Click or drag files here</Text>
                  </Box>

                  <Box className="grid grid-cols-5 gap-4">
                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />

                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />
                    <Box className="h-28 rounded-xl bg-gray-100 dark:bg-dark-200/50" />
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
                        className="px-2 text-[0.95rem]"
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
                        className="px-2 text-[0.95rem]"
                        leftIcon={
                          <Hourglass className="h-4 w-auto" strokeWidth={2.8} />
                        }
                        value="updates"
                      >
                        Updates
                      </Tabs.Tab>
                      <Tabs.Tab
                        className="px-2 text-[0.95rem]"
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
                      <Flex className="relative" direction="column" gap={6}>
                        <Box className="pointer-events-none absolute left-4 top-0 z-0 h-[95%] border-l-[1.5px] border-gray-100 dark:border-dark-200" />
                        {activites.map((activity) => (
                          <Activity key={activity.id} {...activity} />
                        ))}
                      </Flex>
                      <Flex align="start">
                        <Box className="relative mt-6 flex aspect-square items-center bg-white p-[0.3rem] dark:bg-dark">
                          <Box className="pointer-events-none absolute bottom-10 left-4 z-0 h-[95%] border-l-[1.5px] border-gray-100 dark:border-dark-200" />
                          <Avatar
                            name="Joseph Mukorivo"
                            size="sm"
                            src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
                          />
                        </Box>
                        <Flex
                          className="ml-1 mt-4 min-h-[6rem] w-full rounded-lg border border-gray-100 px-4 py-4 text-[0.95rem] shadow dark:border-dark-200/80 dark:bg-dark-200/50 dark:shadow-dark-200/50"
                          direction="column"
                          gap={2}
                          justify="between"
                        >
                          <Box
                            className="outline-none"
                            contentEditable
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
                            >
                              <span className="sr-only">Attach files</span>
                            </Button>
                            <Button className="px-3" color="tertiary" size="sm">
                              Comment
                            </Button>
                          </Flex>
                        </Flex>
                      </Flex>
                    </Tabs.Panel>
                    <Tabs.Panel value="updates">
                      <Flex
                        className="relative ml-1"
                        direction="column"
                        gap={6}
                      >
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
          </Panel>
          <PanelResizeHandle />
          <Panel defaultSize={26} maxSize={35} minSize={25}>
            <Box className="h-full overflow-y-auto bg-gray-50/20 pb-6 dark:bg-dark-200/40">
              <Box className="flex h-16 items-center border-b border-gray-100 dark:border-dark-200">
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
              <Container className="pt-6">
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
                  <Text className="truncate" color="muted" fontWeight="medium">
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
                  >
                    Add
                  </Button>
                </Flex>
              </Container>
            </Box>
          </Panel>
        </PanelGroup>
      </BodyContainer>
    </>
  );
}
