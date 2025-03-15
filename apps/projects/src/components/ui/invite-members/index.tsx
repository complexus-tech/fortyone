"use client";

import { Button, Dialog, Select, TextArea, Text, Flex } from "ui";
import { useState, type Dispatch, type SetStateAction } from "react";
import { toast } from "sonner";
import { cn } from "lib";
import { useQueryClient } from "@tanstack/react-query";
import { InviteMembersIcon } from "icons";
import { useTeams } from "@/modules/teams/hooks/teams";
import { inviteMembers } from "@/modules/invitations/actions/invite";
import type { NewInvitation } from "@/modules/invitations/types";
import { useMembers } from "@/lib/hooks/members";
import { invitationKeys } from "@/constants/keys";

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
  const { data: teams = [] } = useTeams();
  const { data: members = [] } = useMembers();
  const [formState, setFormState] = useState<InviteFormState>({
    emails: "",
    role: "member",
    teamIds: [],
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleEmailsChange = (value: string) => {
    setFormState((prev) => ({ ...prev, emails: value }));
  };

  const handleRoleChange = (value: string) => {
    setFormState((prev) => ({ ...prev, role: value }));
  };

  const handleTeamToggle = (teamId: string) => {
    setFormState((prev) => {
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
    setIsSubmitting(true);
    const toastId = "invite-members";

    const emailList = formState.emails
      .split(/[,\n]/)
      .map((email) => email.trim())
      .filter((email) => email !== "");

    if (emailList.length === 0) {
      toast.warning("Validation error", {
        description: "Please enter at least one email address",
      });
      setIsSubmitting(false);
      return;
    }

    const invalidEmails = validateEmails(emailList);
    if (invalidEmails.length > 0) {
      toast.warning("Invalid emails", {
        description: `Invalid email format: ${invalidEmails.join(", ")}`,
      });
      setIsSubmitting(false);
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
      setIsSubmitting(false);
      return;
    }

    if (!formState.role) {
      toast.warning("Role is required", {
        description: "Please select a role",
      });
      setIsSubmitting(false);
      return;
    }

    if (formState.teamIds.length === 0) {
      toast.warning("At least one team is required", {
        description:
          "Members should be added to a team to be able to work together",
      });
      setIsSubmitting(false);
      return;
    }
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

    const res = await inviteMembers(invites);
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
        queryKey: invitationKeys.pending,
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
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content className="max-w-2xl">
        <Dialog.Header>
          <Dialog.Title className="flex items-center gap-2 px-6 pt-0.5 text-lg">
            <InviteMembersIcon className="relative top-px" />
            Invite members to your workspace
          </Dialog.Title>
        </Dialog.Header>
        <Dialog.Body className="pb-6">
          <Text className="mb-2" color="muted">
            Email addresses
          </Text>
          <TextArea
            className="border py-4 leading-normal dark:bg-transparent"
            onChange={(e) => {
              handleEmailsChange(e.target.value);
            }}
            placeholder="e.g, member1@example.com, member2@example.com"
            rows={4}
            value={formState.emails}
          />
          <Text className="mb-2 mt-6" color="muted">
            Role
          </Text>
          <Select onValueChange={handleRoleChange} value={formState.role}>
            <Select.Trigger className="h-[2.8rem] border bg-transparent px-4 text-base dark:bg-transparent">
              <Select.Input placeholder="Select role" />
            </Select.Trigger>
            <Select.Content>
              {ROLE_OPTIONS.map((role) => (
                <Select.Option
                  className="text-base"
                  key={role.id}
                  value={role.id}
                >
                  {role.name}
                </Select.Option>
              ))}
            </Select.Content>
          </Select>
          <Text className="mb-2 mt-6" color="muted">
            Teams - members will be added to the selected teams
          </Text>
          <Flex gap={2} wrap>
            {teams.map((team) => (
              <Button
                className={cn("dark:bg-transparent", {
                  "ring-2": formState.teamIds.includes(team.id),
                })}
                color="tertiary"
                key={team.id}
                leftIcon={
                  <span
                    className="size-2 rounded-full"
                    style={{ backgroundColor: team.color }}
                  />
                }
                onClick={() => {
                  handleTeamToggle(team.id);
                }}
                size="sm"
                title={team.name}
                variant="outline"
              >
                <span className="inline-block max-w-[12ch] truncate">
                  {team.name}
                </span>
              </Button>
            ))}
          </Flex>
          {formState.teamIds.length > 0 && (
            <Text className="mt-1 pl-0.5" color="muted" fontSize="sm">
              {formState.teamIds.length} team
              {formState.teamIds.length > 1 ? "s" : ""} selected
            </Text>
          )}
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
      </Dialog.Content>
    </Dialog>
  );
};
