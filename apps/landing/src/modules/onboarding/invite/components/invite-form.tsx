"use client";

import { Box, Text } from "ui";
import { useState, useEffect } from "react";
import { InfoIcon, PlusIcon } from "icons";
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

      {members.length < 5 ? (
        <button
          className="mt-3 flex items-center gap-1 opacity-70 transition hover:opacity-85"
          onClick={addMember}
          type="button"
        >
          <PlusIcon className="h-[1.1rem] text-foreground" />
          Add another colleague
        </button>
      ) : (
        <Text className="mt-3 flex items-center gap-1" color="muted">
          <InfoIcon className="h-[1.1rem]" />
          You can invite more colleagues later
        </Text>
      )}
    </Box>
  );
};
