import { Box, Tabs, Text, Flex, Divider, Avatar, Button, Menu } from "ui";
import { MoreVerticalIcon, ProjectsIcon, UserIcon } from "icons";
import { RowWrapper, IssueStatusIcon, PriorityIcon } from "@/components/ui";

export const Sidebar = () => {
  const overview = [
    {
      title: "Email",
      value: "josemukorivo@gmail.com",
    },
    {
      title: "Location",
      value: "Harare, Zimbabwe",
    },
    {
      title: "Joined",
      value: "Jan 16, 2024",
    },
    {
      title: "Teams",
      value: "Web, Mobile, Design",
    },
  ];

  return (
    <Box className="py-6">
      <Box className="mb-6 px-6">
        <Flex align="center" className="mb-3" justify="between">
          <Text className="flex items-center gap-1" fontWeight="medium">
            <Avatar
              className="mr-1"
              name="Joseph Mukorivo"
              size="sm"
              src="https://lh3.googleusercontent.com/ogw/AGvuzYY32iGR6_5Wg1K3NUh7jN2ciCHB12ClyNHIJ1zOZQ=s64-c-mo"
            />
            Joseph Mukorivo
            <Text as="span" color="muted" fontSize="sm" fontWeight="semibold">
              (josemukorivo)
            </Text>
          </Text>
          <Menu>
            <Menu.Button asChild>
              <Button
                color="tertiary"
                leftIcon={<MoreVerticalIcon className="h-5 w-auto" />}
                size="sm"
                variant="naked"
              >
                <span className="sr-only">More options</span>
              </Button>
            </Menu.Button>
            <Menu.Items align="end">
              <Menu.Group>
                <Menu.Item>
                  <UserIcon className="h-4 w-auto" />
                  Edit Profile
                </Menu.Item>
              </Menu.Group>
            </Menu.Items>
          </Menu>
        </Flex>
      </Box>
      <Divider className="mb-6" />
      <Box className="px-6">
        {overview.map(({ title, value }) => (
          <Flex
            align="center"
            className="mb-6"
            gap={2}
            justify="between"
            key={title}
          >
            <Text>{title}:</Text>
            <Text color="muted">{value}</Text>
          </Flex>
        ))}
      </Box>
      <Divider className="my-6" />
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
                <Text color="muted">4</Text>
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
                <Text color="muted">4</Text>
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
                <Text color="muted">4</Text>
              </RowWrapper>
            ))}
          </Tabs.Panel>
          <Tabs.Panel value="projects">
            {new Array(4).fill(1).map((_, idx) => (
              <RowWrapper className="px-1 py-2" key={idx}>
                <Flex align="center" gap={2}>
                  <ProjectsIcon className="h-[1.15rem] w-auto" />
                  <Text color="muted">High</Text>
                </Flex>
                <Text color="muted">4</Text>
              </RowWrapper>
            ))}
          </Tabs.Panel>
        </Tabs>
      </Box>
    </Box>
  );
};
