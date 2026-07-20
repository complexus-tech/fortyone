"use client";

import { useState } from "react";
import {
  Box,
  Flex,
  Text,
  Button,
  Avatar,
  Menu,
  Input,
  TextArea,
  Dialog,
} from "ui";
import { MoreHorizontalIcon, DeleteIcon, EditIcon } from "icons";
import { useSession } from "@/lib/auth/client";
import type { Member } from "@/types";
import { ConfirmDialog, RowWrapper } from "@/components/ui";
import { useUserRole } from "@/hooks/role";
import { useSubscriptionFeatures } from "@/lib/hooks/subscription-features";
import { useRemoveMemberMutation } from "@/modules/teams/hooks/remove-member-mutation";
import { useUpdateTeamMemberAIContextMutation } from "@/modules/teams/hooks/update-team-member-ai-context-mutation";
import { openDialogAfterMenuClose } from "@/utils/menu-dialog-state";

type TeamMemberRowProps = {
  member: Member;
  teamId: string;
};

export const TeamMemberRow = ({ member, teamId }: TeamMemberRowProps) => {
  const { data: session } = useSession();
  const { userRole } = useUserRole();
  const { hasFeature } = useSubscriptionFeatures();
  const { mutate: removeMember } = useRemoveMemberMutation();
  const [isRemoveOpen, setIsRemoveOpen] = useState(false);
  const [isRoleDialogOpen, setIsRoleDialogOpen] = useState(false);
  const isCurrentUser = session ? session.user.id === member.id : false;
  const isAdmin = userRole === "admin";
  const canUseBackgroundMaya = hasFeature("backgroundMaya");

  const handleRemoveMember = () => {
    removeMember({ teamId, memberId: member.id });
    setIsRemoveOpen(false);
  };

  return (
    <RowWrapper className="py-4 last:border-b-0 md:px-6">
      <Box className="w-full">
        <Flex align="center" gap={3} justify="between">
          <Flex align="center" gap={3}>
            <Avatar name={member.fullName} src={member.avatarUrl} />
            <Box>
              <Text className="font-medium">
                {member.fullName}
                <Text as="span" className="hidden md:inline" color="muted">
                  ({member.username})
                </Text>
              </Text>
              <Text color="muted">{member.email}</Text>
            </Box>
          </Flex>

          <Flex align="center" gap={3}>
            <Text
              className="hidden w-20 first-letter:uppercase md:block"
              color="muted"
            >
              {member.role}
            </Text>
            <Menu>
              <Menu.Button>
                <Button
                  asIcon
                  color="tertiary"
                  leftIcon={<MoreHorizontalIcon />}
                  size="sm"
                >
                  <span className="sr-only">More options</span>
                </Button>
              </Menu.Button>
              <Menu.Items align="end">
                <Menu.Group>
                  {isAdmin ? (
                    <Menu.Item
                      onSelect={() => {
                        openDialogAfterMenuClose(setIsRoleDialogOpen);
                      }}
                    >
                      <EditIcon className="h-[1.15rem]" />
                      Set work focus
                    </Menu.Item>
                  ) : null}
                  <Menu.Item
                    onSelect={() => {
                      openDialogAfterMenuClose(setIsRemoveOpen);
                    }}
                  >
                    <DeleteIcon className="h-[1.15rem]" />
                    {isCurrentUser ? "Leave team" : "Remove member"}
                  </Menu.Item>
                </Menu.Group>
              </Menu.Items>
            </Menu>
          </Flex>
        </Flex>
      </Box>

      <Dialog onOpenChange={setIsRoleDialogOpen} open={isRoleDialogOpen}>
        <Dialog.Content size="md">
          <Dialog.Header>
            <Dialog.Title className="px-6 pt-0.5 text-lg">
              Work focus
            </Dialog.Title>
            <Dialog.Description className="mt-2">
              Describe what {member.fullName} usually works on so Maya can
              choose the right owner when assigning work.
            </Dialog.Description>
          </Dialog.Header>
          <Dialog.Body>
            <TeamMemberAIContextEditor
              canUseBackgroundMaya={canUseBackgroundMaya}
              key={`${member.id}:${member.teamAiRoleTitle ?? ""}:${member.teamAiRoleDescription ?? ""}`}
              member={member}
              onSaved={() => {
                setIsRoleDialogOpen(false);
              }}
              teamId={teamId}
            />
          </Dialog.Body>
        </Dialog.Content>
      </Dialog>

      <ConfirmDialog
        confirmText={isCurrentUser ? "Leave team" : "Remove member"}
        description={
          isCurrentUser
            ? `Are you sure you want to leave this team? ${userRole === "admin" ? "You can always join again later." : ""}`
            : `Are you sure you want to remove ${member.fullName} from this team?`
        }
        isOpen={isRemoveOpen}
        onClose={() => {
          setIsRemoveOpen(false);
        }}
        onConfirm={handleRemoveMember}
        title={isCurrentUser ? "Leave team" : "Remove member"}
      />
    </RowWrapper>
  );
};

const TeamMemberAIContextEditor = ({
  canUseBackgroundMaya,
  member,
  onSaved,
  teamId,
}: {
  canUseBackgroundMaya: boolean;
  member: Member;
  onSaved: () => void;
  teamId: string;
}) => {
  const { mutate: updateAIContext, isPending: isUpdatingAIContext } =
    useUpdateTeamMemberAIContextMutation();
  const [roleTitle, setRoleTitle] = useState(member.teamAiRoleTitle ?? "");
  const [roleDescription, setRoleDescription] = useState(
    member.teamAiRoleDescription ?? "",
  );
  const hasAIContextChanges =
    roleTitle.trim() !== (member.teamAiRoleTitle ?? "") ||
    roleDescription.trim() !== (member.teamAiRoleDescription ?? "");

  const handleSaveAIContext = () => {
    updateAIContext(
      {
        teamId,
        memberId: member.id,
        roleTitle,
        roleDescription,
      },
      {
        onSuccess: (res) => {
          if (!res.error?.message) {
            onSaved();
          }
        },
      },
    );
  };

  return (
    <Box className="grid gap-4">
      <Input
        disabled={!canUseBackgroundMaya}
        label="Role"
        onChange={(event) => {
          setRoleTitle(event.target.value);
        }}
        placeholder="Frontend engineer"
        value={roleTitle}
      />
      <TextArea
        className="min-h-11 py-2 leading-6"
        disabled={!canUseBackgroundMaya}
        label="What they usually work on"
        onChange={(event) => {
          setRoleDescription(event.target.value);
        }}
        placeholder="Works mostly on React, design systems, and frontend bugs."
        rows={3}
        value={roleDescription}
      />
      <Button
        className="justify-center"
        color="primary"
        disabled={
          !canUseBackgroundMaya || !hasAIContextChanges || isUpdatingAIContext
        }
        onClick={handleSaveAIContext}
        size="md"
        type="button"
      >
        Save
      </Button>
      {!canUseBackgroundMaya && (
        <Text color="muted">
          Maya background assignment is available on paid plans.
        </Text>
      )}
      {canUseBackgroundMaya &&
        !member.teamAiRoleTitle &&
        !member.teamAiRoleDescription &&
        member.inferredTeamAiRoleTitle ? (
          <Text color="muted">
            Maya has learned that this person usually works as a{" "}
            {member.inferredTeamAiRoleTitle.toLowerCase()} from recent team
            work.
          </Text>
        ) : null}
    </Box>
  );
};
