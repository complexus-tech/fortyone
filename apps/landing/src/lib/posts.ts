import fs from "node:fs";
import path from "node:path";
import matter from "gray-matter";

const postsDir = path.join(process.cwd(), "src/content/blog");

export type PostMetadata = {
  title: string;
  description: string;
  date: string;
  featuredImage: string;
};

export function getAllPosts() {
  const filenames = fs.readdirSync(postsDir);
  return filenames
    .map((name) => {
      const filePath = path.join(postsDir, name);
      const source = fs.readFileSync(filePath, "utf8");
      const { data: metadata } = matter(source);

      return {
        slug: name.replace(/\.mdx$/, ""),
        metadata: metadata as PostMetadata,
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
  return { slug, metadata: metadata as PostMetadata, content };
}
