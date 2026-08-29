import { FeedbackSettings } from "@/modules/settings/workspace/feedback";

export const metadata = {
  title: "Settings › Feedback",
};

export default function Page() {
  const anonymousFeedbackAvailable =
    process.env.NODE_ENV !== "production" ||
    Boolean(process.env.VERCEL) ||
    Boolean(process.env.FEEDBACK_TRUSTED_CLIENT_IP_HEADER?.trim());

  return (
    <FeedbackSettings anonymousFeedbackAvailable={anonymousFeedbackAvailable} />
  );
}
