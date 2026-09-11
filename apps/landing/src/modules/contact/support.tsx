import { MarketingResourceCard } from "@/components/shared/marketing-surface";
import { Container } from "@/components/ui";

const resources = [
  {
    title: "Help center",
    description:
      "Find setup guides and answers to help your team get the most out of FortyOne.",
    href: "https://docs.fortyone.app",
    icon: "documents",
    tone: "blue",
  },
  {
    title: "Product guides",
    description:
      "Explore practical ideas for connecting team goals, customer feedback, and daily work.",
    href: "/blog",
    icon: "blog",
    tone: "lime",
  },
  {
    title: "Integrations",
    description:
      "See how FortyOne connects with the tools your team already uses.",
    href: "/features/integrations",
    icon: "integrations",
    tone: "aqua",
  },
  {
    title: "Developers",
    description:
      "Explore the API and MCP tools for building FortyOne into your workflow.",
    href: "/developers",
    icon: "developers",
    tone: "lilac",
  },
] as const;

export const Support = () => (
  <section
    aria-labelledby="contact-resources-title"
    className="pt-20 pb-8 md:pt-28 md:pb-12"
  >
    <Container>
      <div className="text-center">
        <p className="text-text-muted text-sm">Additional resources</p>
        <h2
          className="mt-5 text-4xl font-medium text-balance md:text-5xl"
          id="contact-resources-title"
        >
          Find the help you need.
        </h2>
      </div>
      <div className="mt-10 grid gap-5 sm:grid-cols-2 lg:mt-14 lg:grid-cols-4 lg:gap-6">
        {resources.map((resource) => (
          <MarketingResourceCard key={resource.href} {...resource} />
        ))}
      </div>
    </Container>
  </section>
);
