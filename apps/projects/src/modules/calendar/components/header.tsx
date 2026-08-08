"use client";

import { CalendarIcon, PlusIcon } from "icons";
import { Box, BreadCrumbs, Button, Flex } from "ui";
import { HeaderContainer, MobileMenuButton } from "@/components/shared";
import { useTerminology } from "@/hooks";

export const CalendarHeader = ({ onSchedule }: { onSchedule: () => void }) => {
  const { getTermDisplay } = useTerminology();

  return (
    <HeaderContainer className="justify-between">
      <Flex align="center" gap={2}>
        <MobileMenuButton />
        <BreadCrumbs
          breadCrumbs={[
            {
              name: "Calendar",
              icon: <CalendarIcon />,
            },
          ]}
        />
      </Flex>
      <Box className="hidden md:block">
        <Button
          className="shrink-0"
          color="invert"
          data-header-schedule-story-button
          leftIcon={<PlusIcon className="text-current dark:text-current" />}
          onClick={onSchedule}
          size="sm"
        >
          Schedule {getTermDisplay("storyTerm")}
        </Button>
      </Box>
    </HeaderContainer>
  );
};
