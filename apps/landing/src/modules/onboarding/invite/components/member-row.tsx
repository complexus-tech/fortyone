"use client";

import { Box, Input, Button } from "ui";
import { CloseIcon } from "icons";
import { cn } from "lib";

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
    <Box className="flex items-start">
      <Box className="flex-1">
        <Input
          className={cn("h-[2.6rem] focus:ring-0", {
            "rounded-r-none": isRemovable,
          })}
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
          className="h-[2.6rem] rounded-l-none border-l-0 dark:border-dark-100 dark:bg-dark-300/20 md:h-[2.7rem]"
          color="tertiary"
          onClick={onRemove}
        >
          <CloseIcon className="h-[1.1rem]" />
          <span className="sr-only">Remove</span>
        </Button>
      ) : null}
    </Box>
  );
};
