import type { Metadata } from "next";
import { getAllPosts } from "@/lib/posts";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
import { BlogFeed } from "./blog-feed";
import { BlogJsonLd } from "./json-ld";

export const metadata: Metadata = {
  title: "Project Management Resources & Guides | FortyOne Blog",
  description:
    "Practical guides for connecting goals, customer feedback, planning, and daily project work with clearer ownership and delivery decisions.",
  keywords: [
    "project management blog",
    "OKR guides",
    "team collaboration tips",
    "project planning resources",
    "agile project management",
    "strategy execution",
    "team productivity tips",
    "project management best practices",
  ],
  alternates: {
    canonical: getCanonicalUrl("/blog"),
  },
  openGraph: {
    title: "Project Management Resources & Guides | FortyOne Blog",
    description:
      "Practical guides for connecting goals, customer feedback, planning, and daily project work.",
    url: getCanonicalUrl("/blog"),
    siteName: "FortyOne",
    type: "website",
    images: [DEFAULT_SOCIAL_IMAGE],
  },
  twitter: {
    card: "summary_large_image",
    title: "Project Management Resources & Guides | FortyOne Blog",
    description:
      "Practical guides for connecting goals, customer feedback, planning, and daily project work.",
    images: [DEFAULT_TWITTER_IMAGE],
  },
};

export default function Page() {
  return (
    <main className="bg-background text-foreground pt-20 pb-24 md:pb-32">
      <BlogJsonLd />
      <BlogFeed posts={getAllPosts()}>
        <header className="mx-auto max-w-3xl pt-14 pb-12 text-center md:pt-24 md:pb-16">
          <h1 className="text-5xl font-semibold text-balance md:text-6xl">
            Better decisions. Stronger teams.
          </h1>
          <p className="text-text-description mx-auto mt-7 max-w-2xl text-pretty">
            Practical thinking on strategy, customer feedback, planning, and the
            everyday decisions that turn priorities into progress.
          </p>
        </header>
      </BlogFeed>
    </main>
  );
}
