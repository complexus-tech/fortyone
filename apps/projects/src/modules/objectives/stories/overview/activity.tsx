import { ArrowDownIcon, ClockIcon, ObjectiveIcon } from "icons";
import { Button, Flex, Text, Tabs } from "ui";
import { useParams } from "next/navigation";
import { ObjectiveHealthIcon } from "@/components/ui";
import { HealthMenu } from "@/components/ui/health-menu";
import { useObjective, useUpdateObjectiveMutation } from "../../hooks";
import type { ObjectiveHealth } from "../../types";
import { Summary } from "./summary";
import { Updates } from "./updates";

export const Activity = () => {
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const { data: objective } = useObjective(objectiveId);
  const updateMutation = useUpdateObjectiveMutation();

  const handleUpdate = (health: ObjectiveHealth) => {
    updateMutation.mutate({ objectiveId, data: { health } });
  };

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
            leftIcon={<ClockIcon />}
            value="updates"
          >
            Updates
          </Tabs.Tab>
        </Tabs.List>
        <Flex align="center" gap={2}>
          <Text color="muted">Health:</Text>

          <HealthMenu>
            <HealthMenu.Trigger>
              <Button
                className="gap-1"
                color="tertiary"
                leftIcon={<ObjectiveHealthIcon health={objective?.health} />}
                rightIcon={<ArrowDownIcon className="h-3.5" />}
                size="sm"
              >
                {objective?.health || "No health"}
              </Button>
            </HealthMenu.Trigger>
            <HealthMenu.Items
              health={objective?.health}
              setHealth={(health) => {
                handleUpdate(health);
              }}
            />
          </HealthMenu>
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
