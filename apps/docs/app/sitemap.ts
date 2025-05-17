import type { MetadataRoute } from "next";
import fs from "fs";
import path from "path";

const DOCS_BASE_URL = "https://docs.complexus.tech";
const DOCS_CONTENT_DIR = path.join(process.cwd(), "content/docs");

// Simplified static routes: only the root of the docs.
const staticRoutes: MetadataRoute.Sitemap = [
  {
    url: `${DOCS_BASE_URL}/`,
    lastModified: new Date(),
    changeFrequency: "weekly",
    priority: 1,
  },
];

const getMdxFiles = (dir: string): string[] => {
  let files: string[] = [];
  try {
    const items = fs.readdirSync(dir, { withFileTypes: true });
    for (const item of items) {
      const fullPath = path.join(dir, item.name);
      if (item.isDirectory()) {
        files = files.concat(getMdxFiles(fullPath));
      } else if (item.isFile() && item.name.endsWith(".mdx")) {
        files.push(fullPath);
      }
    }
  } catch (error) {
    // console.error(`Error reading directory ${dir}:`, error);
    // If the directory doesn't exist or isn't readable, return empty array
    // This can happen during certain build phases or if content/docs is not yet created
  }
  return files;
};

const getDocsSitemapEntries = (): MetadataRoute.Sitemap => {
  const mdxFiles = getMdxFiles(DOCS_CONTENT_DIR);
  const entries: MetadataRoute.Sitemap = mdxFiles.map((file) => {
    const stats = fs.statSync(file);
    const relativePath = path.relative(DOCS_CONTENT_DIR, file);

    let slug = relativePath
      .replace(/\\/g, "/") // Normalize path separators for Windows
      .replace(/\//g, "/") // Ensure forward slashes if any mixed separators remain
      .replace(/\.mdx$/, "");

    if (slug.endsWith("/index")) {
      slug = slug.substring(0, slug.length - "/index".length);
    } else if (slug === "index") {
      slug = ""; // Root index.mdx maps to '/'
    }

    // Ensure leading slash for slugs, unless it's the root
    const finalSlug = slug ? `/${slug}` : "";
    const url = `${DOCS_BASE_URL}${finalSlug}`;

    return {
      url,
      lastModified: stats.mtime,
      changeFrequency: "weekly",
      priority: 0.7,
    };
  });
  return entries;
};

export default function sitemap(): MetadataRoute.Sitemap {
  const docEntries = getDocsSitemapEntries();

  // Start with a copy of static routes to avoid modifying the original array
  let allEntries = [...staticRoutes];

  // Get the URL of the static root entry
  const staticRootUrl = staticRoutes.length > 0 ? staticRoutes[0].url : null;

  for (const docEntry of docEntries) {
    // If the current docEntry is for the root, update the existing static root entry if it's newer
    if (docEntry.url === staticRootUrl) {
      if (
        allEntries[0] &&
        docEntry.lastModified &&
        allEntries[0].lastModified &&
        docEntry.lastModified > new Date(allEntries[0].lastModified)
      ) {
        allEntries[0].lastModified = docEntry.lastModified;
      }
    } else {
      // For other pages, just add them
      allEntries.push(docEntry);
    }
  }

  // If DOCS_CONTENT_DIR/index.mdx does not exist, but staticRoutes[0] (root) is present,
  // we should ensure its lastModified is recent or based on some other heuristic if preferred.
  // For now, it defaults to new Date() if not updated by a root index.mdx.

  return allEntries;
}
