import { PublicPortalShell } from "./portal-shell";
import { RoadmapBoard } from "./roadmap-board";
import type { PublicPortal, PublicPortalViewer } from "./types";

export const PublicPortalRoadmapPage = ({
  portal,
  viewer,
}: {
  portal: PublicPortal;
  viewer?: PublicPortalViewer | null;
}) => (
  <PublicPortalShell activeTab="roadmap" portal={portal} viewer={viewer}>
    <RoadmapBoard key={portal.slug} portal={portal} />
  </PublicPortalShell>
);
