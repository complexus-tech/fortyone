import { Box, Tabs, Text, Flex, ProgressBar, Divider, Badge } from "ui";
import { IssueIcon } from "icons";
import { RowWrapper, IssueStatusIcon, PriorityIcon } from "@/components/ui";

export const Sidebar = () => {
  return (
    <Box className="py-8">
      <Flex align="center" className="mb-6 px-6" justify="between">
        <Text className="flex items-center gap-2" fontSize="lg">
          <IssueIcon className="h-5 w-auto" strokeWidth={2} />
          My issues
        </Text>
        <Badge color="tertiary" size="lg">
          Assigned
        </Badge>
      </Flex>

      <Divider className="mb-6" />

      <Box className="px-6">
        <Text className="mb-4" fontSize="lg">
          Overview
        </Text>
        <Tabs defaultValue="status">
          <Tabs.List className="mx-0 mb-3">
            <Tabs.Tab value="status">Status</Tabs.Tab>
            <Tabs.Tab value="labels">Labels</Tabs.Tab>
            <Tabs.Tab value="priority">Priority</Tabs.Tab>
            <Tabs.Tab value="projects">Projects</Tabs.Tab>
          </Tabs.List>
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
          <Tabs.Panel value="projects">
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

        <Text className="mb-4 mt-6" fontSize="lg">
          Upcoming due dates
        </Text>

        <Flex
          align="center"
          className="min-h-80 rounded-xl bg-gray-50/80 px-4 dark:bg-dark-300/80"
          direction="column"
          justify="center"
        >
          <IssueIcon className="mt-6 h-20 w-auto rotate-12" />
          <Text className="mt-4" fontWeight="medium">
            No issues due soon.
          </Text>
          <Text className="mt-2" color="muted">
            Issues with due dates will appear here.
          </Text>
        </Flex>
      </Box>
    </Box>
  );
};
