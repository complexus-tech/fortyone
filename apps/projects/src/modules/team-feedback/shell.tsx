"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { cn } from "lib";
import { Box } from "ui";
import { BodyContainer } from "@/components/shared";
import { ListTeamFeedback } from "./list";

export const TeamFeedbackShell = ({
  children,
  teamId,
}: {
  children: ReactNode;
  teamId: string;
}) => {
  const pathname = usePathname();
  const hasSelectedFeedback = pathname.includes(`/teams/${teamId}/feedback/`);

  return (
    <BodyContainer className="grid h-dvh md:grid-cols-[340px_auto]">
      <Box className={cn(hasSelectedFeedback ? "hidden md:block" : "block")}>
        <ListTeamFeedback />
      </Box>
      <Box className={cn(hasSelectedFeedback ? "block" : "hidden", "md:block")}>
        {children}
      </Box>
    </BodyContainer>
  );
};
