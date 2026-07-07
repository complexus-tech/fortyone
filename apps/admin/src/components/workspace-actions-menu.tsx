"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import {
  ArrowDown2Icon,
  CalendarPlusIcon,
  DeleteIcon,
  NewTabIcon,
  ReloadIcon,
  UndoIcon,
} from "icons";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Input, Menu, Text, TextArea } from "ui";
import {
  requestWorkspaceSubscriptionSync,
  updateWorkspaceDeleted,
  updateWorkspaceTrial,
} from "@/lib/admin-api";
import { formatDateTime, formatTrialState } from "@/lib/format";
import type { WorkspaceSummary } from "@/lib/types";

type DialogMode = "delete" | "restore" | "sync" | "trial";

const toDateTimeLocal = (date: Date) => {
  const pad = (value: number) => String(value).padStart(2, "0");

  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(
    date.getDate(),
  )}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
};

const suggestedTrialEnd = (trialEndsOn: string | null) => {
  const now = new Date();
  const current = trialEndsOn ? new Date(trialEndsOn) : null;
  const base = current && current.getTime() > now.getTime() ? current : now;
  const suggested = new Date(base);
  suggested.setDate(suggested.getDate() + 7);
  suggested.setHours(17, 0, 0, 0);
  return suggested;
};

const stripeCustomerUrl = (customerId: string) =>
  `https://dashboard.stripe.com/customers/${customerId}`;

const stripeSubscriptionUrl = (subscriptionId: string) =>
  `https://dashboard.stripe.com/subscriptions/${subscriptionId}`;

export const WorkspaceActionsMenu = ({
  workspace,
}: {
  workspace: WorkspaceSummary;
}) => {
  const router = useRouter();
  const [mode, setMode] = useState<DialogMode | null>(null);
  const [isPending, startTransition] = useTransition();
  const defaultTrialEnd = toDateTimeLocal(
    suggestedTrialEnd(workspace.trialEndsOn),
  );
  const minTrialEnd = toDateTimeLocal(new Date());
  const isDeleted = Boolean(workspace.deletedAt);

  const closeDialog = () => {
    setMode(null);
  };

  const handleSubmit = (formData: FormData) => {
    if (!mode) {
      return;
    }

    const reason = String(formData.get("reason") ?? "").trim();
    if (!reason) {
      toast.error("Reason is required.");
      return;
    }

    const trialEndsOn = String(formData.get("trialEndsOn") ?? "");
    if (mode === "trial" && !trialEndsOn) {
      toast.error("Trial end date is required.");
      return;
    }

    startTransition(async () => {
      try {
        if (mode === "trial") {
          await updateWorkspaceTrial(workspace.id, {
            trialEndsOn: new Date(trialEndsOn).toISOString(),
            reason,
          });
          toast.success("Trial updated");
        } else if (mode === "sync") {
          await requestWorkspaceSubscriptionSync(workspace.id, { reason });
          toast.success("Subscription synced");
        } else {
          await updateWorkspaceDeleted(workspace.id, {
            deleted: mode === "delete",
            reason,
          });
          toast.success(
            mode === "delete" ? "Workspace deleted" : "Workspace restored",
          );
        }

        closeDialog();
        router.refresh();
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Unable to complete action.";
        toast.error("Action failed", { description: message });
      }
    });
  };

  const dialogTitle = {
    delete: "Delete workspace",
    restore: "Restore workspace",
    sync: "Sync subscription",
    trial: "Extend workspace trial",
  }[mode ?? "trial"];

  return (
    <>
      <Menu>
        <Menu.Button>
          <Button
            color="tertiary"
            rightIcon={<ArrowDown2Icon className="h-4 text-current!" />}
            type="button"
          >
            Actions
          </Button>
        </Menu.Button>
        <Menu.Items align="end" className="w-64">
          <Menu.Group>
            <Menu.Item
              onSelect={() => {
                setMode("trial");
              }}
            >
              <CalendarPlusIcon className="h-[1.15rem]" />
              Extend trial
            </Menu.Item>
            <Menu.Item
              onSelect={() => {
                setMode("sync");
              }}
            >
              <ReloadIcon className="h-[1.15rem]" />
              Sync subscription
            </Menu.Item>
            <Menu.Item
              className={isDeleted ? undefined : "text-danger"}
              onSelect={() => {
                setMode(isDeleted ? "restore" : "delete");
              }}
            >
              {isDeleted ? (
                <UndoIcon className="h-[1.15rem]" />
              ) : (
                <DeleteIcon className="h-[1.15rem] text-current!" />
              )}
              {isDeleted ? "Restore workspace" : "Delete workspace"}
            </Menu.Item>
          </Menu.Group>
          {workspace.stripeCustomerId || workspace.stripeSubscriptionId ? (
            <>
              <Menu.Separator />
              <Menu.Group>
                {workspace.stripeCustomerId ? (
                  <Menu.Item asChild>
                    <a
                      href={stripeCustomerUrl(workspace.stripeCustomerId)}
                      rel="noreferrer"
                      target="_blank"
                    >
                      <NewTabIcon className="h-[1.15rem]" />
                      Stripe customer
                    </a>
                  </Menu.Item>
                ) : null}
                {workspace.stripeSubscriptionId ? (
                  <Menu.Item asChild>
                    <a
                      href={stripeSubscriptionUrl(
                        workspace.stripeSubscriptionId,
                      )}
                      rel="noreferrer"
                      target="_blank"
                    >
                      <NewTabIcon className="h-[1.15rem]" />
                      Stripe subscription
                    </a>
                  </Menu.Item>
                ) : null}
              </Menu.Group>
            </>
          ) : null}
        </Menu.Items>
      </Menu>

      <Dialog
        onOpenChange={(open) => {
          if (!open) {
            closeDialog();
          }
        }}
        open={Boolean(mode)}
      >
        <Dialog.Content size="md">
          <form action={handleSubmit}>
            <Dialog.Header className="px-6">
              <Dialog.Title>
                <Text
                  as="span"
                  className="text-[1.35rem]"
                  fontWeight="semibold"
                >
                  {dialogTitle}
                </Text>
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Box className="border-border bg-surface-muted/70 rounded-lg border-[0.5px] p-3">
                <Text fontWeight="semibold">{workspace.name}</Text>
                <Text className="mt-1 text-[0.95rem]" color="muted">
                  Current trial: {formatDateTime(workspace.trialEndsOn)} ·{" "}
                  {formatTrialState(workspace.trialEndsOn)}
                </Text>
              </Box>
              <Box className="mt-4 space-y-4">
                {mode === "trial" ? (
                  <Input
                    defaultValue={defaultTrialEnd}
                    label="New trial end"
                    min={minTrialEnd}
                    name="trialEndsOn"
                    required
                    type="datetime-local"
                  />
                ) : null}
                <TextArea
                  className="min-h-28 leading-6"
                  label="Reason"
                  name="reason"
                  placeholder="Customer success request, billing remediation, security review..."
                  required
                />
              </Box>
            </Dialog.Body>
            <Dialog.Footer justify="between">
              <Button
                color="tertiary"
                disabled={isPending}
                onClick={closeDialog}
                type="button"
                variant="naked"
              >
                Cancel
              </Button>
              <Flex className="gap-2">
                <Button
                  color={mode === "delete" ? "danger" : "primary"}
                  loading={isPending}
                  loadingText="Saving..."
                  type="submit"
                >
                  Save action
                </Button>
              </Flex>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
