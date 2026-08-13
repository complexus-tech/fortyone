import { PublicPortalGuestPreferencesPage } from "@/modules/public-portal";
import { getFeedbackPreferencesAction } from "@/modules/public-portal/actions";
import { getPublicPortalOrNotFound } from "@/modules/public-portal/query";
import { getPublicPortalParticipant } from "@/modules/public-portal/viewer";

export default async function FeedbackPreferencesPage({
  params,
}: {
  params: Promise<{ portalSlug: string }>;
}) {
  const { portalSlug } = await params;
  const [portal, participant, preferencesResponse] = await Promise.all([
    getPublicPortalOrNotFound(portalSlug, {}, { revalidateSeconds: 300 }),
    getPublicPortalParticipant(portalSlug),
    getFeedbackPreferencesAction(portalSlug),
  ]);

  return (
    <PublicPortalGuestPreferencesPage
      initialPreferences={preferencesResponse.data ?? null}
      participant={participant}
      portal={portal}
    />
  );
}
