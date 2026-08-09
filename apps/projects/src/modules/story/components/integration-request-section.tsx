"use client";

import { useStoryIntegrationRequestLinks } from "@/modules/integration-requests/hooks/use-story-request-links";
import { IntegrationRequestBanner } from "./integration-request-banner";

const Banner = ({ storyId }: { storyId: string }) => {
  const { data: links = [] } = useStoryIntegrationRequestLinks(storyId);
  if (links.length === 0) return null;
  return <IntegrationRequestBanner links={links} />;
};

export const IntegrationRequestSection = { Banner };
