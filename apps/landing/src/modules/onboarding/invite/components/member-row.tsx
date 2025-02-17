"use client";

import { Box, Input, Select, Button } from "ui";
import { DeleteIcon } from "icons";

type MemberRowProps = {
  email: string;
  role: string;
  onEmailChange: (email: string) => void;
  onRoleChange: (role: string) => void;
  onRemove: () => void;
  isRemovable?: boolean;
};

export const MemberRow = ({
  email,
  role,
  onEmailChange,
  onRoleChange,
  onRemove,
  isRemovable = true,
}: MemberRowProps) => {
  return (
    <Box className="flex items-start gap-2 px-1">
      <Box className="flex-1">
        <Input
          className="bg-gray-50/30 dark:bg-dark-100/30"
          onChange={(e) => {
            onEmailChange(e.target.value);
          }}
          placeholder="colleague@company.com"
          type="email"
          value={email}
        />
      </Box>
      <Select onValueChange={onRoleChange} value={role}>
        <Select.Trigger className="h-[2.8rem] w-28 text-base">
          <Select.Input />
        </Select.Trigger>
        <Select.Content>
          <Select.Group>
            <Select.Option value="member">Member</Select.Option>
            <Select.Option value="admin">Admin</Select.Option>
            <Select.Option value="viewer">Viewer</Select.Option>
          </Select.Group>
        </Select.Content>
      </Select>
      {isRemovable ? (
        <Button
          asIcon
          className="md:h-[2.8rem]"
          color="tertiary"
          onClick={onRemove}
        >
          <DeleteIcon className="h-[1.1rem]" />
        </Button>
      ) : null}
    </Box>
  );
};
