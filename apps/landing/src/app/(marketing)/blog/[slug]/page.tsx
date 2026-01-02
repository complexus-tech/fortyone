import { MDXRemote } from "next-mdx-remote/rsc";
import type { Metadata } from "next";
import { notFound } from "next/navigation";
import { BlurImage, Box, Flex, Text } from "ui";
import { ArrowLeft2Icon } from "icons";
import Link from "next/link";
import { getPostBySlug, getAllPosts } from "@/lib/posts";
import { mdxComponents } from "@/mdx-components";
import { CallToAction } from "@/components/shared";
import { Container } from "@/components/ui";

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
  return {
    title: post?.metadata.title,
    description: post?.metadata.description,
    openGraph: {
      images: post?.metadata.featuredImage,
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

  return (
    <>
      <Container
        as="article"
        className="grid grid-cols-[1fr_4fr_1fr] items-start gap-10 py-32"
      >
        <Link className="flex items-center gap-1 pt-4" href="/blog">
          <ArrowLeft2Icon className="dark:text-foreground" />
          <span className="opacity-80">All blogs</span>
        </Link>
        <Box className="mx-auto max-w-3xl">
          <Flex gap={2}>
            <Text className="opacity-80">
              {new Date(post.metadata.date as string).toLocaleDateString(
                "en-US",
                {
                  month: "long",
                  day: "numeric",
                  year: "numeric",
                },
              )}
            </Text>
            &bull;
            <Text className="opacity-80">6 min read</Text>
          </Flex>
          <Text
            as="h1"
            className="mb-8 mt-4 text-5xl font-semibold leading-tight"
          >
            {post.metadata.title}
          </Text>
          <Box className="mb-6 rounded-[0.9rem] border border-border p-1.5 d">
            <BlurImage
              alt={post.metadata.title}
              className="aspect-16/8 rounded-[0.6rem]"
              src={post.metadata.featuredImage}
            />
          </Box>
          <Box className="prose prose-lg prose-stone max-w-full dark:prose-invert prose-headings:font-semibold prose-a:text-primary prose-pre:text-[1.1rem] prose-pre:bg-surface-muted prose-pre:text-foreground">
            <MDXRemote components={mdxComponents} source={post.content} />
          </Box>
        </Box>
      </Container>
      <CallToAction />
    </>
  );
}
