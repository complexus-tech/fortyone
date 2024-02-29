"use client";
import {
  Box,
  BreadCrumbs,
  Button,
  Divider,
  Flex,
  Popover,
  Switch,
  Text,
} from "ui";
import { ArrowDownIcon, IssueIcon, PreferencesIcon } from "icons";
import { HeaderContainer } from "@/components/layout";
import { SideDetailsSwitch } from "@/components/ui";

export const Header = ({
  isExpanded,
  setIsExpanded,
}: {
  isExpanded: boolean | null;
  setIsExpanded: (isExpanded: boolean) => void;
}) => {
  return (
    <HeaderContainer className="justify-between">
      <BreadCrumbs
        breadCrumbs={[
          {
            name: "My issues",
            icon: <IssueIcon className="h-5 w-auto" strokeWidth={2} />,
          },
          { name: "Assigned" },
        ]}
      />
      <Flex align="center" gap={2}>
        <Popover>
          <Popover.Trigger>
            <Button
              color="tertiary"
              leftIcon={<PreferencesIcon className="h-4 w-auto" />}
              rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
              size="sm"
              variant="outline"
            >
              Display
            </Button>
          </Popover.Trigger>
          <Popover.Content align="end" className="max-w-[24rem]">
            <Flex
              align="center"
              className="my-2 px-4"
              gap={2}
              justify="between"
            >
              <Text color="muted">Group by</Text>
              <Button
                color="tertiary"
                rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
                size="sm"
              >
                Status
              </Button>
            </Flex>
            <Flex
              align="center"
              className="mb-3 px-4"
              gap={2}
              justify="between"
            >
              <Text color="muted">Order by</Text>
              <Button
                color="tertiary"
                rightIcon={<ArrowDownIcon className="h-4 w-auto" />}
                size="sm"
              >
                Priority
              </Button>
            </Flex>
            <Divider className="my-2" />
            <Box className="max-w-[27rem] px-4 py-2">
              <Text className="mb-4" fontWeight="medium">
                Display options
              </Text>
              <Text className="mb-4" color="muted">
                <label
                  className="flex select-none items-center justify-between gap-2"
                  htmlFor="more"
                >
                  Show empty columns <Switch id="more" />
                </label>
              </Text>
              <Text className="mb-2" color="muted">
                Display columns
              </Text>

              <Flex gap={2} wrap>
                <Button rounded="sm" size="xs">
                  Status
                </Button>
                <Button rounded="sm" size="xs">
                  Assignee
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Priority
                </Button>
                <Button rounded="sm" size="xs">
                  Due date
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Created
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Updated
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Sprint
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Epic
                </Button>
                <Button color="tertiary" rounded="sm" size="xs">
                  Labels
                </Button>
              </Flex>
            </Box>
            <Divider className="mb-3 mt-2" />
            <Text className="mb-2 px-4" color="muted">
              <label
                className="flex select-none items-center justify-between gap-2"
                htmlFor="more"
              >
                Show empty groups <Switch id="more" />
              </label>
            </Text>
            <Text className="mb-3 px-4" color="muted">
              <label
                className="flex select-none items-center justify-between gap-2"
                htmlFor="more"
              >
                Show sub issues <Switch id="more" />
              </label>
            </Text>
            <Divider className="mb-3" />
            <Text className="mb-2 px-4" color="primary" fontWeight="medium">
              Reset to default
            </Text>
          </Popover.Content>
        </Popover>

        <span className="text-gray-200 dark:text-dark-100">|</span>
        <SideDetailsSwitch
          isExpanded={isExpanded}
          setIsExpanded={setIsExpanded}
        />
      </Flex>
    </HeaderContainer>
  );
};
