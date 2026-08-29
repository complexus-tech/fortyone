import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { ArrowRightIcon } from "icons";
import { cn } from "lib";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { getAllPosts } from "@/lib/posts";
import {
  DEFAULT_SOCIAL_IMAGE,
  DEFAULT_TWITTER_IMAGE,
  getCanonicalUrl,
} from "@/lib/seo";
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

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  day: "numeric",
  month: "short",
  year: "numeric",
});

type BlogPostSummary = ReturnType<typeof getAllPosts>[number];

function PostMeta({ post }: { post: BlogPostSummary }) {
  return (
    <Text className="text-text-muted text-sm">
      <time dateTime={post.metadata.date}>
        {dateFormatter.format(new Date(post.metadata.date))}
      </time>
      <span aria-hidden="true"> · </span>
      <span>{post.metadata.category}</span>
    </Text>
  );
}

function FeaturedPostCard({
  className,
  large = false,
  post,
  priority = false,
}: {
  className?: string;
  large?: boolean;
  post: BlogPostSummary;
  priority?: boolean;
}) {
  return (
    <article className={className}>
      <Link
        className="group bg-surface-elevated border-border flex h-full flex-col overflow-hidden rounded-[1.75rem] border transition-transform duration-300 ease-out hover:-translate-y-1 focus-visible:ring-2 focus-visible:ring-inset motion-reduce:transition-none motion-reduce:hover:translate-y-0"
        href={`/blog/${post.slug}`}
      >
        <Box
          className={cn(
            "bg-surface-muted relative overflow-hidden",
            large ? "aspect-[16/10]" : "aspect-[16/8.4]",
          )}
        >
          <Image
            alt=""
            className="object-cover transition-transform duration-500 ease-out group-hover:scale-[1.02] motion-reduce:transition-none motion-reduce:group-hover:scale-100"
            fill
            priority={priority}
            sizes={
              large
                ? "(max-width: 1023px) 100vw, 66vw"
                : "(max-width: 1023px) 100vw, 34vw"
            }
            src={post.metadata.featuredImage}
          />
          <Box className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent" />
        </Box>
        <Box
          className={cn("flex flex-1 flex-col", large ? "p-6 md:p-8" : "p-5")}
        >
          <PostMeta post={post} />
          <Text
            as="h2"
            className={cn(
              "mt-3 font-semibold text-pretty",
              large ? "text-3xl md:text-4xl" : "text-xl md:text-2xl",
            )}
          >
            {post.metadata.title}
          </Text>
          <Text
            className={cn(
              "text-text-description mt-3 line-clamp-2 leading-relaxed",
              large ? "max-w-2xl text-base" : "text-sm",
            )}
          >
            {post.metadata.description}
          </Text>
          <Text className="text-text-muted mt-auto pt-6 text-sm">
            {post.metadata.author}
            <span aria-hidden="true"> · </span>
            {post.readingTime} min read
          </Text>
        </Box>
      </Link>
    </article>
  );
}

export default function Page() {
  const posts = getAllPosts();
  const [leadPost, ...secondaryPosts] = posts;

  return (
    <main className="bg-background text-foreground pt-28 pb-24 md:pt-36 md:pb-32">
      <BlogJsonLd />
      <Container>
        <Box className="max-w-3xl" data-landing-reveal>
          <Text className="text-text-muted text-sm font-semibold tracking-wide uppercase">
            FortyOne blog
          </Text>
          <Text
            as="h1"
            className="mt-5 text-5xl font-semibold text-balance md:text-7xl"
          >
            Ideas for keeping important work moving.
          </Text>
          <Text className="text-text-description mt-6 max-w-2xl text-base leading-relaxed text-pretty md:text-lg">
            Practical thinking on strategy, customer feedback, planning, and the
            everyday decisions that turn priorities into progress.
          </Text>
        </Box>

        {leadPost ? (
          <Box className="mt-14 grid gap-5 lg:grid-cols-[minmax(0,1.6fr)_minmax(20rem,0.8fr)] lg:grid-rows-2">
            <FeaturedPostCard
              className="lg:row-span-2"
              large
              post={leadPost}
              priority
            />
            {secondaryPosts.slice(0, 2).map((post) => (
              <FeaturedPostCard key={post.slug} post={post} />
            ))}
          </Box>
        ) : null}

        <Box className="mt-24 md:mt-32">
          <Box className="border-border flex items-end justify-between gap-6 border-b pb-5">
            <Box>
              <Text as="h2" className="text-3xl font-semibold md:text-4xl">
                All posts
              </Text>
              <Text className="text-text-muted mt-2 text-sm">
                Guides for clearer planning and connected delivery.
              </Text>
            </Box>
            <Text className="text-text-muted hidden text-sm sm:block">
              {posts.length} articles
            </Text>
          </Box>

          <Box>
            {posts.map((post) => (
              <Link
                className="group border-border hover:bg-surface-muted/50 grid gap-3 border-b py-5 transition-colors focus-visible:ring-2 focus-visible:ring-inset md:grid-cols-[10.5rem_minmax(0,1fr)_9rem_4rem] md:items-center md:gap-6 md:px-3"
                href={`/blog/${post.slug}`}
                key={post.slug}
              >
                <PostMeta post={post} />
                <Text as="h3" className="font-semibold text-pretty md:text-lg">
                  {post.metadata.title}
                </Text>
                <Text className="text-text-muted text-sm md:text-right">
                  {post.metadata.author}
                </Text>
                <Box className="flex items-center justify-between gap-3 md:justify-end">
                  <Text className="text-text-muted text-sm tabular-nums">
                    {post.readingTime}m
                  </Text>
                  <ArrowRightIcon className="size-4 transition-transform duration-200 group-hover:translate-x-1 motion-reduce:transition-none" />
                </Box>
              </Link>
            ))}
          </Box>
        </Box>
      </Container>
    </main>
  );
}
