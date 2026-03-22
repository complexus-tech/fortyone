"use client";

import { useStoryGitHubLinks } from "@/lib/hooks/github";
import { GitHubBanner } from "./github-banner";

/**
 * Wrapper component that owns the GitHub links query independently
 * from the parent MainDetails component. This prevents the query's
 * loading→success state transitions from triggering the tiptap
 * editor useEffect in the parent (which can cause infinite update loops).
 */

const Banner = ({ storyId }: { storyId: string }) => {
  const { data: links = [] } = useStoryGitHubLinks(storyId);
  if (links.length === 0) return null;
  return <GitHubBanner links={links} storyId={storyId} />;
};

export const GitHubSection = { Banner };
