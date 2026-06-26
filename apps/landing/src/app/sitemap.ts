import type { MetadataRoute } from "next";
import { comparisons } from "@/lib/comparisons";
import { features } from "@/lib/features";
import { useCases } from "@/lib/use-cases";

const useCaseRoutes: MetadataRoute.Sitemap = useCases.map(({ slug }) => ({
  url: `https://www.fortyone.app/use-cases/${slug}`,
  lastModified: new Date(),
  changeFrequency: "monthly",
  priority: 0.8,
}));

const featureRoutes: MetadataRoute.Sitemap = features.map(({ slug }) => ({
  url: `https://www.fortyone.app/features/${slug}`,
  lastModified: new Date(),
  changeFrequency: "monthly",
  priority: 0.8,
}));

const comparisonRoutes: MetadataRoute.Sitemap = comparisons.map(({ slug }) => ({
  url: `https://www.fortyone.app/compare/${slug}`,
  lastModified: new Date(),
  changeFrequency: "monthly",
  priority: 0.7,
}));

// List all static routes with their priorities and change frequencies
const routes: MetadataRoute.Sitemap = [
  {
    url: "https://www.fortyone.app/",
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 1,
  },
  {
    url: "https://www.fortyone.app/ai-project-manager",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.9,
  },
  {
    url: "https://www.fortyone.app/compare",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/pricing",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/blog",
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/contact",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.5,
  },
  {
    url: "https://www.fortyone.app/terms",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.3,
  },
  {
    url: "https://www.fortyone.app/privacy",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.3,
  },
];

export default function sitemap(): MetadataRoute.Sitemap {
  // Combine static and dynamic routes
  return [...routes, ...featureRoutes, ...useCaseRoutes, ...comparisonRoutes];
}
