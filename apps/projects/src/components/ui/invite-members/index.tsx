"use client";

import {
  Button,
  Dialog,
  Select,
  TextArea,
  Text,
  Flex,
  Wrapper,
  Checkbox,
} from "ui";
import { useState, type Dispatch, type SetStateAction } from "react";
import { toast } from "sonner";
import { useQueryClient } from "@tanstack/react-query";
import { InviteMembersIcon, WarningIcon } from "icons";
import { useRouter } from "next/navigation";
import { useTeams } from "@/modules/teams/hooks/teams";
import { inviteMembers } from "@/modules/invitations/actions/invite";
import type { NewInvitation } from "@/modules/invitations/types";
import { useMembers } from "@/lib/hooks/members";
import { invitationKeys } from "@/constants/keys";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useUserRole, useWorkspacePath } from "@/hooks";
import { FeatureGuard } from "../feature-guard";

type InviteFormState = {
  emails: string;
  role: string;
  teamIds: string[];
};

const ROLE_OPTIONS = [
  { id: "admin", name: "Admin" },
  { id: "member", name: "Member" },
  { id: "guest", name: "Guest" },
];

export const InviteMembersDialog = ({
  isOpen,
  setIsOpen,
}: {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}) => {
  const queryClient = useQueryClient();
  const { userRole } = useUserRole();
  const { workspaceSlug, withWorkspace } = useWorkspacePath();
  const { remaining, tier, getLimit } = useSubscriptionFeatures();
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const [formState, setFormState] = useState<InviteFormState>({
    emails: "",
    role: "admin",
    teamIds: teams.length > 0 ? [teams[0].id] : [],
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const router = useRouter();
  const handleEmailsChange = (value: string) => {
    setFormState((prev) => ({ ...prev, emails: value }));
  };

  const handleRoleChange = (value: string) => {
    setFormState((prev) => ({ ...prev, role: value }));
  };

  const handleTeamToggle = (teamId: string) => {
    setFormState((prev) => {
      if (prev.teamIds.includes(teamId) && prev.teamIds.length === 1) {
        toast.warning("At least one team is required", {
          description: "Members should be added to at least one team",
        });
        return prev;
      }

      const teamIds = prev.teamIds.includes(teamId)
        ? prev.teamIds.filter((id) => id !== teamId)
        : [...prev.teamIds, teamId];

      return { ...prev, teamIds };
    });
  };

  const validateEmails = (emails: string[]): string[] => {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    return emails.filter((email) => !emailRegex.test(email));
  };

  const handleSubmit = async () => {
    const toastId = "invite-members";

    const emailList = formState.emails
      .split(/[,\n]/)
      .map((email) => email.trim())
      .filter((email) => email !== "");

    if (emailList.length === 0) {
      toast.warning("Validation error", {
        description: "Please enter at least one email address",
      });
      return;
    }

    if (remaining("maxMembers", members.length) === 0) {
      toast.warning("Reached members limit", {
        description: "Upgrade to a higher plan to invite more members",
        action:
          userRole === "admin"
            ? {
                label: "Upgrade",
                onClick: () => {
                  router.push("/settings/workspace/billing");
                },
              }
            : undefined,
      });
      return;
    }

    const invalidEmails = validateEmails(emailList);
    if (invalidEmails.length > 0) {
      toast.warning("Invalid emails", {
        description: `Invalid email format: ${invalidEmails.join(", ")}`,
      });
      return;
    }

    // Filter out existing members
    const existingEmails = members.map((member) => member.email.toLowerCase());
    const newEmails = emailList.filter(
      (email) => !existingEmails.includes(email.toLowerCase()),
    );
    const existingCount = emailList.length - newEmails.length;

    // If all emails belong to existing members
    if (newEmails.length === 0) {
      toast.warning(
        emailList.length === 1
          ? "Member already exists"
          : "Members already exist",
        {
          description:
            emailList.length === 1
              ? "The entered email address already belongs to the workspace"
              : "All entered email addresses already belong to the workspace",
        },
      );
      return;
    }

    if (!formState.role) {
      toast.warning("Role is required", {
        description: "Please select a role",
      });
      return;
    }

    setIsSubmitting(true);
    toast.loading("Sending invites", {
      id: toastId,
      description: "Please wait...",
    });
    // Create invitation objects only for new emails
    const invites: NewInvitation[] = newEmails.map((email) => ({
      email,
      role: formState.role,
      teamIds: formState.teamIds.length > 0 ? formState.teamIds : undefined,
    }));

    const res = await inviteMembers(invites, workspaceSlug);
    if (res.error?.message) {
      toast.error("Failed to send invites", {
        description: res.error.message,
        id: toastId,
      });
      setIsSubmitting(false);
      return;
    }
    if (existingCount > 0) {
      toast.info(`${existingCount} emails skipped`, {
        description: `Invitations sent to ${newEmails.length} email(s), ${existingCount} email(s) skipped (already members).`,
        id: toastId,
      });
    } else {
      queryClient.invalidateQueries({
        queryKey: invitationKeys.pending(workspaceSlug),
      });
      toast.info("Success", {
        description: "Invitations sent to member emails",
        id: toastId,
      });
    }
    setIsOpen(false);
    setIsSubmitting(false);
  };

  return (
    <Dialog
      onOpenChange={setIsOpen}
      open={userRole !== "admin" ? false : isOpen}
    >
      <Dialog.Content className="max-w-2xl">
        <Dialog.Header>
          <Dialog.Title className="flex items-center gap-2 px-6 pt-0.5 text-lg">
            <InviteMembersIcon className="relative top-px" />
            Invite members to your workspace
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Description>
          Adding members to your workspace will add more seats to your plan once
          the member accepts the invitation.
        </Dialog.Description>
        <FeatureGuard
          count={members.length}
          fallback={
            <Dialog.Body className="mt-2 pb-6">
              <Wrapper className="border-warning bg-warning/10 dark:border-warning/20 dark:bg-warning/10 flex items-center justify-between gap-3 border p-4">
                <Flex align="center" gap={3}>
                  <WarningIcon className="text-warning dark:text-warning shrink-0" />
                  <Text>
                    You&apos;ve reached the limit of {getLimit("maxMembers")}{" "}
                    members on your {tier.replace("free", "hobby")} plan.{" "}
                    {userRole === "admin"
                      ? `Upgrade ${tier === "pro" ? "to business" : ""} to invite more members.`
                      : "Ask your admin to upgrade to invite more members."}
                  </Text>
                </Flex>
                {userRole === "admin" && (
                  <Button
                    className="shrink-0"
                    color="warning"
                    href={withWorkspace("/settings/workspace/billing")}
                  >
                    Upgrade now
                  </Button>
                )}
              </Wrapper>
            </Dialog.Body>
          }
          feature="maxMembers"
        >
          <Dialog.Body className="mt-2 pb-6">
            <Text className="mb-2" color="muted">
              Email addresses*
            </Text>
            <TextArea
              className="border-border resize-none border bg-transparent py-4 leading-normal"
              onChange={(e) => {
                handleEmailsChange(e.target.value);
              }}
              placeholder="e.g, member1@example.com, member2@example.com"
              rows={4}
              value={formState.emails}
            />
            <Text className="mt-6 mb-2" color="muted">
              Role
            </Text>
            <Select onValueChange={handleRoleChange} value={formState.role}>
              <Select.Trigger className="h-[2.8rem] border bg-white px-4 text-base dark:bg-transparent">
                <Select.Input placeholder="Select role" />
              </Select.Trigger>
              <Select.Content>
                {ROLE_OPTIONS.map((role) => (
                  <Select.Option
                    className="text-base"
                    disabled={
                      (tier === "free" || tier === "trial") &&
                      ["guest", "member"].includes(role.id)
                    }
                    key={role.id}
                    value={role.id}
                  >
                    {role.name}
                    {(tier === "free" || tier === "trial") &&
                      role.id !== "admin" && (
                        <span className="italic">
                          {" "}
                          - upgrade to invite members with role {role.name}
                        </span>
                      )}
                  </Select.Option>
                ))}
              </Select.Content>
            </Select>
            <Text className="mt-6 mb-2" color="muted">
              Teams - members will be added to the selected teams
            </Text>
            <Flex gap={2} wrap>
              {teams.map((team) => (
                <Button
                  className="gap-2 dark:bg-transparent"
                  color="tertiary"
                  key={team.id}
                  leftIcon={
                    <Checkbox
                      checked={formState.teamIds.includes(team.id)}
                      className="rounded-[0.4rem]"
                    />
                  }
                  onClick={() => {
                    handleTeamToggle(team.id);
                  }}
                  rounded="full"
                  title={team.name}
                  variant="outline"
                >
                  <span className="inline-block max-w-[12ch] truncate">
                    {team.name}
                  </span>
                </Button>
              ))}
            </Flex>
          </Dialog.Body>
          <Dialog.Footer className="justify-end">
            <Button
              loading={isSubmitting}
              loadingText="Sending"
              onClick={handleSubmit}
            >
              Send invites
            </Button>
          </Dialog.Footer>
        </FeatureGuard>
      </Dialog.Content>
    </Dialog>
  );
};
