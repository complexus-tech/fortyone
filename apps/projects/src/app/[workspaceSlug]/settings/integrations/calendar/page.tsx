import type { Metadata } from "next";
import { CalendarIntegrationSettings } from "@/modules/settings/workspace/integrations/calendar";

export const metadata: Metadata = {
  title: "Settings › Google Calendar",
};

export default function GoogleCalendarIntegrationPage() {
  return <CalendarIntegrationSettings />;
}
