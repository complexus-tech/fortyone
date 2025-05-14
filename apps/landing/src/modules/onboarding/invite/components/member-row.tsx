"use client";

import { Box, Input, Button } from "ui";
import { DeleteIcon } from "icons";

type MemberRowProps = {
  email: string;
  onEmailChange: (email: string) => void;
  onRemove: () => void;
  isRemovable?: boolean;
};

export const MemberRow = ({
  email,
  onEmailChange,
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
