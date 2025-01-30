import {
  Container,
  Text,
  Box,
  Divider,
  Flex,
  Wrapper,
  Button,
  ProgressBar,
  Menu,
  CircleProgressBar,
} from "ui";
import { MoreHorizontalIcon, OKRIcon } from "icons";
import { BoardDividedPanel } from "@/components/ui";
import { Sidebar } from "../sidebar";
import { Activity } from "./activity";
import { Properties } from "./properties";

export const Overview = () => {
  return (
    <BoardDividedPanel autoSaveId="teams:objectives:stories:divided-panel">
      <BoardDividedPanel.MainPanel>
        <Container className="h-[calc(100vh-7.7rem)] overflow-y-auto pt-6">
          <Box>
            <Text className="mb-4 text-3xl antialiased" fontWeight="semibold">
              Overview
            </Text>
            <Text className="opacity-80" color="muted" fontSize="lg">
              objective description
            </Text>
          </Box>
          <Properties />
          <Divider className="my-8" />
          <Activity />
          <Box className="my-8">
            <Flex align="center" className="mb-3" justify="between">
              <Text className="text-lg antialiased" fontWeight="semibold">
                Key Results
              </Text>
              <Button color="tertiary" size="sm">
                Add Key Result
              </Button>
            </Flex>
            <CircleProgressBar className="size-24" progress={56} />
            <Wrapper className="flex items-center justify-between gap-2 rounded-[0.65rem]">
              <Flex align="center" gap={3}>
                <OKRIcon strokeWidth={2.8} />
                <Text>Deploy to production by end of Q3</Text>
              </Flex>
              <Flex
                align="center"
                className="divide-x divide-gray-100 dark:divide-dark-100/80"
              >
                <Flex
                  align="center"
                  className="px-6"
                  direction="column"
                  gap={1}
                >
                  <Text color="muted">Current</Text>
                  <Text fontSize="sm">75%</Text>
                </Flex>
                <Flex
                  align="center"
                  className="px-6"
                  direction="column"
                  gap={1}
                >
                  <Text color="muted">Current</Text>
                  <Text fontSize="sm">75%</Text>
                </Flex>
                <Flex
                  align="center"
                  className="px-6"
                  direction="column"
                  gap={1}
                >
                  <Text color="muted">Progress</Text>
                  <Flex align="center" gap={2}>
                    <ProgressBar className="w-16" progress={75} />
                    <Text fontSize="sm">75%</Text>
                  </Flex>
                </Flex>

                <Menu>
                  <Menu.Button>
                    <Button
                      asIcon
                      color="tertiary"
                      leftIcon={<MoreHorizontalIcon />}
                      rounded="full"
                      size="sm"
                    >
                      <span className="sr-only">Edit</span>
                    </Button>
                  </Menu.Button>
                  <Menu.Items>
                    <Menu.Group>
                      <Menu.Item>Edit</Menu.Item>
                      <Menu.Item>Delete</Menu.Item>
                    </Menu.Group>
                  </Menu.Items>
                </Menu>
              </Flex>
            </Wrapper>
          </Box>
        </Container>
      </BoardDividedPanel.MainPanel>
      <BoardDividedPanel.SideBar className="h-[calc(100vh-7.7rem)]" isExpanded>
        <Sidebar className="h-[calc(100vh-7.7rem)] overflow-y-auto" />
      </BoardDividedPanel.SideBar>
    </BoardDividedPanel>
  );
};
