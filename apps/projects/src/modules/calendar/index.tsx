"use client";

import { useState } from "react";
import dynamic from "next/dynamic";
import { CalendarRouteSkeleton } from "./components/calendar-skeleton";

const PersonalCalendar = dynamic(
  () =>
    import("./components/calendar").then((module) => module.PersonalCalendar),
  {
    loading: () => <CalendarRouteSkeleton />,
    ssr: false,
  },
);

export const CalendarPage = () => {
  const [isScheduleDialogOpen, setIsScheduleDialogOpen] = useState(false);

  return (
    <PersonalCalendar
      isScheduleDialogOpen={isScheduleDialogOpen}
      onScheduleDialogOpenChange={setIsScheduleDialogOpen}
    />
  );
};
