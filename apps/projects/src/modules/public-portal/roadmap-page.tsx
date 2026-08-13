import { PublicPortalShell } from "./portal-shell";
import { RoadmapBoard } from "./roadmap-board";
import type { PublicPortal, PublicPortalParticipant } from "./types";
import { anonymousPublicPortalParticipant } from "./participant";

export const PublicPortalRoadmapPage = ({
  participant = anonymousPublicPortalParticipant,
  portal,
}: {
  participant?: PublicPortalParticipant;
  portal: PublicPortal;
}) => (
  <PublicPortalShell
    activeTab="roadmap"
    participant={participant}
    portal={portal}
  >
    <RoadmapBoard key={portal.slug} portal={portal} />
  </PublicPortalShell>
);
