import {
  IMPORT_ESTIMATE_VALUES,
  IMPORT_MAX_LINK_TITLE_LENGTH,
  IMPORT_MAX_TASK_LINKS,
  isValidImportDurationMinutes,
  isValidImportEstimateValue,
  normalizeImportLinkUrl,
} from "@/modules/settings/workspace/imports/schema";

const normalizeExplicitEffortNumber = <T extends number>(
  value: unknown,
  isValid: (candidate: unknown) => candidate is T,
  markInvalid: () => void,
): T | null => {
  if (value === undefined || value === null) return null;
  if (isValid(value)) return value;
  markInvalid();
  return null;
};

export const normalizeDecodedTaskEffort = (
  decoded: Record<string, unknown>,
) => {
  if (!Array.isArray(decoded.tasks)) {
    return { decoded, warnings: [] as string[] };
  }

  let invalidEstimateCount = 0;
  let invalidDurationCount = 0;
  let invalidFocusCount = 0;
  let focusWithoutDurationCount = 0;
  let focusExceedsDurationCount = 0;
  const tasks = decoded.tasks.map((task) => {
    if (!task || typeof task !== "object" || Array.isArray(task)) return task;
    const values = task as Record<string, unknown>;
    const rawEstimate = values.estimateValue;
    const rawDuration = values.estimatedDurationMinutes;
    const rawFocus = values.minimumFocusBlockMinutes;
    const estimateValue = normalizeExplicitEffortNumber(
      rawEstimate,
      isValidImportEstimateValue,
      () => {
        invalidEstimateCount += 1;
      },
    );
    const estimatedDurationMinutes = normalizeExplicitEffortNumber(
      rawDuration,
      isValidImportDurationMinutes,
      () => {
        invalidDurationCount += 1;
      },
    );
    let minimumFocusBlockMinutes = normalizeExplicitEffortNumber(
      rawFocus,
      isValidImportDurationMinutes,
      () => {
        invalidFocusCount += 1;
      },
    );
    if (
      minimumFocusBlockMinutes !== null &&
      estimatedDurationMinutes === null
    ) {
      focusWithoutDurationCount += 1;
      minimumFocusBlockMinutes = null;
    } else if (
      minimumFocusBlockMinutes !== null &&
      estimatedDurationMinutes !== null &&
      minimumFocusBlockMinutes > estimatedDurationMinutes
    ) {
      focusExceedsDurationCount += 1;
      minimumFocusBlockMinutes = null;
    }

    return {
      ...values,
      estimateValue,
      estimatedDurationMinutes,
      minimumFocusBlockMinutes,
    };
  });
  const warnings = [
    ...(invalidEstimateCount
      ? [
          `${invalidEstimateCount} invalid or ambiguous task complexity ${invalidEstimateCount === 1 ? "estimate was" : "estimates were"} omitted; FortyOne accepts only explicit values ${IMPORT_ESTIMATE_VALUES.join(", ")}.`,
        ]
      : []),
    ...(invalidDurationCount
      ? [
          `${invalidDurationCount} invalid or ambiguous task estimated ${invalidDurationCount === 1 ? "duration was" : "durations were"} omitted; duration must be an explicit whole number of minutes from 1 to 2400.`,
        ]
      : []),
    ...(invalidFocusCount
      ? [
          `${invalidFocusCount} invalid or ambiguous task minimum focus ${invalidFocusCount === 1 ? "block was" : "blocks were"} omitted; focus must be an explicit whole number of minutes from 1 to 2400.`,
        ]
      : []),
    ...(focusWithoutDurationCount
      ? [
          `${focusWithoutDurationCount} task minimum focus ${focusWithoutDurationCount === 1 ? "block was" : "blocks were"} omitted because a valid estimated duration is required.`,
        ]
      : []),
    ...(focusExceedsDurationCount
      ? [
          `${focusExceedsDurationCount} task minimum focus ${focusExceedsDurationCount === 1 ? "block was" : "blocks were"} omitted because it exceeded the estimated duration.`,
        ]
      : []),
  ];

  return { decoded: { ...decoded, tasks }, warnings };
};

export const normalizeDecodedTaskLinks = (decoded: Record<string, unknown>) => {
  if (!Array.isArray(decoded.tasks)) {
    return { decoded, warnings: [] as string[] };
  }

  let invalidLinkCount = 0;
  let duplicateLinkCount = 0;
  let excessLinkCount = 0;
  let adjustedTitleCount = 0;
  const tasks = decoded.tasks.map((task) => {
    if (!task || typeof task !== "object" || Array.isArray(task)) return task;
    const values = task as Record<string, unknown>;
    if (values.links === undefined) return { ...values, links: [] };
    if (!Array.isArray(values.links)) {
      invalidLinkCount += 1;
      return { ...values, links: [] };
    }

    const seenUrls = new Set<string>();
    const links: { title: string | null; url: string }[] = [];
    for (const candidate of values.links) {
      if (
        !candidate ||
        typeof candidate !== "object" ||
        Array.isArray(candidate)
      ) {
        invalidLinkCount += 1;
        continue;
      }
      const link = candidate as Record<string, unknown>;
      if (Object.keys(link).some((key) => key !== "title" && key !== "url")) {
        invalidLinkCount += 1;
        continue;
      }
      const url =
        typeof link.url === "string" ? normalizeImportLinkUrl(link.url) : null;
      if (!url) {
        invalidLinkCount += 1;
        continue;
      }
      if (seenUrls.has(url)) {
        duplicateLinkCount += 1;
        continue;
      }
      seenUrls.add(url);
      if (links.length >= IMPORT_MAX_TASK_LINKS) {
        excessLinkCount += 1;
        continue;
      }

      let title: string | null = null;
      if (typeof link.title === "string") {
        const trimmedTitle = link.title.trim();
        if (trimmedTitle) {
          title = trimmedTitle.slice(0, IMPORT_MAX_LINK_TITLE_LENGTH);
          if (title.length !== trimmedTitle.length) adjustedTitleCount += 1;
        }
      } else if (link.title !== null && link.title !== undefined) {
        adjustedTitleCount += 1;
      }
      links.push({ title, url });
    }

    return { ...values, links };
  });
  const warnings = [
    ...(invalidLinkCount
      ? [
          `${invalidLinkCount} unsafe or malformed task ${invalidLinkCount === 1 ? "link was" : "links were"} omitted; only absolute HTTP or HTTPS URLs are supported.`,
        ]
      : []),
    ...(duplicateLinkCount
      ? [
          `${duplicateLinkCount} duplicate task ${duplicateLinkCount === 1 ? "link was" : "links were"} deduplicated by canonical URL.`,
        ]
      : []),
    ...(excessLinkCount
      ? [
          `${excessLinkCount} task ${excessLinkCount === 1 ? "link was" : "links were"} omitted because a work item can import at most ${IMPORT_MAX_TASK_LINKS} links.`,
        ]
      : []),
    ...(adjustedTitleCount
      ? [
          `${adjustedTitleCount} task link ${adjustedTitleCount === 1 ? "title was" : "titles were"} truncated or omitted to fit FortyOne's link title contract.`,
        ]
      : []),
  ];

  return { decoded: { ...decoded, tasks }, warnings };
};
