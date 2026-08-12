import { FEEDBACK_WIDGET_LOADER_SOURCE } from "@/modules/feedback-widget/loader-source";

export const dynamic = "force-static";

export const GET = () =>
  new Response(FEEDBACK_WIDGET_LOADER_SOURCE, {
    headers: {
      "Cache-Control":
        "public, max-age=300, s-maxage=86400, stale-while-revalidate=604800",
      "Content-Type": "text/javascript; charset=utf-8",
      "X-Content-Type-Options": "nosniff",
    },
  });
