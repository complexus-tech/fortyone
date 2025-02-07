"use client";

import { useState } from "react";
import { Box, Button, Flex } from "ui";
import { InviteForm } from "./components/invite-form";

type Member = {
  email: string;
  role: string;
};

export const InviteTeam = () => {
  const [members, setMembers] = useState<Member[]>([]);
  const [isSkipped, setIsSkipped] = useState(false);

  const isValid = isSkipped || members.some((m) => m.email.trim() !== "");

  return (
    <Box className="min-h-screen">
      <Flex align="center" className="min-h-screen" justify="center">
        <Box className="w-full max-w-xl px-4">
          <Box className="rounded-xl bg-white p-6 shadow-sm dark:bg-dark-300">
            <InviteForm onFormChange={setMembers} />
            <button
              className="mt-4 text-sm text-gray hover:text-gray-250 dark:text-gray-200 dark:hover:text-white"
              onClick={() => {
                setIsSkipped(!isSkipped);
              }}
              type="button"
            >
              {isSkipped ? "Want to invite team members?" : "I'll do it later"}
            </button>
          </Box>
          <Button
            align="center"
            className="mt-6"
            disabled={!isValid}
            fullWidth
            rounded="lg"
            size="lg"
          >
            {isSkipped ? "Skip for now" : "Continue"}
          </Button>
        </Box>
      </Flex>
    </Box>
  );
};
