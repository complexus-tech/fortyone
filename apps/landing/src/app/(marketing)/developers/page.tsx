import type { Metadata } from "next";
import { Button } from "ui";
import { Container } from "@/components/ui";
import {
  MarketingHero,
  MarketingResourceCard,
} from "@/components/shared/marketing-surface";
import { NavigationMenuIcon } from "@/components/shared/navigation-menu-icon";
import { CallToAction } from "@/components/shared/cta";
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
    icon: "api-reference",
    tone: "blue",
  },
  {
    title: "Model Context Protocol",
    description:
      "Connect ChatGPT, Claude, Cursor, Codex, and other MCP clients to permission-aware FortyOne tools.",
    href: "/server.json",
    icon: "ai-planning",
    tone: "lilac",
  },
  {
    title: "Product documentation",
    description:
      "Read guides for FortyOne product concepts, workflows, integrations, and workspace administration.",
    href: "https://docs.fortyone.app",
    icon: "documents",
    tone: "lime",
  },
  {
    title: "API reference",
    description:
      "Explore API endpoints, authentication, request parameters, and response examples for your integration.",
    href: "https://docs.fortyone.app/api-reference",
    icon: "api-reference",
    tone: "aqua",
  },
] as const;

export default function Page() {
  return (
    <>
      <main>
        <MarketingHero
          description="Discover the public FortyOne API contract, agent-readable content, and Model Context Protocol endpoint. Authenticated product data remains governed by workspace and team permissions."
          eyebrow="FortyOne for developers and agents"
          id="developers-title"
          title={
            <>
              Build on your
              <br />
              team&apos;s context.
            </>
          }
        >
          <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Button
              className="w-full justify-center sm:w-auto"
              color="invert"
              href="https://docs.fortyone.app/api-reference"
              rounded="md"
              size="lg"
            >
              Explore the API
            </Button>
            <Button
              className="w-full justify-center sm:w-auto"
              color="invert"
              href="/server.json"
              rounded="md"
              size="lg"
              variant="outline"
            >
              View MCP metadata
            </Button>
          </div>
        </MarketingHero>
        <section
          aria-labelledby="developer-resources"
          className="pt-20 pb-16 md:pt-28 md:pb-24"
        >
          <Container>
            <div className="text-center">
              <p className="text-text-muted text-sm">Developer resources</p>
              <h2
                className="mt-5 text-4xl font-medium text-balance md:text-5xl"
                id="developer-resources"
              >
                Everything you need to build.
              </h2>
            </div>
            <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:mt-14 lg:grid-cols-4 lg:gap-6">
              {resources.map((resource) => (
                <MarketingResourceCard key={resource.href} {...resource} />
              ))}
            </div>
            <section
              aria-labelledby="authentication-title"
              className="border-border mt-16 grid gap-6 border-t pt-12 md:mt-24 md:grid-cols-[1fr_1.4fr] md:gap-16 md:pt-16"
            >
              <div>
                <NavigationMenuIcon
                  className="size-11"
                  name="developers"
                  tone="aqua"
                />
                <h2
                  className="mt-5 text-3xl font-medium text-balance md:text-4xl"
                  id="authentication-title"
                >
                  Your permissions stay in place.
                </h2>
              </div>
              <div className="text-text-description space-y-5 self-center">
                <p>
                  The remote MCP server uses a separate FortyOne OAuth
                  connection with PKCE, audience-bound tokens, refresh-token
                  rotation, revocation, and an explicit consent screen.
                </p>
                <p>
                  Every tool runs as the connected user and reuses existing
                  workspace and team permission checks. Create tools also
                  require explicit user confirmation.
                </p>
              </div>
            </section>
          </Container>
        </section>
      </main>
      <CallToAction />
    </>
  );
}
