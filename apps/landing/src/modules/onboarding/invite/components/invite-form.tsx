"use client";

import { Box, Button } from "ui";
import { useState, useEffect } from "react";
import { PlusIcon } from "icons";
import { MemberRow } from "./member-row";

type Member = {
  email: string;
};

type InviteFormProps = {
  onFormChange: (members: Member[]) => void;
};

export const InviteForm = ({ onFormChange }: InviteFormProps) => {
  const [members, setMembers] = useState<Member[]>([
    { email: "" },
    { email: "" },
  ]);

  // Initialize parent state when component mounts
  useEffect(() => {
    onFormChange(members);
  }, [members, onFormChange]);

  const addMember = () => {
    const newMembers = [...members, { email: "" }];
    setMembers(newMembers);
    onFormChange(newMembers);
  };

  const removeMember = (index: number) => {
    const newMembers = members.filter((_, i) => i !== index);
    setMembers(newMembers);
    onFormChange(newMembers);
  };

  const updateMember = (index: number, email: string) => {
    const newMembers = members.map((member, i) =>
      i === index ? { email } : member,
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
            key={`member-${index}`}
            onEmailChange={(email) => {
              updateMember(index, email);
            }}
            onRemove={() => {
              removeMember(index);
            }}
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
