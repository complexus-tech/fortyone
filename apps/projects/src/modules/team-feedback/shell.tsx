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
    <BodyContainer className="grid h-full min-h-0 min-w-0 overflow-hidden md:grid-cols-[340px_minmax(0,1fr)]">
      <Box
        className={cn(
          "min-h-0 min-w-0 overflow-hidden",
          hasSelectedFeedback ? "hidden md:block" : "block",
        )}
      >
        <ListTeamFeedback />
      </Box>
      <Box
        className={cn(
          "min-h-0 min-w-0 overflow-hidden md:block",
          hasSelectedFeedback ? "block" : "hidden",
        )}
      >
        {children}
      </Box>
    </BodyContainer>
  );
};
