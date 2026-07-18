import type { Metadata } from "next";
import Link from "next/link";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { getAllPosts } from "@/lib/posts";
import { BlogJsonLd } from "./json-ld";
import styles from "./blog-list.module.css";

export const metadata: Metadata = {
  title: "Project Management Resources & Guides | FortyOne Blog",
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
    title: "Project Management Resources & Guides | FortyOne Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
  twitter: {
    title: "Project Management Resources & Guides | FortyOne Blog",
    description:
      "Access expert insights, best practices, and guides on OKR implementation, project management, and team collaboration strategies.",
  },
};

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "short",
  year: "numeric",
});

const ExternalMark = () => (
  <svg
    aria-hidden="true"
    className="text-text-muted size-[13px] fill-none stroke-current stroke-[1.25] transition-transform duration-200 [stroke-linecap:round] [stroke-linejoin:round] group-hover:translate-x-0.5 group-hover:-translate-y-0.5 motion-reduce:transition-none motion-reduce:group-hover:translate-0"
    viewBox="0 0 16 16"
  >
    <path d="M4 12 12 4M6 4h6v6" />
  </svg>
);

export default function Page() {
  const posts = getAllPosts();

  return (
    <Container
      className="w-[calc(100%-2.5rem)] max-w-[640px] px-0 pt-32 pb-32 md:px-0 md:pt-40"
      full
    >
      <BlogJsonLd />
      <Text
        as="h1"
        className={`${styles.bodyFont} text-base leading-6 font-medium`}
      >
        Blog
      </Text>

      <Box className="border-border mt-[72px] border-t">
        {posts.map(({ slug, metadata: postMetadata }) => (
          <Link
            className="group border-border grid grid-cols-[minmax(0,1fr)_72px_16px] items-center gap-3 border-b py-4 md:grid-cols-[minmax(0,1fr)_86px_16px] md:gap-5"
            href={`/blog/${slug}`}
            key={slug}
          >
            <Text
              as="h2"
              className={`${styles.bodyFont} min-w-0 text-base font-medium`}
            >
              <span className={styles.writingLinkTitle}>
                {postMetadata.title}
              </span>
            </Text>
            <time
              className="text-text-muted group-hover:text-text-secondary text-right text-[13px] tabular-nums transition-colors duration-200"
              dateTime={postMetadata.date}
            >
              {dateFormatter.format(new Date(postMetadata.date))}
            </time>
            <ExternalMark />
          </Link>
        ))}
      </Box>
    </Container>
  );
}
