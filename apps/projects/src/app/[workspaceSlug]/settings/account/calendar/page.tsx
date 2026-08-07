import type { Metadata } from "next";
import { CalendarIntegrationSettings } from "@/modules/settings/workspace/integrations/calendar";

export const metadata: Metadata = {
  title: "Settings › Calendar",
};

export default function Page() {
  return <CalendarIntegrationSettings />;
}
