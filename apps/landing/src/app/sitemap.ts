import type { MetadataRoute } from "next";

// List all static routes with their priorities and change frequencies
const routes: MetadataRoute.Sitemap = [
  {
    url: "https://www.fortyone.app/",
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 1,
  },
  {
    url: "https://www.fortyone.app/product/stories",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/product/objectives",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/product/okrs",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.fortyone.app/product/sprints",
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
    url: "https://www.fortyone.app/developers",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.7,
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
  return [...routes];
}
