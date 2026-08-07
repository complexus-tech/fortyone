"use client";

import type { ReactNode } from "react";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { DocumentsList } from "./documents-list";

export const DocumentsShell = ({ children }: { children: ReactNode }) => {
  return (
    <BodyContainer className="grid h-dvh min-w-0 md:grid-cols-[320px_minmax(0,1fr)]">
      <Box className="hidden md:block">
        <DocumentsList />
      </Box>
      <Box className="min-w-0">{children}</Box>
    </BodyContainer>
  );
};
