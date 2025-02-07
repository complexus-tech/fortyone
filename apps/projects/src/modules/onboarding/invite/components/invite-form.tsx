"use client";

import { Box, Text, Button } from "ui";
import { useState } from "react";
import { MemberRow } from "./member-row";

type Member = {
  email: string;
  role: string;
};

type InviteFormProps = {
  onFormChange: (members: Member[]) => void;
};

export const InviteForm = ({ onFormChange }: InviteFormProps) => {
  const [members, setMembers] = useState<Member[]>([
    { email: "", role: "member" },
  ]);

  const addMember = () => {
    setMembers([...members, { email: "", role: "member" }]);
  };

  const removeMember = (index: number) => {
    const newMembers = members.filter((_, i) => i !== index);
    setMembers(newMembers);
    onFormChange(newMembers);
  };

  const updateMember = (index: number, field: keyof Member, value: string) => {
    const newMembers = members.map((member, i) =>
      i === index ? { ...member, [field]: value } : member,
    );
    setMembers(newMembers);
    onFormChange(newMembers);
  };

  return (
    <Box className="text-black dark:text-white">
      <Text as="h1" className="mb-2 text-2xl font-semibold">
        Invite your teammates
      </Text>
      <Text className="mb-8 text-gray dark:text-gray-200">
        Complexus is meant to be used with your team. Invite some co-workers to
        test it out with.
      </Text>

      <Box className="space-y-4">
        {members.map((member, index) => (
          <MemberRow
            email={member.email}
            isRemovable={members.length > 1}
            key={index}
            onEmailChange={(email) => {
              updateMember(index, "email", email);
            }}
            onRemove={() => {
              removeMember(index);
            }}
            onRoleChange={(role) => {
              updateMember(index, "role", role);
            }}
            role={member.role}
          />
        ))}
      </Box>

      <Button
        className="mt-4 text-gray hover:text-gray-250 dark:text-gray-200 dark:hover:text-white"
        color="tertiary"
        onClick={addMember}
        variant="naked"
      >
        Add another
      </Button>
    </Box>
  );
};
