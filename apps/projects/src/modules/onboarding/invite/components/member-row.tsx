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
    <Box className="flex items-start gap-3">
      <Input
        className="flex-1 border-gray-100 bg-white text-black placeholder:text-gray dark:border-dark-50 dark:bg-dark-300 dark:text-white dark:placeholder:text-gray-200"
        onChange={(e) => {
          onEmailChange(e.target.value);
        }}
        placeholder="colleague@company.com"
        type="email"
        value={email}
      />
      <Select onValueChange={onRoleChange} value={role}>
        <Select.Trigger className="w-40 border-gray-100 bg-white text-black dark:border-dark-50 dark:bg-dark-300 dark:text-white">
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
          className="mt-1 text-danger hover:bg-danger/10 dark:text-danger dark:hover:bg-danger/5"
          color="danger"
          onClick={onRemove}
          variant="naked"
        >
          <DeleteIcon className="h-4 w-4" />
        </Button>
      ) : null}
    </Box>
  );
};
