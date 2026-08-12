"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { CalendarContentSkeleton } from "./components/calendar-skeleton";
import { CalendarHeader } from "./components/header";

const PersonalCalendar = dynamic(
  () =>
    import("./components/calendar").then((module) => module.PersonalCalendar),
  {
    loading: () => <CalendarContentSkeleton />,
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
