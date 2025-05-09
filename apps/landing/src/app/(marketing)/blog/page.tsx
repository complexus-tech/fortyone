import type { Metadata } from "next";
import Link from "next/link";
import { BlurImage, Box, Flex, Text, buttonVariants } from "ui";
import { cn } from "lib";
import { getAllPosts } from "@/lib/posts";
import { Container } from "@/components/ui";
import { CallToAction } from "@/components/shared";
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
  const firstPost = posts[0];
  const remainingPosts = posts.slice(1);
  return (
    <>
      <Container className="max-w-7xl pt-12 md:pt-16">
        <BlogJsonLd />
        <Flex className="mb-8 mt-16 text-center" justify="center">
          <span
            className={cn(
              buttonVariants({
                color: "tertiary",
                rounded: "full",
                size: "sm",
              }),
              "px-3 text-sm md:text-base",
            )}
          >
            Blog
          </span>
        </Flex>
        <Box className="group mb-12 grid grid-cols-1 items-center gap-8 md:grid-cols-[1.5fr_1fr]">
          <Box className="rounded-[0.9rem] border border-dark-50 bg-dark-100/60 p-1.5">
            <BlurImage
              alt={firstPost.metadata.title}
              className="aspect-video rounded-[0.6rem]"
              src={firstPost.metadata.featuredImage}
            />
          </Box>
          <Box>
            <Text className="mb-4 opacity-80">
              {firstPost.metadata.date
                ? new Date(firstPost.metadata.date).toLocaleDateString(
                    "en-US",
                    {
                      month: "long",
                      day: "numeric",
                      year: "numeric",
                    },
                  )
                : null}
            </Text>
            <Text
              as="h3"
              className="mb-4 text-4xl font-semibold group-hover:underline"
            >
              {firstPost.metadata.title}
            </Text>
            <Text className="line-clamp-4 text-lg" color="muted">
              {firstPost.metadata.description}
            </Text>
          </Box>
        </Box>
        <Box className="mb-20 grid grid-cols-1 gap-x-8 gap-y-12 md:grid-cols-3">
          {remainingPosts.map(
            ({ slug, metadata: { title, description, featuredImage } }) => (
              <Link className="group" href={`/blog/${slug}`} key={slug}>
                <Box className="rounded-[0.9rem] border border-dark-50 bg-dark-100/60 p-1">
                  <BlurImage
                    alt={title}
                    className="aspect-video rounded-[0.6rem]"
                    src={featuredImage}
                  />
                </Box>
                <Text
                  as="h3"
                  className="mb-3 mt-5 text-2xl group-hover:underline"
                >
                  {title}
                </Text>
                <Text className="line-clamp-4" color="muted">
                  {description}
                </Text>
              </Link>
            ),
          )}
        </Box>
      </Container>
      <CallToAction />
    </>
  );
}
