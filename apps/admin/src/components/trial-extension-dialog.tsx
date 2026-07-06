"use client";

import { useState, useTransition } from "react";
import { useRouter } from "next/navigation";
import { CalendarPlusIcon } from "icons";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Input, Text, TextArea } from "ui";
import { updateWorkspaceTrial } from "@/lib/admin-api";
import { formatDateTime, formatTrialState } from "@/lib/format";
import type { WorkspaceSummary } from "@/lib/types";

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

export const TrialExtensionDialog = ({
  workspace,
}: {
  workspace: WorkspaceSummary;
}) => {
  const router = useRouter();
  const [open, setOpen] = useState(false);
  const [isPending, startTransition] = useTransition();
  const defaultTrialEnd = toDateTimeLocal(
    suggestedTrialEnd(workspace.trialEndsOn),
  );
  const minTrialEnd = toDateTimeLocal(new Date());

  const handleSubmit = (formData: FormData) => {
    const trialEndsOn = String(formData.get("trialEndsOn") ?? "");
    const reason = String(formData.get("reason") ?? "").trim();

    if (!trialEndsOn || !reason) {
      toast.error("Trial end date and reason are required.");
      return;
    }

    startTransition(async () => {
      try {
        await updateWorkspaceTrial(workspace.id, {
          trialEndsOn: new Date(trialEndsOn).toISOString(),
          reason,
        });
        toast.success("Trial updated", {
          description: `${workspace.name} now has an extended trial window.`,
        });
        setOpen(false);
        router.refresh();
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Unable to update trial.";
        toast.error("Trial update failed", { description: message });
      }
    });
  };

  return (
    <Dialog onOpenChange={setOpen} open={open}>
      <Dialog.Trigger asChild>
        <Button color="tertiary">
          <CalendarPlusIcon className="h-5" />
          Extend trial
        </Button>
      </Dialog.Trigger>
      <Dialog.Content size="md">
        <form action={handleSubmit}>
          <Dialog.Header className="px-6">
            <Dialog.Title>
              <Text as="span" className="text-[1.35rem]" fontWeight="semibold">
                Extend workspace trial
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
              <Input
                defaultValue={defaultTrialEnd}
                label="New trial end"
                min={minTrialEnd}
                name="trialEndsOn"
                required
                type="datetime-local"
              />
              <TextArea
                className="min-h-28 leading-6"
                label="Reason"
                name="reason"
                placeholder="Customer success request, sales-led onboarding extension, billing remediation..."
                required
              />
            </Box>
          </Dialog.Body>
          <Dialog.Footer justify="between">
            <Button
              color="tertiary"
              disabled={isPending}
              onClick={() => {
                setOpen(false);
              }}
              type="button"
              variant="naked"
            >
              Cancel
            </Button>
            <Flex className="gap-2">
              <Button
                color="primary"
                loading={isPending}
                loadingText="Updating..."
                type="submit"
              >
                Update trial
              </Button>
            </Flex>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog>
  );
};
