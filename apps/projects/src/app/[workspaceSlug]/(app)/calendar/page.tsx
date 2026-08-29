import type { Metadata } from "next";
import { CalendarPage } from "@/modules/calendar";

export const metadata: Metadata = {
  title: "Calendar",
  description: "Plan work around your connected calendar.",
};

export default function Page() {
  return <CalendarPage />;
}
