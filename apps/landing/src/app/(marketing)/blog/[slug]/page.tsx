import type { Metadata } from "next";
import { ArrowLeft2Icon } from "icons";
import Image from "next/image";
import Link from "next/link";
import { compileMDX } from "next-mdx-remote/rsc";
import { notFound } from "next/navigation";
import { CallToAction } from "@/components/shared";
import { Container } from "@/components/ui";
import { getAllPosts, getPostBySlug } from "@/lib/posts";
import { getCanonicalUrl } from "@/lib/seo";
import { PostCard, PostMeta } from "../post-card";
import { articleMdxComponents } from "./article-components";
import { ArticleMap } from "./article-map";
import { createArticleMarkdown } from "./article-markdown";
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

  const markdown = createArticleMarkdown();
  const { content } = await compileMDX({
    source: post.content,
    components: articleMdxComponents,
    options: markdown.options,
  });
  const canonicalUrl = getCanonicalUrl(`/blog/${slug}`);
  const featuredImageUrl = getCanonicalUrl(post.metadata.featuredImage);
  const relatedPosts = getAllPosts()
    .filter((candidate) => candidate.slug !== post.slug)
    .slice(0, 3);
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
      <article
        className="bg-background text-foreground pt-24 pb-24 md:pb-32"
      >
        <Container>
          <header className={styles.articleHeader} id="article-top">
            <Link
              className="text-text-muted hover:text-foreground focus-visible:outline-ring inline-flex items-center gap-1.5 rounded-md text-sm transition-colors focus-visible:outline-2"
              href="/blog"
            >
              <ArrowLeft2Icon className="size-4" />
              All posts
            </Link>
            <div className="mt-8 flex justify-center">
              <PostMeta appearance="plain" post={post} />
            </div>
            <h1 className={styles.articleTitle}>{post.metadata.title}</h1>
            <p className="text-text-description mx-auto mt-6 max-w-2xl text-pretty">
              {post.metadata.description}
            </p>
            <p className="text-text-muted mt-7 text-sm">
              {post.metadata.author}
              <span aria-hidden="true"> · </span>
              {post.readingTime} min read
            </p>
          </header>
          <div
            className={`${styles.articleCover} bg-surface-muted relative overflow-hidden rounded-2xl sm:rounded-[2rem]`}
          >
            <Image
              alt=""
              className="object-cover"
              fill
              priority
              sizes="(max-width: 1023px) 100vw, 960px"
              src={post.metadata.featuredImage}
            />
          </div>
          <div className={styles.readingLayout}>
            <ArticleMap headings={markdown.headings} />
            <div className={styles.articleBody}>{content}</div>
          </div>

          {relatedPosts.length > 0 ? (
            <section
              aria-labelledby="keep-reading-title"
              className="border-border mt-20 border-t pt-10 md:mt-28 md:pt-12"
            >
              <div className="flex items-center justify-between gap-4">
                <h2 className="text-3xl font-semibold" id="keep-reading-title">
                  Keep reading
                </h2>
                <Link
                  className="text-text-muted hover:text-foreground text-sm underline underline-offset-4"
                  href="/blog"
                >
                  All posts
                </Link>
              </div>
              <div className="mt-8 grid gap-x-8 gap-y-12 sm:grid-cols-2 lg:grid-cols-3">
                {relatedPosts.map((relatedPost) => (
                  <PostCard key={relatedPost.slug} post={relatedPost} />
                ))}
              </div>
            </section>
          ) : null}
        </Container>
      </article>
      <CallToAction />
    </>
  );
}
