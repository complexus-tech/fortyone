import { Button, Dialog, Select, TextArea, Text, Checkbox } from "ui";
import { useState, type Dispatch, type SetStateAction } from "react";
import { toast } from "sonner";
import { useTeams } from "@/modules/teams/hooks/teams";
import { inviteMembers } from "@/modules/invitations/actions/invite";
import type { NewInvitation } from "@/modules/invitations/types";

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
  const { data: teams = [] } = useTeams();
  const [formState, setFormState] = useState<InviteFormState>({
    emails: "",
    role: "member", // Default to member role
    teamIds: [],
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Reset form when dialog opens
  if (!isOpen) {
    setFormState({
      emails: "",
      role: "member",
      teamIds: [],
    });
  }

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

    try {
      // Parse emails
      const emailList = formState.emails
        .split(/[,\n]/)
        .map((email) => email.trim())
        .filter((email) => email !== "");

      // Validate inputs
      if (emailList.length === 0) {
        toast.error("Please enter at least one email address");
        return;
      }

      const invalidEmails = validateEmails(emailList);
      if (invalidEmails.length > 0) {
        toast.error(`Invalid email format: ${invalidEmails.join(", ")}`);
        return;
      }

      if (!formState.role) {
        toast.error("Please select a role");
        return;
      }

      // Create invitation objects
      const invites: NewInvitation[] = emailList.map((email) => ({
        email,
        role: formState.role,
        teamIds: formState.teamIds.length > 0 ? formState.teamIds : undefined,
      }));

      // Call the server action
      const response = await inviteMembers(invites);

      if (response.error) {
        toast.error(response.error.message || "Failed to send invites");
        return;
      }

      // Show success toast
      toast.success("Invitations sent successfully");

      // Close dialog
      setIsOpen(false);
    } catch (err) {
      // Handle unexpected errors
      toast.error(
        err instanceof Error ? err.message : "An unknown error occurred",
      );
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Dialog onOpenChange={setIsOpen} open={isOpen}>
      <Dialog.Content className="max-w-2xl">
        <Dialog.Header className="border-b-[0.5px] border-gray-100 dark:border-dark-100">
          <Dialog.Title className="px-6 pt-0.5 text-lg">
            Invite members to your workspace
          </Dialog.Title>
        </Dialog.Header>

        <Dialog.Body className="py-6">
          <TextArea
            className="border-[0.5px] dark:bg-transparent"
            label="Email addresses"
            onChange={(e) => {
              handleEmailsChange(e.target.value);
            }}
            placeholder="Enter email addresses separated by commas or new lines"
            rows={3}
            value={formState.emails}
          />
          <Text className="mb-2 mt-6">Role</Text>
          <Select onValueChange={handleRoleChange} value={formState.role}>
            <Select.Trigger className="h-[2.8rem] border-[0.5px] px-4 text-base dark:bg-transparent">
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

          <Text className="mb-2 mt-6">Teams (Optional)</Text>

          <div className="max-h-48 overflow-y-auto rounded border-[0.5px] p-3 dark:bg-transparent">
            {teams.length === 0 ? (
              <Text className="py-2 text-center" color="muted">
                No teams available
              </Text>
            ) : (
              <div className="space-y-2">
                {teams.map((team) => (
                  <div className="flex items-center space-x-2" key={team.id}>
                    <Checkbox
                      checked={formState.teamIds.includes(team.id)}
                      id={`team-${team.id}`}
                      onCheckedChange={() => {
                        handleTeamToggle(team.id);
                      }}
                    />
                    <label
                      className="flex cursor-pointer items-center gap-2 text-sm"
                      htmlFor={`team-${team.id}`}
                    >
                      <span
                        className="h-2 w-2 rounded-full"
                        style={{ backgroundColor: team.color }}
                      />
                      {team.name}
                    </label>
                  </div>
                ))}
              </div>
            )}
          </div>
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
