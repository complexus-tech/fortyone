import { ArrowDownIcon, HealthIcon, ObjectiveIcon } from "icons";
import { Button, Flex, Text, Tabs } from "ui";
import { Summary } from "./summary";
import { Updates } from "./updates";

export const Activity = () => {
  return (
    <Tabs defaultValue="summary">
      <Flex align="center" className="mb-3" justify="between">
        <Tabs.List className="mx-0">
          <Tabs.Tab
            className="gap-1.5"
            leftIcon={<ObjectiveIcon className="h-[1.1rem]" />}
            value="summary"
          >
            Summary
          </Tabs.Tab>
          <Tabs.Tab
            className="gap-1.5"
            leftIcon={<HealthIcon />}
            value="updates"
          >
            Updates
          </Tabs.Tab>
        </Tabs.List>
        <Flex align="center" gap={2}>
          <Text color="muted">Health:</Text>
          <Button
            className="gap-1"
            color="tertiary"
            leftIcon={<HealthIcon />}
            rightIcon={<ArrowDownIcon className="h-3.5" />}
            size="sm"
          >
            On Track
          </Button>
        </Flex>
      </Flex>
      <Tabs.Panel value="summary">
        <Summary />
      </Tabs.Panel>
      <Tabs.Panel value="updates">
        <Updates />
      </Tabs.Panel>
    </Tabs>
  );
};
