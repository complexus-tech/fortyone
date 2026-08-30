import { toast } from "sonner";
import type { PublicRequest, SimilarPublicFeedback } from "../types";

export const MAX_FEEDBACK_TITLE_LENGTH = 200;

export const copyAnonymousFeedbackTrackingLink = async (
  trackingUrl: string,
) => {
  try {
    await navigator.clipboard.writeText(trackingUrl);
    toast.success("Tracking link copied");
  } catch {
    toast.error("Unable to copy tracking link");
  }
};

export const getFeedbackSubmitLabel = ({
  hasDuplicate,
  isSubmitting,
  requiresIdentity,
}: {
  hasDuplicate: boolean;
  isSubmitting: boolean;
  requiresIdentity: boolean;
}) => {
  if (hasDuplicate) return "View existing feedback";
  if (isSubmitting) return "Submitting...";
  if (requiresIdentity) return "Continue";
  return "Submit feedback";
};

export const toSimilarFeedbackRequest = (
  item: SimilarPublicFeedback,
): PublicRequest => ({
  authorAvatar: item.authorAvatar,
  authorId: item.authorId,
  authorName: item.authorName || "Unknown contributor",
  boardId: "",
  commentCount: item.commentCount,
  comments: [],
  createdAtLabel: "",
  description: "",
  id: item.id,
  slug: item.slug,
  status: item.status ?? "pending",
  storyLinks: [],
  title: item.title,
  voteCount: item.voteCount,
});
