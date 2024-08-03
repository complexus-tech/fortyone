import { Box, Tabs, Text, Flex, ProgressBar, Divider, Badge, Avatar } from "ui";
import { StoryIcon } from "icons";
import { RowWrapper, StoryStatusIcon, PriorityIcon } from "@/components/ui";
import { useTeamStories } from "@/modules/stories/hooks/team-stories";
import { useParams } from "next/navigation";

export const Sidebar = () => {
  const { teamId } = useParams<{ teamId: string }>();
  const { data: stories = [] } = useTeamStories(teamId);
  const totalStories = stories.length;
  return (
    <Box className="py-8">
      <Flex align="center" className="mb-6 px-6" justify="between">
        <Text className="flex items-center gap-2" fontSize="lg">
          <StoryIcon className="h-5 w-auto" strokeWidth={2} />
          All stories
        </Text>
        <Badge className="uppercase" color="tertiary">
          Web design
        </Badge>
      </Flex>

      <Divider className="mb-6" />

      <Box className="px-6">
        <Text className="mb-4" fontSize="lg">
          Overview
        </Text>
        <Tabs defaultValue="status">
          <Tabs.List className="mx-0 mb-3">
            <Tabs.Tab value="assignees">Assignees</Tabs.Tab>
            <Tabs.Tab value="status">Status</Tabs.Tab>
            <Tabs.Tab value="labels">Labels</Tabs.Tab>
            <Tabs.Tab value="priority">Priority</Tabs.Tab>
          </Tabs.List>
          <Tabs.Panel value="assignees">
            {new Array(4).fill(1).map((_, idx) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={idx}>
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
          <Tabs.Panel value="status">
            {new Array(4).fill(1).map((_, idx) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={idx}>
                <Flex align="center" gap={2}>
                  <StoryStatusIcon />
                  <Text color="muted">Backlog</Text>
                </Flex>
                <Flex align="center" gap={2}>
                  <ProgressBar className="w-20" progress={25} />
                  <Text color="muted">25% of {totalStories}</Text>
                </Flex>
              </RowWrapper>
            ))}
          </Tabs.Panel>
          <Tabs.Panel value="labels">
            {new Array(4).fill(1).map((_, idx) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={idx}>
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

          <Tabs.Panel value="priority">
            {new Array(4).fill(1).map((_, idx) => (
              <RowWrapper className="px-1 py-2 md:px-0" key={idx}>
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
  );
};
