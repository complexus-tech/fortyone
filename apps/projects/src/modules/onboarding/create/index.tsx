"use client";

import { Box } from "ui";
import { Logo } from "@/components/ui/logo";
import { CreateWorkspaceForm } from "./components/create-workspace-form";

export const CreateWorkspace = ({ callbackUrl }: { callbackUrl?: string }) => {
  return (
    <Box className="w-full px-6 md:max-w-lg">
      <Logo asIcon />
      <CreateWorkspaceForm callbackUrl={callbackUrl} />
    </Box>
  );
};
