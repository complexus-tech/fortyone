import type { MetadataRoute } from "next";
import { source } from "@/lib/source";

const DOCS_BASE_URL = "https://docs.fortyone.app";

export default function sitemap(): MetadataRoute.Sitemap {
  return source.getPages().map((page) => ({
    url: new URL(page.url, DOCS_BASE_URL).toString(),
    changeFrequency: "weekly",
    priority: page.url === "/" ? 1 : 0.7,
  }));
}
