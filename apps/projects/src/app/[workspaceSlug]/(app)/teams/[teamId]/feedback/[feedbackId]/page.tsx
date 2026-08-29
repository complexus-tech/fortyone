import type { Metadata } from "next";
import { TeamFeedbackDetails } from "@/modules/team-feedback/details";

export const metadata: Metadata = {
  title: "Feedback",
};

export default async function Page({
  params,
}: {
  params: Promise<{ feedbackId: string }>;
}) {
  const { feedbackId } = await params;
  return <TeamFeedbackDetails feedbackId={feedbackId} />;
}
