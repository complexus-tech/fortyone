import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";

const postsDir = path.join(process.cwd(), "src/content/blog");

export type PostMetadata = {
  author: string;
  category: string;
  title: string;
  description: string;
  date: string;
  featuredImage: string;
};

export const getReadingTime = (content: string) =>
  Math.max(1, Math.ceil(content.trim().split(/\s+/).length / 220));

export function getAllPosts() {
  const filenames = fs.readdirSync(postsDir);
  return filenames
    .map((name) => {
      const filePath = path.join(postsDir, name);
      const source = fs.readFileSync(filePath, "utf8");
      const { content, data: metadata } = matter(source);

      return {
        slug: name.replace(/\.mdx$/, ""),
        metadata: metadata as PostMetadata,
        readingTime: getReadingTime(content),
      };
    })
    .sort(
      (a, b) =>
        new Date(b.metadata.date).getTime() -
        new Date(a.metadata.date).getTime(),
    );
}

export function getPostBySlug(slug: string) {
  const fullPath = path.join(postsDir, `${slug}.mdx`);
  const source = fs.readFileSync(fullPath, "utf8");
  const { data: metadata, content } = matter(source);
  return {
    slug,
    metadata: metadata as PostMetadata,
    content,
    readingTime: getReadingTime(content),
  };
}
