import { MDXRemote } from "next-mdx-remote/rsc";
import Image from "next/image";
import type { Metadata } from "next";
import { getPostBySlug, getAllPosts } from "@/lib/posts";
import { mdxComponents } from "@/mdx-components";

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

  return (
    <article>
      <h1>{post?.metadata.title}</h1>
      <Image
        alt={post?.metadata.title}
        height={450}
        src={post?.metadata.featuredImage}
        width={800}
      />
      <MDXRemote components={mdxComponents} source={post?.content} />
    </article>
  );
}
