"use client";

import type { Dispatch, SetStateAction } from "react";
import { Button, Dialog, Divider, Flex, Text, Wrapper } from "ui";
import { CrownIcon } from "icons";
import { useUserRole } from "@/hooks/role";
import { useTerminology } from "@/hooks/use-terminology-display";
import { useWorkspacePath } from "@/hooks/use-workspace-path";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";

type ObjectivePlanLimitDialogProps = {
  isOpen: boolean;
  objectiveCount: number;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
};

/**
 * The entitlement fallback for the objective creation flow.
 *
 * Keeping it separate from the form lets the form own only objective state and
 * makes the billing messaging independently reusable and testable.
 */
export const ObjectivePlanLimitDialog = ({
  isOpen,
  objectiveCount,
  setIsOpen,
}: ObjectivePlanLimitDialogProps) => {
  const { withWorkspace } = useWorkspacePath();
  const { userRole } = useUserRole();
  const { getTermDisplay } = useTerminology();
  const { tier, getLimit } = useSubscriptionFeatures();
  const objectiveLimit = getLimit("maxObjectives");
  const objectiveTerm = getTermDisplay("objectiveTerm", {
    variant: objectiveLimit !== 1 ? "plural" : "singular",
  });

  return (
    <Dialog open={isOpen}>
      <Dialog.Content hideClose>
        <Dialog.Header className="flex items-center gap-2 px-6 pt-6 text-xl">
          <CrownIcon className="text-warning relative -top-px h-6" />
          <Dialog.Title>
            {objectiveLimit === 0 ? (
              <>
                Your plan does not support creating{" "}
                {getTermDisplay("objectiveTerm", { variant: "plural" })}
              </>
            ) : (
              <>
                {getTermDisplay("objectiveTerm", {
                  variant: "plural",
                  capitalize: true,
                })}{" "}
                Limit Reached
              </>
            )}
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body>
          <Text className="mb-4 dark:font-normal" color="muted" fontSize="lg">
            You&apos;ve reached the limit of {objectiveLimit} {objectiveTerm} on
            your {tier.replace("free", "hobby")} plan.{" "}
            {userRole === "admin" ? "Upgrade " : "Ask your admin to upgrade "}
            to create unlimited{" "}
            {getTermDisplay("objectiveTerm", { variant: "plural" })} and unlock
            premium features.
          </Text>
          <Wrapper className="bg-surface/60">
            <Flex align="center" gap={3} justify="between">
              <Text color="muted" fontSize="lg">
                Current plan:
              </Text>
              <Text fontSize="lg" transform="capitalize">
                {tier.replace("free", "hobby")}
              </Text>
            </Flex>
            <Divider className="my-3" />
            <Flex align="center" gap={3} justify="between">
              <Text color="muted" fontSize="lg">
                {getTermDisplay("objectiveTerm", {
                  variant: "plural",
                  capitalize: true,
                })}
                :
              </Text>
              <Text color="primary" fontSize="lg">
                {objectiveCount}/{objectiveLimit}
              </Text>
            </Flex>
          </Wrapper>
          {userRole === "admin" && (
            <Button
              align="center"
              className="mt-4 border-0"
              fullWidth
              href={withWorkspace("/settings/workspace/billing")}
              rounded="lg"
              size="lg"
            >
              Upgrade now
            </Button>
          )}
          <Button
            align="center"
            className="mt-3 mb-2 border-[0.5px]"
            color="tertiary"
            fullWidth
            onClick={() => {
              setIsOpen(false);
            }}
            rounded="lg"
            size="lg"
          >
            Maybe later
          </Button>
        </Dialog.Body>
      </Dialog.Content>
    </Dialog>
  );
};
