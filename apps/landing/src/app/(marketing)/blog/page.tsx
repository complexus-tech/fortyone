import type { Metadata } from "next";
import Link from "next/link";
import Image from "next/image";
import { getAllPosts } from "@/lib/posts";
import { BlogJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Project Management Resources & Guides | Complexus Blog",
  description:
    "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  keywords: [
    "project management blog",
    "OKR guides",
    "team collaboration tips",
    "project management resources",
    "agile project management",
    "OKR implementation",
    "team productivity tips",
    "project management best practices",
  ],
  openGraph: {
    title: "Project Management Resources & Guides | Complexus Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
  twitter: {
    title: "Project Management Resources & Guides | Complexus Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
};

export default function Page() {
  const posts = getAllPosts();
  return (
    <>
      <BlogJsonLd />
      <h1>Blog</h1>
      <ul>
        {posts.map(
          ({ slug, metadata: { title, description, featuredImage } }) => (
            <li key={slug}>
              <Link href={`/blog/${slug}`}>
                <Image
                  alt={title}
                  height={250}
                  src={featuredImage}
                  width={400}
                />
                <h2>{title}</h2>
                <p>{description}</p>
              </Link>
            </li>
          ),
        )}
      </ul>
    </>
  );
}
