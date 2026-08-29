import { PublicPortalNotFoundState } from "@/modules/public-portal/not-found-state";

export default function PublicFeedbackNotFound() {
  return (
    <PublicPortalNotFoundState
      description="This feedback item may have been removed, or the link may be incorrect."
      title="Feedback not found"
    />
  );
}
