import type {
  IntegrationRequestComment,
  IntegrationRequestThreadActivity,
} from "./types";

export const getCommentDeliveryLabel = (
  status: IntegrationRequestComment["deliveryStatus"],
) => {
  switch (status) {
    case "sent":
      return "Sent";
    case "sending":
      return "Sending";
    case "retrying":
      return "Retrying";
    case "failed":
      return "Failed";
    case "not-sent":
      return "Not sent";
    default:
      return null;
  }
};

export const shouldPollRequestThread = (
  activity: IntegrationRequestThreadActivity | undefined,
) =>
  activity?.comments.some(
    (comment) =>
      comment.direction === "outbound" &&
      (comment.deliveryStatus === "sending" ||
        comment.deliveryStatus === "retrying"),
  ) ?? false;
