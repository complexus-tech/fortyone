"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { Box, Skeleton } from "ui";
import { CalendarHeader } from "./components/header";

const PersonalCalendar = dynamic(
  () =>
    import("./components/calendar").then((module) => module.PersonalCalendar),
  {
    loading: () => (
      <Box className="flex h-[calc(100dvh-4rem)] flex-col overflow-hidden">
        <Skeleton className="h-18 w-full shrink-0 rounded-none" />
        <Skeleton className="min-h-0 w-full flex-1 rounded-none" />
      </Box>
    ),
    ssr: false,
  },
);

export const CalendarPage = () => {
  const [isScheduleDialogOpen, setIsScheduleDialogOpen] = useState(false);

  return (
    <>
      <CalendarHeader
        onSchedule={() => {
          setIsScheduleDialogOpen(true);
        }}
      />
      <PersonalCalendar
        isScheduleDialogOpen={isScheduleDialogOpen}
        onScheduleDialogOpenChange={setIsScheduleDialogOpen}
      />
    </>
  );
};
