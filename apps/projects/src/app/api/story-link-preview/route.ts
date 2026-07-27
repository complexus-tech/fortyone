import { buildStoryLinkPreviewHtml } from "@/story-link-preview";

export const GET = (request: Request) => {
  const faviconUrl = new URL("/favicon.ico", request.url).toString();

  return new Response(buildStoryLinkPreviewHtml(faviconUrl), {
    headers: {
      "Cache-Control": "public, max-age=3600, stale-while-revalidate=86400",
      "Content-Type": "text/html; charset=utf-8",
      "X-Robots-Tag": "noindex, nofollow, noarchive",
    },
  });
};
