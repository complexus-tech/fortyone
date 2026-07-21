import type { MetadataRoute } from "next";
import { comparisons } from "@/lib/comparisons";
import { features } from "@/lib/features";
import { getAllPosts } from "@/lib/posts";
import { getCanonicalUrl } from "@/lib/seo";
import { useCases } from "@/lib/use-cases";

const useCaseRoutes: MetadataRoute.Sitemap = useCases.map(({ slug }) => ({
  url: getCanonicalUrl(`/use-cases/${slug}`),
}));

const featureRoutes: MetadataRoute.Sitemap = features.map(({ slug }) => ({
  url: getCanonicalUrl(`/features/${slug}`),
}));

const comparisonRoutes: MetadataRoute.Sitemap = comparisons.map(({ slug }) => ({
  url: getCanonicalUrl(`/compare/${slug}`),
}));

const routes: MetadataRoute.Sitemap = [
  { url: getCanonicalUrl("/") },
  { url: getCanonicalUrl("/ai-project-manager") },
  { url: getCanonicalUrl("/pricing") },
  { url: getCanonicalUrl("/blog") },
  { url: getCanonicalUrl("/contact") },
  { url: getCanonicalUrl("/terms") },
  { url: getCanonicalUrl("/privacy") },
];

const blogRoutes: MetadataRoute.Sitemap = getAllPosts().map(
  ({ slug, metadata }) => ({
    url: getCanonicalUrl(`/blog/${slug}`),
    lastModified: new Date(metadata.date),
  }),
);

export default function sitemap(): MetadataRoute.Sitemap {
  return [
    ...routes,
    ...featureRoutes,
    ...useCaseRoutes,
    ...comparisonRoutes,
    ...blogRoutes,
  ];
}
