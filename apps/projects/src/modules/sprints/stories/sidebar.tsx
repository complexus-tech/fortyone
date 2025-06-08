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
} from "icons";
import { useParams } from "next/navigation";
import { format } from "date-fns";
import { RowWrapper, StoryStatusIcon, PriorityIcon } from "@/components/ui";
import { useSprintStories } from "@/modules/stories/hooks/sprint-stories";
import { useMembers } from "@/lib/hooks/members";
import { useStatuses } from "@/lib/hooks/statuses";
import { useLabels } from "@/lib/hooks/labels";
import type { StoryPriority } from "@/modules/stories/types";
import { useSprint } from "../hooks/sprint-details";

export const Sidebar = () => {
  const { sprintId } = useParams<{ sprintId: string }>();
  const { data: stories = [] } = useSprintStories(sprintId);
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useStatuses();
  const { data: labels = [] } = useLabels();
  const { data: sprint } = useSprint(sprintId);

  if (!sprint) {
    return null;
  }

  const totalStories = stories.length;

  // Compute sprint status
  const getSprintStatus = () => {
    const now = new Date();
    const startDate = new Date(sprint.startDate);
    const endDate = new Date(sprint.endDate);

    if (now < startDate) return "Not Started";
    if (now > endDate) return "Completed";
    return "In Progress";
  };

  // Calculate story statistics
  const storiesByStatus = statuses.map((status) => {
    const count = stories.filter((s) => s.statusId === status.id).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { status, count, percentage };
  });

  const priorities: StoryPriority[] = [
    "Urgent",
    "High",
    "Medium",
    "Low",
    "No Priority",
  ];
  const storiesByPriority = priorities.map((priority) => {
    const count = stories.filter((s) => s.priority === priority).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { priority, count, percentage };
  });

  const storiesByAssignee = members.map((member) => {
    const count = stories.filter((s) => s.assigneeId === member.id).length;
    const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
    return { member, count, percentage };
  });

  const unassignedCount = stories.filter((s) => !s.assigneeId).length;
  const unassignedPercentage =
    totalStories > 0 ? (unassignedCount / totalStories) * 100 : 0;

  // Calculate label statistics
  const allLabels = stories.reduce<Record<string, number>>((acc, story) => {
    story.labels?.forEach((labelId) => {
      if (!acc[labelId]) acc[labelId] = 0;
      acc[labelId]++;
    });
    return acc;
  }, {});

  const labelStats = Object.entries(allLabels)
    .map(([labelId, count]) => {
      const label = labels.find((l) => l.id === labelId);
      const percentage = totalStories > 0 ? (count / totalStories) * 100 : 0;
      return { label, count, percentage };
    })
    .filter((stat) => stat.label);

  const completedStories = stories.filter((s) => {
    const status = statuses.find((st) => st.id === s.statusId);
    return status?.category === "completed";
  }).length;

  const sprintProgress =
    totalStories > 0 ? (completedStories / totalStories) * 100 : 0;

  return (
    <Box className="h-full bg-gray-50/30 py-6 dark:bg-dark-300/50">
      <Box className="px-6">
        <Flex align="center" justify="between">
          <Text className="flex items-center gap-1.5" fontSize="lg">
            <SprintsIcon className="relative -top-px h-[1.4rem] w-auto" />
            {sprint.name}
          </Text>
          <Menu>
            <Menu.Button>
              <Button
                asIcon
                color="tertiary"
                leftIcon={<MoreVerticalIcon />}
                rounded="full"
                size="sm"
                variant="naked"
              >
                <span className="sr-only">More options</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end" className="w-56">
              <Menu.Group className="px-4">
                <Text className="mb-2 mt-1" color="muted">
                  Manage Sprint
                </Text>
              </Menu.Group>
              <Menu.Separator className="mb-1.5" />
              <Menu.Group>
                <Menu.Item>
                  <EditIcon className="h-[1.1rem] w-auto" />
                  Edit sprint
                </Menu.Item>
                <Menu.Item>
                  <LinkIcon />
                  Copy link
                </Menu.Item>
                <Menu.Item>
                  <StarIcon />
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
          <Badge
            className="h-8 px-2 text-base capitalize tracking-wide"
            color={getSprintStatus() === "In Progress" ? "primary" : "tertiary"}
          >
            {getSprintStatus()}
          </Badge>
          <Badge
            className="h-8 px-2 text-base capitalize tracking-wide"
            color="tertiary"
          >
            {format(new Date(sprint.startDate), "d MMM")} -{" "}
            {format(new Date(sprint.endDate), "d MMM")}
          </Badge>
        </Flex>
        <Flex align="center" className="mb-2 mt-3" gap={2} justify="between">
          <Text>Sprint Progress</Text>
          <Text>{Math.round(sprintProgress)}%</Text>
        </Flex>
        <ProgressBar className="h-2" progress={sprintProgress} />
      </Box>
      <Divider className="mb-6 mt-6" />
      <Box className="px-6">
        <Text className="mb-3">Stories Overview</Text>
        <Tabs defaultValue="assignees">
          <Tabs.List className="mx-0 mb-3">
            <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
            <Tabs.Tab value="status">Status</Tabs.Tab>
            <Tabs.Tab value="labels">Labels</Tabs.Tab>
            <Tabs.Tab value="priority">Priority</Tabs.Tab>
          </Tabs.List>

          <Tabs.Panel value="assignees">
            {storiesByAssignee
              .filter((stat) => stat.count > 0)
              .map(({ member, count, percentage }) => (
                <RowWrapper className="px-1 py-2 md:px-0" key={member.id}>
                  <Flex align="center" gap={2}>
                    <Avatar
                      name={member.fullName}
                      size="xs"
                      src={member.avatarUrl}
                    />
                    <Text color="muted">{member.username}</Text>
                  </Flex>
                  <Flex align="center" gap={2}>
                    <ProgressBar className="w-20" progress={percentage} />
                    <Text color="muted">
                      {count} of {totalStories}
                    </Text>
                  </Flex>
                </RowWrapper>
              ))}
            {unassignedCount > 0 && (
              <RowWrapper className="px-1 py-2 md:px-0">
                <Flex align="center" gap={2}>
                  <Avatar size="xs" />
                  <Text color="muted">Unassigned</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar
                    className="w-20"
                    progress={unassignedPercentage}
                  />
                  <Text color="muted">
                    {unassignedCount} of {totalStories}
                  </Text>
                </Flex>
              </RowWrapper>
            )}
          </Tabs.Panel>

          <Tabs.Panel value="status">
            {storiesByStatus
              .filter((stat) => stat.count > 0)
              .map(({ status, count, percentage }) => (
                <RowWrapper className="px-1 py-2 md:px-0" key={status.id}>
                  <Flex align="center" gap={2}>
                    <StoryStatusIcon statusId={status.id} />
                    <Text color="muted">{status.name}</Text>
                  </Flex>
                  <Flex align="center" gap={2}>
                    <ProgressBar className="w-20" progress={percentage} />
                    <Text color="muted">
                      {count} of {totalStories}
                    </Text>
                  </Flex>
                </RowWrapper>
              ))}
          </Tabs.Panel>

          <Tabs.Panel value="labels">
            {labelStats.map(({ label, count, percentage }) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={label!.id}>
                <Flex align="center" gap={2}>
                  <span
                    className="block size-2 rounded-full"
                    style={{ backgroundColor: label!.color }}
                  />
                  <Text color="muted">{label!.name}</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar className="w-20" progress={percentage} />
                  <Text color="muted">
                    {count} of {totalStories}
                  </Text>
                </Flex>
              </RowWrapper>
            ))}
          </Tabs.Panel>

          <Tabs.Panel value="priority">
            {storiesByPriority
              .filter((stat) => stat.count > 0)
              .map(({ priority, count, percentage }) => (
                <RowWrapper className="px-1 py-2 md:px-0" key={priority}>
                  <Flex align="center" gap={2}>
                    <PriorityIcon priority={priority} />
                    <Text color="muted">{priority}</Text>
                  </Flex>
                  <Flex align="center" gap={2}>
                    <ProgressBar className="w-20" progress={percentage} />
                    <Text color="muted">
                      {count} of {totalStories}
                    </Text>
                  </Flex>
                </RowWrapper>
              ))}
          </Tabs.Panel>
        </Tabs>
      </Box>
    </Box>
  );
};
