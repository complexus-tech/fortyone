import type { MetadataRoute } from "next";

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      {
        userAgent: "*",
        allow: ["/"],
        disallow: ["/api/*", "/auth-callback/*", "/onboarding/*", "/verify/*"],
      },
    ],
    sitemap: "https://www.fortyone.app/sitemap.xml",
  };
}
