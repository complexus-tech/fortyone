import type { Metadata } from "next";
import { SelectTeamFeedbackMessage } from "@/modules/team-feedback/message";

export const metadata: Metadata = {
  title: "Feedback",
};

export default function Page() {
  return <SelectTeamFeedbackMessage />;
}
