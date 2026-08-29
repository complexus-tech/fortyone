"use client";

import { useStoryFeedbackLinks } from "@/modules/team-feedback/hooks/use-story-feedback-links";
import { FeedbackBanner } from "./feedback-banner";

const Banner = ({ storyId }: { storyId: string }) => {
  const { data: links = [] } = useStoryFeedbackLinks(storyId);

  if (links.length === 0) return null;

  return <FeedbackBanner links={links} />;
};

export const FeedbackSection = { Banner };
