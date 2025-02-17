"use client";

import { Box, Button } from "ui";
import { useState } from "react";
import { PlusIcon } from "icons";
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
      <Box className="space-y-4">
        {members.map((member, index) => (
          <MemberRow
            email={member.email}
            isRemovable={members.length > 1}
            // eslint-disable-next-line react/no-array-index-key -- we need a unique key for each member
            key={`${member.email}-${index}`}
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
        align="center"
        className="mt-3 px-0 dark:hover:bg-transparent"
        color="tertiary"
        leftIcon={<PlusIcon />}
        onClick={addMember}
        variant="naked"
      >
        Add another colleague
      </Button>
    </Box>
  );
};
