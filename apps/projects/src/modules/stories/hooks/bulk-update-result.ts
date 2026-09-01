import type { BulkStoryUpdateResult } from "../actions/bulk-update-stories";

type FailedBulkStoryUpdateItem = BulkStoryUpdateResult["items"][number] & {
  success: false;
};

const getFailureDescription = (
  failedItems: FailedBulkStoryUpdateItem[],
  failedCount: number,
) => {
  const errorsByMessage = new Map<string, number>();

  for (const item of failedItems) {
    const message = item.error?.trim();
    if (!message) continue;
    errorsByMessage.set(message, (errorsByMessage.get(message) ?? 0) + 1);
  }

  const reasons = Array.from(errorsByMessage, ([message, count]) =>
    count > 1 ? `${message} (${count})` : message,
  );

  if (reasons.length === 0) {
    return `${failedCount} ${failedCount === 1 ? "story" : "stories"} could not be updated.`;
  }

  const visibleReasons = reasons.slice(0, 2).join(" ");
  const remainingReasonCount = reasons.length - 2;

  return remainingReasonCount > 0
    ? `${visibleReasons} ${remainingReasonCount} more error${remainingReasonCount === 1 ? "" : "s"} occurred.`
    : visibleReasons;
};

export class BulkStoryUpdateFailure extends Error {
  readonly failedCount: number;
  readonly failedStoryIds: string[];
  readonly totalCount: number;

  constructor(result: BulkStoryUpdateResult) {
    const failedItems = result.items.filter(
      (item): item is FailedBulkStoryUpdateItem => !item.success,
    );
    const inferredFailedCount = Math.max(
      result.failedCount,
      failedItems.length,
      result.totalCount - result.succeededCount,
      result.partial ? 1 : 0,
    );

    super(getFailureDescription(failedItems, inferredFailedCount));
    this.name = "BulkStoryUpdateFailure";
    this.failedCount = inferredFailedCount;
    this.failedStoryIds = failedItems.map(({ storyId }) => storyId);
    this.totalCount = result.totalCount;
  }
}

export const assertBulkStoryUpdateSucceeded = (
  result: BulkStoryUpdateResult,
) => {
  const hasFailedItem = result.items.some(({ success }) => !success);
  const hasFailure =
    result.partial ||
    result.failedCount > 0 ||
    result.succeededCount < result.totalCount ||
    hasFailedItem;

  if (hasFailure) {
    throw new BulkStoryUpdateFailure(result);
  }

  return result;
};
