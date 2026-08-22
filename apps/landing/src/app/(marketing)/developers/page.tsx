import type { Metadata } from "next";
import { Box, Text } from "ui";
import { Container } from "@/components/ui";
import { getCanonicalUrl } from "@/lib/seo";

export const metadata: Metadata = {
  title: "Developers | FortyOne",
  description:
    "Discover FortyOne API documentation, the OpenAPI description, MCP server, authentication boundaries, and integration resources.",
  alternates: { canonical: getCanonicalUrl("/developers") },
};

const resources = [
  {
    title: "OpenAPI 3.1",
    description:
      "Inspect typed public endpoints, unique operation IDs, response schemas, and authentication requirements.",
    href: "/openapi.json",
    label: "View OpenAPI JSON",
  },
  {
    title: "Model Context Protocol",
    description:
      "Connect ChatGPT, Claude, Cursor, Codex, and other MCP clients to permission-aware FortyOne tools.",
    href: "/server.json",
    label: "View MCP metadata",
  },
  {
    title: "Product documentation",
    description:
      "Read guides for FortyOne product concepts, workflows, integrations, and workspace administration.",
    href: "https://docs.fortyone.app",
    label: "Read the docs",
  },
];

export default function Page() {
  return (
    <main>
      <Container className="pt-28 pb-20 md:pt-40 md:pb-28">
        <Box className="max-w-3xl">
          <Text className="text-primary mb-4 font-mono text-sm uppercase">
            FortyOne for developers and agents
          </Text>
          <Text
            as="h1"
            className="text-5xl font-semibold text-balance md:text-6xl"
          >
            Build from a documented, machine-readable surface.
          </Text>
          <Text className="text-text-muted mt-6 max-w-2xl text-xl leading-relaxed">
            Discover the public FortyOne API contract, agent-readable content,
            and Model Context Protocol endpoint. Authenticated product data
            remains governed by workspace and team permissions.
          </Text>
        </Box>

        <Box className="mt-16 grid gap-4 md:mt-24 md:grid-cols-3">
          {resources.map(({ title, description, href, label }) => (
            <section
              className="border-border rounded-2xl border p-6"
              key={title}
            >
              <Text as="h2" className="text-xl font-semibold">
                {title}
              </Text>
              <Text className="text-text-muted mt-3 leading-relaxed">
                {description}
              </Text>
              <a
                className="mt-6 inline-block font-medium underline underline-offset-4"
                href={href}
              >
                {label}
              </a>
            </section>
          ))}
        </Box>

        <section className="mt-16 max-w-3xl md:mt-24">
          <Text as="h2" className="text-3xl font-semibold">
            Authentication boundary
          </Text>
          <Text className="text-text-muted mt-4 leading-relaxed">
            The remote MCP server uses a separate FortyOne OAuth connection with
            PKCE, audience-bound tokens, refresh-token rotation, revocation, and
            an explicit consent screen. Every tool runs as the connected user
            and reuses existing workspace and team permission checks. Create
            tools also require explicit user confirmation.
          </Text>
        </section>
      </Container>
    </main>
  );
}
