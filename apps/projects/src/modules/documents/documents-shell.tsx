"use client";

import type { ReactNode } from "react";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { DocumentsList } from "./documents-list";

export const DocumentsShell = ({ children }: { children: ReactNode }) => {
  return (
    <BodyContainer className="grid h-full min-h-0 min-w-0 overflow-hidden md:grid-cols-[320px_minmax(0,1fr)]">
      <Box className="hidden min-h-0 overflow-hidden md:block">
        <DocumentsList />
      </Box>
      <Box className="min-h-0 min-w-0 overflow-hidden">{children}</Box>
    </BodyContainer>
  );
};
