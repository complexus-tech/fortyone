"use client";

import { useState, useTransition } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  ArrowDown2Icon,
  FileLockedIcon,
  FileUnlockedIcon,
  LockIcon,
  NewTabIcon,
  ReloadIcon,
  UndoIcon,
} from "icons";
import { toast } from "sonner";
import { Box, Button, Dialog, Flex, Menu, Text, TextArea } from "ui";
import {
  requestUserSessionRevocation,
  resetUserAIUsage,
  updateUserState,
} from "@/lib/admin-api";
import type { UserSummary } from "@/lib/types";

type UserAction =
  | "activate"
  | "deactivate"
  | "grant_internal"
  | "revoke_internal"
  | "revoke_sessions"
  | "reset_ai_usage";

const actionCopy: Record<UserAction, { title: string; submit: string }> = {
  activate: {
    title: "Reactivate user",
    submit: "Reactivate user",
  },
  deactivate: {
    title: "Deactivate user",
    submit: "Deactivate user",
  },
  grant_internal: {
    title: "Grant internal access",
    submit: "Grant access",
  },
  revoke_internal: {
    title: "Revoke internal access",
    submit: "Revoke access",
  },
  revoke_sessions: {
    title: "Request session revocation",
    submit: "Record request",
  },
  reset_ai_usage: {
    title: "Reset AI usage",
    submit: "Reset usage",
  },
};

export const UserActionsMenu = ({ user }: { user: UserSummary }) => {
  const router = useRouter();
  const [action, setAction] = useState<UserAction | null>(null);
  const [isPending, startTransition] = useTransition();

  const closeDialog = () => {
    setAction(null);
  };

  const handleSubmit = (formData: FormData) => {
    if (!action) {
      return;
    }

    const reason = String(formData.get("reason") ?? "").trim();
    if (!reason) {
      toast.error("Reason is required.");
      return;
    }

    startTransition(async () => {
      try {
        if (action === "reset_ai_usage") {
          const reset = await resetUserAIUsage(user.id, { reason });
          toast.success("AI usage reset", {
            description: `${reset.previousMessageCount} messages cleared across ${reset.workspaceCount} workspace${reset.workspaceCount === 1 ? "" : "s"}.`,
          });
        } else if (action === "revoke_sessions") {
          await requestUserSessionRevocation(user.id, { reason });
        } else if (action === "activate" || action === "deactivate") {
          await updateUserState(user.id, {
            isActive: action === "activate",
            reason,
          });
        } else {
          await updateUserState(user.id, {
            isInternal: action === "grant_internal",
            reason,
          });
        }

        if (action !== "reset_ai_usage") {
          toast.success("User action recorded");
        }
        closeDialog();
        router.refresh();
      } catch (error) {
        const message =
          error instanceof Error ? error.message : "Unable to update user.";
        toast.error("User action failed", { description: message });
      }
    });
  };

  const copy = action ? actionCopy[action] : null;

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
        <Menu.Items align="end" className="w-80">
          <Menu.Group>
            <Menu.Item
              className={user.isActive ? "text-danger" : undefined}
              onSelect={() => {
                setAction(user.isActive ? "deactivate" : "activate");
              }}
            >
              {user.isActive ? (
                <FileLockedIcon className="h-[1.15rem] text-current!" />
              ) : (
                <FileUnlockedIcon className="h-[1.15rem]" />
              )}
              {user.isActive ? "Deactivate user" : "Reactivate user"}
            </Menu.Item>
            <Menu.Item
              onSelect={() => {
                setAction(
                  user.isInternal ? "revoke_internal" : "grant_internal",
                );
              }}
            >
              {user.isInternal ? (
                <UndoIcon className="h-[1.15rem]" />
              ) : (
                <LockIcon className="h-[1.15rem]" />
              )}
              {user.isInternal
                ? "Revoke internal access"
                : "Grant internal access"}
            </Menu.Item>
            <Menu.Item
              onSelect={() => {
                setAction("revoke_sessions");
              }}
            >
              <FileLockedIcon className="h-[1.15rem]" />
              Request session revocation
            </Menu.Item>
            <Menu.Item
              onSelect={() => {
                setAction("reset_ai_usage");
              }}
            >
              <ReloadIcon className="h-[1.15rem]" />
              Reset AI usage
            </Menu.Item>
          </Menu.Group>
          {user.lastUsedWorkspaceId ? (
            <>
              <Menu.Separator />
              <Menu.Group>
                <Menu.Item asChild>
                  <Link href={`/workspaces/${user.lastUsedWorkspaceId}`}>
                    <NewTabIcon className="h-[1.15rem]" />
                    Last used workspace
                  </Link>
                </Menu.Item>
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
        open={Boolean(action)}
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
                  {copy?.title}
                </Text>
              </Dialog.Title>
            </Dialog.Header>
            <Dialog.Body>
              <Box className="border-border bg-surface-muted/70 rounded-lg border-[0.5px] p-3">
                <Text fontWeight="semibold">
                  {user.fullName || user.username}
                </Text>
                <Text className="mt-1 text-[0.95rem]" color="muted">
                  {user.email}
                </Text>
              </Box>
              <Box className="mt-4">
                {action === "reset_ai_usage" ? (
                  <Text className="mb-4 leading-6" color="muted">
                    This resets the user&apos;s current-month Maya message usage
                    in every workspace they belong to. Their chat history is
                    preserved.
                  </Text>
                ) : null}
                <TextArea
                  className="min-h-28 leading-6"
                  label="Reason"
                  name="reason"
                  placeholder={
                    action === "reset_ai_usage"
                      ? "Customer support request, billing correction..."
                      : "Security review, support escalation, internal access request..."
                  }
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
                  color={action === "deactivate" ? "danger" : "primary"}
                  loading={isPending}
                  loadingText="Saving..."
                  type="submit"
                >
                  {copy?.submit ?? "Save action"}
                </Button>
              </Flex>
            </Dialog.Footer>
          </form>
        </Dialog.Content>
      </Dialog>
    </>
  );
};
