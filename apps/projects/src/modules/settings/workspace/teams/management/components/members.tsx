"use client";

import { useState } from "react";
import { Box, Button, Input } from "ui";
import { SearchIcon, PlusIcon } from "icons";
import type { Team } from "@/modules/teams/types";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { useTeamMembers } from "@/lib/hooks/team-members";
import { AssigneesMenu } from "@/components/ui";
import { useAddMemberMutation } from "@/modules/teams/hooks/add-member-mutation";
import { useMembers } from "@/lib/hooks/members";
import { TeamMemberRow } from "./team-member-row";

export const MembersSettings = ({ team }: { team: Team }) => {
  const { data: allMembers = [] } = useMembers();
  const { data: members = [] } = useTeamMembers(team.id);
  const { mutate: addMember } = useAddMemberMutation();
  const [search, setSearch] = useState("");

  const handleAddMember = (memberId: string) => {
    addMember({ teamId: team.id, memberId });
  };

  const isAllMembersAdded = allMembers.length === members.length;

  return (
    <Box>
      {members.length > 10 && (
        <Box className="mb-4">
          <Input
            leftIcon={<SearchIcon className="h-4" />}
            onChange={(e) => {
              setSearch(e.target.value);
            }}
            placeholder="Search members..."
            type="search"
            value={search}
            variant="solid"
          />
        </Box>
      )}

      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <AssigneesMenu>
              <AssigneesMenu.Trigger>
                <Button
                  color="tertiary"
                  disabled={isAllMembersAdded}
                  leftIcon={<PlusIcon className="text-white dark:text-white" />}
                >
                  Add Member
                </Button>
              </AssigneesMenu.Trigger>
              <AssigneesMenu.Items
                disallowEmptySelection
                excludeUsers={members.map(({ id }) => id)}
                onAssigneeSelected={(memberId) => {
                  if (memberId) {
                    handleAddMember(memberId);
                  }
                }}
                placeholder="Add member..."
              />
            </AssigneesMenu>
          }
          description="Manage team members and their roles."
          title="Team Members"
        />

        <Box className="divide-y divide-gray-100 dark:divide-dark-100">
          {members.map((member) => (
            <TeamMemberRow key={member.id} member={member} teamId={team.id} />
          ))}
        </Box>
      </Box>
    </Box>
  );
};
