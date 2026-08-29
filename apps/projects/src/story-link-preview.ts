const SLACK_CRAWLER_PATTERN = /\bSlackbot(?:-LinkExpanding)?\b/i;
const STORY_PATH_PATTERN = /^\/(?:[^/]+\/)?story\/[^/]+(?:\/[^/]+)?\/?$/i;
const WORK_PATH_PATTERN = /^\/(?:[^/]+\/)?work\/[^/]+\/?$/i;
const UUID_PATH_SEGMENT =
  "[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}";
const REQUEST_PATH_PATTERN = new RegExp(
  `^/(?:[^/]+/)?teams/${UUID_PATH_SEGMENT}/requests/${UUID_PATH_SEGMENT}/?$`,
  "i",
);

export const isMinimalLinkPreviewPath = (pathname: string) =>
  STORY_PATH_PATTERN.test(pathname) ||
  WORK_PATH_PATTERN.test(pathname) ||
  REQUEST_PATH_PATTERN.test(pathname);

export const isSlackMinimalLinkPreview = (
  pathname: string,
  userAgent: string | null,
) =>
  Boolean(userAgent && SLACK_CRAWLER_PATTERN.test(userAgent)) &&
  isMinimalLinkPreviewPath(pathname);

export const buildMinimalLinkPreviewHtml = (
  faviconUrl: string,
) => `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="robots" content="noindex, nofollow, noarchive">
    <link rel="icon" href="${faviconUrl}" sizes="16x16 32x32" type="image/x-icon">
  </head>
  <body></body>
</html>`;
