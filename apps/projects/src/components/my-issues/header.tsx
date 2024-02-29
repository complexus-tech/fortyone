"use client";
import { Box, BreadCrumbs, Button, Divider, Flex, Popover, Text } from "ui";
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
              <span className="sr-onlyy">Display</span>
            </Button>
          </Popover.Trigger>
          <Popover.Content align="end">
            <Box className="max-w-[27rem] px-4 py-2">
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
              </Flex>
            </Box>
            <Divider className="my-2" />
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
