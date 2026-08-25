import Link from "next/link";
import { Box, Button, Text } from "ui";
import { Container } from "@/components/ui";

export default function NotFound() {
  return (
    <main>
      <Container className="py-32 md:py-44">
        <Box className="mx-auto max-w-2xl text-center">
          <Text className="text-text-muted mb-4 font-mono text-sm">404</Text>
          <Text
            as="h1"
            className="text-4xl font-semibold text-balance md:text-5xl"
          >
            This page does not exist.
          </Text>
          <Text className="text-text-muted mx-auto mt-5 max-w-xl text-lg">
            Try the FortyOne home page, browse the documentation, or use the
            sitemap to find the resource you need.
          </Text>
          <Box className="mt-8 flex flex-wrap justify-center gap-3">
            <Button color="invert" href="/" rounded="md">
              Go home
            </Button>
            <Button
              color="tertiary"
              href="https://docs.fortyone.app"
              rounded="md"
            >
              Read the docs
            </Button>
          </Box>
          <Link
            className="text-text-muted mt-6 inline-block text-sm underline underline-offset-4"
            href="/sitemap.xml"
          >
            View sitemap
          </Link>
        </Box>
      </Container>
    </main>
  );
}
