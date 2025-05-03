import type { MetadataRoute } from "next";

// List all static routes with their priorities and change frequencies
const routes: MetadataRoute.Sitemap = [
  {
    url: "https://www.complexus.tech/",
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 1,
  },
  {
    url: "https://www.complexus.tech/product/stories",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/product/objectives",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/product/okrs",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/product/sprints",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/pricing",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/blog",
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 0.8,
  },
  {
    url: "https://www.complexus.tech/developers",
    lastModified: new Date(),
    changeFrequency: "monthly",
    priority: 0.7,
  },
  {
    url: "https://www.complexus.tech/contact",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.5,
  },
  {
    url: "https://www.complexus.tech/terms",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.3,
  },
  {
    url: "https://www.complexus.tech/privacy",
    lastModified: new Date(),
    changeFrequency: "yearly",
    priority: 0.3,
  },
];

export default function sitemap(): MetadataRoute.Sitemap {
  // Combine static and dynamic routes
  return [...routes];
}
