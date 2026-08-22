"use client";

import { BreadCrumbs, Flex } from "ui";
import {
  HeaderContainer,
  MobileMenuButton,
  useAppCommandAction,
} from "@/components/shared";
import { useTerminology, useUserRole } from "@/hooks";

export const CalendarHeader = ({ onSchedule }: { onSchedule: () => void }) => {
  const { getTermDisplay } = useTerminology();
  const { userRole } = useUserRole();

  useAppCommandAction({
    disabled: userRole === "guest",
    id: "calendar:schedule-story",
    label: `Schedule ${getTermDisplay("storyTerm")}`,
    onSelect: onSchedule,
  });

  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Calendar",
            },
          ]}
        />
      </Flex>
    </HeaderContainer>
  );
};
