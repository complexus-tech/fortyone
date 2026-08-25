import type { Metadata } from "next";
import { ArrowLeft2Icon } from "icons";
import Image from "next/image";
import Link from "next/link";
import { MDXRemote } from "next-mdx-remote/rsc";
import { notFound } from "next/navigation";
import { Box, Flex, Text } from "ui";
import { CallToAction } from "@/components/shared";
import { Container } from "@/components/ui";
import { getAllPosts, getPostBySlug } from "@/lib/posts";
import { getCanonicalUrl } from "@/lib/seo";
import { mdxComponents } from "@/mdx-components";
import styles from "./article.module.css";

export function generateStaticParams() {
  return getAllPosts().map((post) => ({ slug: post.slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) {
    return {};
  }

  const canonicalUrl = getCanonicalUrl(`/blog/${slug}`);
  return {
    title: post.metadata.title,
    description: post.metadata.description,
    alternates: {
      canonical: canonicalUrl,
    },
    openGraph: {
      type: "article",
      url: canonicalUrl,
      title: post.metadata.title,
      description: post.metadata.description,
      images: [post.metadata.featuredImage],
      publishedTime: post.metadata.date,
      siteName: "FortyOne",
    },
    twitter: {
      card: "summary_large_image",
      title: post.metadata.title,
      description: post.metadata.description,
      images: [post.metadata.featuredImage],
    },
  };
}

const dateFormatter = new Intl.DateTimeFormat("en-US", {
  month: "long",
  day: "numeric",
  year: "numeric",
});

export default async function BlogPost({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = getPostBySlug(slug);
  if (!post) {
    return notFound();
  }

  const canonicalUrl = getCanonicalUrl(`/blog/${slug}`);
  const featuredImageUrl = getCanonicalUrl(post.metadata.featuredImage);
  const relatedPosts = getAllPosts()
    .filter((candidate) => candidate.slug !== post.slug)
    .slice(0, 2);
  const articleJsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: post.metadata.title,
    description: post.metadata.description,
    image: [featuredImageUrl],
    datePublished: post.metadata.date,
    dateModified: post.metadata.date,
    mainEntityOfPage: canonicalUrl,
    author: {
      "@type": "Organization",
      name: "FortyOne",
    },
    publisher: {
      "@type": "Organization",
      name: "FortyOne",
      logo: {
        "@type": "ImageObject",
        url: "https://www.fortyone.app/images/logo.png",
      },
    },
  };

  return (
    <>
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(articleJsonLd) }}
        type="application/ld+json"
      />
      <article className="bg-background text-foreground pt-28 pb-24 md:pt-36 md:pb-32">
        <Container>
          <Link
            className="text-text-muted hover:text-foreground inline-flex items-center gap-1.5 text-sm transition-colors"
            href="/blog"
          >
            <ArrowLeft2Icon className="size-4" />
            All posts
          </Link>

          <header className="mt-12 max-w-5xl">
            <Flex align="center" className="flex-wrap gap-2 text-sm">
              <Text color="muted">{post.metadata.category}</Text>
              <span aria-hidden="true">·</span>
              <Text color="muted">
                {dateFormatter.format(new Date(post.metadata.date))}
              </Text>
              <span aria-hidden="true">·</span>
              <Text color="muted">{post.readingTime} min read</Text>
            </Flex>
            <Text
              as="h1"
              className="mt-6 max-w-6xl text-[clamp(2.75rem,6vw,5rem)] leading-[0.98] font-semibold tracking-[-0.045em] text-balance"
            >
              {post.metadata.title}
            </Text>
            <Text className="text-text-description mt-7 max-w-2xl text-base leading-7 text-pretty md:text-lg">
              {post.metadata.description}
            </Text>
            <Text className="text-text-muted mt-6 text-sm">
              By {post.metadata.author}
            </Text>
          </header>

          <Box className="bg-surface-muted relative mt-12 aspect-[16/9] overflow-hidden rounded-[2rem] md:mt-16 md:rounded-[3rem]">
            <Image
              alt=""
              className="object-cover"
              fill
              priority
              sizes="(max-width: 1279px) 100vw, 1200px"
              src={post.metadata.featuredImage}
            />
            <Box className="absolute inset-0 bg-gradient-to-t from-black/15 to-transparent" />
          </Box>

          <Box
            className={`${styles.articleBody} mx-auto mt-14 max-w-[720px] md:mt-20`}
          >
            <MDXRemote components={mdxComponents} source={post.content} />
          </Box>

          {relatedPosts.length > 0 ? (
            <Box className="border-border mt-20 border-t pt-10 md:mt-28 md:pt-12">
              <Text as="h2" className="text-3xl font-semibold">
                Keep reading
              </Text>
              <Box className="mt-7 grid gap-5 md:grid-cols-2">
                {relatedPosts.map((relatedPost) => (
                  <Link
                    className="group bg-surface-elevated border-border grid grid-cols-[7rem_minmax(0,1fr)] overflow-hidden rounded-2xl border transition-transform duration-200 hover:-translate-y-0.5 motion-reduce:transition-none motion-reduce:hover:translate-y-0"
                    href={`/blog/${relatedPost.slug}`}
                    key={relatedPost.slug}
                  >
                    <Box className="bg-surface-muted relative min-h-32">
                      <Image
                        alt=""
                        className="object-cover"
                        fill
                        sizes="112px"
                        src={relatedPost.metadata.featuredImage}
                      />
                    </Box>
                    <Box className="p-4">
                      <Text className="text-text-muted text-xs">
                        {relatedPost.metadata.category} ·{" "}
                        {relatedPost.readingTime}m
                      </Text>
                      <Text as="h3" className="mt-2 font-semibold text-pretty">
                        {relatedPost.metadata.title}
                      </Text>
                    </Box>
                  </Link>
                ))}
              </Box>
            </Box>
          ) : null}
        </Container>
      </article>
      <CallToAction />
    </>
  );
}
