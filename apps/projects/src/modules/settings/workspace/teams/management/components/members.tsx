"use client";

import { useState } from "react";
import { Box, Button, Select, Input, Dialog } from "ui";
import { SearchIcon, PlusIcon } from "icons";
import type { Team } from "@/modules/teams/types";
import { SectionHeader } from "@/modules/settings/components/section-header";
import { TeamMemberRow } from "./team-member-row";

export const MembersSettings = ({ team }: { team: Team }) => {
  const [search, setSearch] = useState("");
  const [isAddOpen, setIsAddOpen] = useState(false);
  const [selectedMember, setSelectedMember] = useState<string>("");
  const [selectedRole, setSelectedRole] = useState<"admin" | "member">(
    "member",
  );

  const handleAddMember = () => {
    if (!selectedMember) return;

    setIsAddOpen(false);
    setSelectedMember("");
  };

  return (
    <Box>
      <Box className="rounded-lg border border-gray-100 bg-white dark:border-dark-100 dark:bg-dark-100/40">
        <SectionHeader
          action={
            <Button
              color="primary"
              leftIcon={<PlusIcon />}
              onClick={() => {
                setIsAddOpen(true);
              }}
            >
              Add Member
            </Button>
          }
          description="Manage team members and their roles."
          title="Team Members"
        />

        <Box className="border-b border-gray-100 px-6 py-4 dark:border-dark-100">
          <Box className="relative max-w-md">
            <SearchIcon className="text-gray-400 absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2" />
            <Input
              className="pl-9"
              onChange={(e) => {
                setSearch(e.target.value);
              }}
              placeholder="Search members..."
              type="search"
              value={search}
            />
          </Box>
        </Box>

        <Box className="divide-y divide-gray-100 dark:divide-dark-100">
          <TeamMemberRow
            key="1"
            member={{
              id: "1",
              username: "john.doe",
              fullName: "John Doe",
              email: "john.doe@example.com",
              avatarUrl: "",
              isActive: true,
              createdAt: "2021-01-01",
              updatedAt: "2021-01-01",
            }}
            onRemove={() => {}}
            teamId={team.id}
            userRole="member"
          />
        </Box>
      </Box>

      <Dialog onOpenChange={setIsAddOpen} open={isAddOpen}>
        <Dialog.Content>
          <Dialog.Header>
            <Dialog.Title>Add Team Member</Dialog.Title>
          </Dialog.Header>

          <Box className="space-y-4 p-6">
            <Select
              defaultValue={selectedMember || undefined}
              onValueChange={(value: string) => {
                setSelectedMember(value);
              }}
            >
              <Select.Trigger>
                <Select.Input placeholder="Select a member" />
              </Select.Trigger>
              <Select.Content>What</Select.Content>
            </Select>

            <Select
              defaultValue={selectedRole}
              onValueChange={(value: string) => {
                setSelectedRole(value as "admin" | "member");
              }}
            >
              <Select.Trigger>
                <Select.Input />
              </Select.Trigger>
              <Select.Content>
                <Select.Option value="admin">Admin</Select.Option>
                <Select.Option value="member">Member</Select.Option>
              </Select.Content>
            </Select>
          </Box>

          <Dialog.Footer>
            <Button
              onClick={() => {
                setIsAddOpen(false);
              }}
              variant="outline"
            >
              Cancel
            </Button>
            <Button disabled={!selectedMember} onClick={handleAddMember}>
              Add Member
            </Button>
          </Dialog.Footer>
        </Dialog.Content>
      </Dialog>
    </Box>
  );
};
