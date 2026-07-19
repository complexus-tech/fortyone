import { PublicPortalNotFoundState } from "@/modules/public-portal/not-found-state";

export default function PublicPortalNotFound() {
  return (
    <PublicPortalNotFoundState
      description="This organization is not currently configured to collect public feedback."
      title="Feedback isn't available"
    />
  );
}
