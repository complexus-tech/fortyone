const SLACK_CRAWLER_PATTERN = /\bSlackbot(?:-LinkExpanding)?\b/i;
const STORY_PATH_PATTERN = /^\/(?:[^/]+\/)?story\/[^/]+(?:\/[^/]+)?\/?$/i;
const WORK_PATH_PATTERN = /^\/(?:[^/]+\/)?work\/[^/]+\/?$/i;

export const isStoryPath = (pathname: string) =>
  STORY_PATH_PATTERN.test(pathname) || WORK_PATH_PATTERN.test(pathname);

export const isSlackStoryLinkPreview = (
  pathname: string,
  userAgent: string | null,
) =>
  Boolean(userAgent && SLACK_CRAWLER_PATTERN.test(userAgent)) &&
  isStoryPath(pathname);

export const buildStoryLinkPreviewHtml = (
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
