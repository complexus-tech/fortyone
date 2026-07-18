import type { Metadata } from "next";
import { ArrowLeft2Icon } from "icons";
import Link from "next/link";
import { MDXRemote } from "next-mdx-remote/rsc";
import { notFound } from "next/navigation";
import { Box, Flex, Text } from "ui";
import { CallToAction } from "@/components/shared";
import { Container } from "@/components/ui";
import { getAllPosts, getPostBySlug } from "@/lib/posts";
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

  const canonicalUrl = `https://www.fortyone.app/blog/${slug}`;
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

const getReadingTime = (content: string) =>
  Math.max(1, Math.ceil(content.trim().split(/\s+/).length / 220));

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

  const canonicalUrl = `https://www.fortyone.app/blog/${slug}`;
  const articleJsonLd = {
    "@context": "https://schema.org",
    "@type": "Article",
    headline: post.metadata.title,
    description: post.metadata.description,
    image: [post.metadata.featuredImage],
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
      <Container
        as="article"
        className="w-[calc(100%-2.5rem)] max-w-[640px] px-0 pt-32 pb-24 md:px-0 md:pt-40"
        full
      >
        <Link
          className="text-text-muted hover:text-foreground mb-12 inline-flex items-center gap-1.5 text-sm transition-colors"
          href="/blog"
        >
          <ArrowLeft2Icon className="size-4" />
          All posts
        </Link>

        <header className="border-border mb-12 border-b pb-10 md:mb-14 md:pb-12">
          <Flex align="center" className="gap-2 text-sm">
            <Text color="muted">
              {dateFormatter.format(new Date(post.metadata.date as string))}
            </Text>
            <span aria-hidden="true">·</span>
            <Text color="muted">{getReadingTime(post.content)} min read</Text>
          </Flex>
          <Text
            as="h1"
            className={`${styles.articleTitle} mt-5 text-[clamp(2rem,5vw,2.75rem)] leading-[1.08] font-medium tracking-[-0.025em]`}
          >
            {post.metadata.title}
          </Text>
          <Text className="mt-5 max-w-2xl text-base leading-7" color="muted">
            {post.metadata.description}
          </Text>
        </header>

        <Box className={styles.articleBody}>
          <MDXRemote components={mdxComponents} source={post.content} />
        </Box>
      </Container>
      <CallToAction />
    </>
  );
}
