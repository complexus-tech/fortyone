import { ArrowDownIcon, ClockIcon, ObjectiveIcon } from "icons";
import { Button, Flex, Text, Tabs, Dialog } from "ui";
import { useParams } from "next/navigation";
import { useSession } from "next-auth/react";
import { useState } from "react";
import { ObjectiveHealthIcon } from "@/components/ui";
import { HealthMenu } from "@/components/ui/health-menu";
import { useIsAdminOrOwner } from "@/hooks/owner";
import { KeyResults } from "@/modules/objectives/stories/overview/key-results";
import { useFeatures, useTerminology } from "@/hooks";
import { useObjective, useUpdateObjectiveMutation } from "../../hooks";
import type { ObjectiveHealth } from "../../types";
import { Summary } from "./summary";
import { Updates } from "./updates";

export const Activity = () => {
  const { data: session } = useSession();
  const features = useFeatures();
  const { getTermDisplay } = useTerminology();
  const { objectiveId } = useParams<{ objectiveId: string }>();
  const [comment, setComment] = useState("");
  const [isCommentOpen, setIsCommentOpen] = useState(false);
  const [health, setHealth] = useState<ObjectiveHealth | null>(null);
  const { data: objective } = useObjective(objectiveId);
  const updateMutation = useUpdateObjectiveMutation();
  const { isAdminOrOwner } = useIsAdminOrOwner(objective?.createdBy);
  const canUpdate = isAdminOrOwner || session?.user?.id === objective?.leadUser;

  const handleUpdate = () => {
    if (health) {
      updateMutation.mutate({ objectiveId, data: { health, comment } });
      setIsCommentOpen(false);
      setComment("");
      setHealth(null);
    }
  };

  return (
    <>
      <Tabs defaultValue="summary">
        <Flex align="center" className="mb-3" justify="between">
          <Tabs.List className="mx-0 md:mx-0">
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
            <Text className="hidden md:block" color="muted">
              Health:
            </Text>
            <HealthMenu>
              <HealthMenu.Trigger>
                <Button
                  className="gap-1"
                  color="tertiary"
                  disabled={!canUpdate}
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
                  setHealth(health);
                  setIsCommentOpen(true);
                }}
              />
            </HealthMenu>
          </Flex>
        </Flex>
        <Tabs.Panel value="summary">
          <Summary />
          {features.keyResultEnabled ? <KeyResults /> : null}
        </Tabs.Panel>
        <Tabs.Panel value="updates">
          <Updates objectiveId={objectiveId} />
        </Tabs.Panel>
      </Tabs>
      <Dialog onOpenChange={setIsCommentOpen} open={isCommentOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title className="flex items-center gap-2 px-6 pt-0.5 text-lg">
              Change {getTermDisplay("objectiveTerm")} heath to{" "}
              <ObjectiveHealthIcon health={health} />
              {health}
            </Dialog.Title>
          </Dialog.Header>
          <Dialog.Description>
            Adding members to your workspace will add more seats to your plan
            once the member accepts the invitation.
          </Dialog.Description>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
