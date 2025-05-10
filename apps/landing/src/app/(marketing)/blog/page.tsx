import type { Metadata } from "next";
import Link from "next/link";
import { BlurImage, Box, Button, Divider, Flex, Text } from "ui";
import { ArrowRight2Icon } from "icons";
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
      <Container className="mt-16 max-w-7xl pt-12 md:pt-16">
        <BlogJsonLd />
        <Link
          className="group grid grid-cols-1 items-center gap-8 md:grid-cols-[1.5fr_1fr]"
          href={`/blog/${firstPost.slug}`}
        >
          <Box className="rounded-[0.9rem] border border-dark-50 bg-dark-100/60 p-2">
            <BlurImage
              alt={firstPost.metadata.title}
              className="aspect-[16/9.5] rounded-[0.6rem]"
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
              className="mb-3 text-4xl font-semibold group-hover:underline"
            >
              {firstPost.metadata.title}
            </Text>
            <Text className="mb-4 line-clamp-4 text-lg opacity-80">
              {firstPost.metadata.description}
            </Text>
            <Button
              className="gap-1"
              color="white"
              rightIcon={<ArrowRight2Icon className="dark:text-dark" />}
            >
              Read More
            </Button>
          </Box>
        </Link>
        <Divider className="my-10" />
        <Box className="mb-20 grid grid-cols-1 gap-x-8 gap-y-12 md:grid-cols-3">
          {[
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
            firstPost,
            ...remainingPosts,
          ].map(({ slug, metadata: { title, featuredImage, date } }, idx) => (
            <Link
              className="group"
              href={`/blog/${slug}`}
              key={`${slug}-${idx}`}
            >
              <Box className="rounded-[0.9rem] border border-dark-50 bg-dark-100/60 p-1.5">
                <BlurImage
                  alt={title}
                  className="aspect-video rounded-[0.6rem]"
                  src={featuredImage}
                />
              </Box>
              <Flex align="center" className="my-3" justify="between">
                <Text className="opacity-80">
                  {date
                    ? new Date(date).toLocaleDateString("en-US", {
                        month: "long",
                        day: "numeric",
                        year: "numeric",
                      })
                    : null}
                </Text>
                <Text className="opacity-80">6 min read</Text>
              </Flex>
              <Text
                as="h3"
                className="text-2xl font-semibold group-hover:underline"
              >
                {title}
              </Text>
            </Link>
          ))}
        </Box>
      </Container>
      <CallToAction />
    </>
  );
}
