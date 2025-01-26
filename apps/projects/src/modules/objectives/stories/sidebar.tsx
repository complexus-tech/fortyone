import { Box, Tabs, Text, Flex, ProgressBar, Divider, Avatar } from "ui";
import { useParams } from "next/navigation";
import { RowWrapper, StoryStatusIcon, PriorityIcon } from "@/components/ui";
import { useObjectiveStories } from "@/modules/stories/hooks/objective-stories";
import { useObjectives } from "@/modules/objectives/hooks/use-objectives";
import { useMembers } from "@/lib/hooks/members";
import { useStatuses } from "@/lib/hooks/statuses";
import { useLabels } from "@/lib/hooks/labels";
import type { StoryPriority } from "@/modules/stories/types";

export const Sidebar = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: stories = [] } = useObjectiveStories(objectiveId);
  const { data: objectives = [] } = useObjectives();
  const { data: members = [] } = useMembers();
  const { data: statuses = [] } = useStatuses();
  const { data: labels = [] } = useLabels();

  const objective = objectives.find((obj) => obj.id === objectiveId)!;
  const totalStories = stories.length;

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
  const storiesWithNoAssignee = stories.filter((s) => !s.assigneeId).length;

  const allLabels = stories.reduce<Record<string, number>>((acc, story) => {
    story.labels.forEach((labelId) => {
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

  return (
    <Box className="py-6">
      <Box className="mb-4 px-6">
        <Text className="mb-3" fontSize="lg">
          Overview
        </Text>
        <Flex direction="column" gap={3}>
          <Flex align="center" justify="between">
            <Text color="muted">Total Stories</Text>
            <Text color="muted">{totalStories}</Text>
          </Flex>
          <Flex align="center" justify="between">
            <Text color="muted">Progress</Text>
            <Flex align="center" gap={2}>
              <ProgressBar
                className="w-20"
                progress={
                  (objective.stats.completed / objective.stats.total) * 100 || 0
                }
              />
              <Text color="muted">
                {Math.round(
                  (objective.stats.completed / objective.stats.total) * 100 ||
                    0,
                )}
                %
              </Text>
            </Flex>
          </Flex>
        </Flex>
      </Box>

      <Divider className="mb-4" />

      <Box className="px-6">
        <Text className="mb-4" fontSize="lg">
          Stories Overview
        </Text>
        <Tabs defaultValue="status">
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
            {storiesWithNoAssignee > 0 && (
              <RowWrapper className="px-1 py-2 md:px-0">
                <Flex align="center" gap={2}>
                  <Avatar size="xs" />
                  <Text color="muted">No Assignee</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar
                    className="w-20"
                    progress={(storiesWithNoAssignee / totalStories) * 100}
                  />
                  <Text color="muted">
                    {storiesWithNoAssignee} of {totalStories}
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
