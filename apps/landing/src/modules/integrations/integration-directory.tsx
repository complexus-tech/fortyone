import { Container } from "@/components/ui";
import {
  IntegrationBrand,
  type IntegrationBrandName,
} from "./integration-brand";

const CONNECTIONS: readonly {
  name: IntegrationBrandName;
  description: string;
}[] = [
  {
    name: "Slack",
    description: "Turn conversations into work with the source attached.",
  },
  {
    name: "GitHub",
    description: "Connect issues, pull requests, and delivery updates.",
  },
  {
    name: "Google Calendar",
    description: "Plan work around meetings and real availability.",
  },
  {
    name: "Outlook Calendar",
    description: "Bring calendar commitments into your work plan.",
  },
  {
    name: "Google Drive",
    description: "Keep work files close to the decisions they support.",
  },
  {
    name: "Figma",
    description: "Bring design context into the delivery conversation.",
  },
  {
    name: "ChatGPT",
    description: "Work with permitted FortyOne context through MCP.",
  },
  {
    name: "Claude",
    description: "Bring your project context into an MCP conversation.",
  },
  {
    name: "Cursor",
    description: "Keep implementation connected to permitted project work.",
  },
];

export function IntegrationDirectory() {
  return (
    <Container
      aria-labelledby="connections-title"
      as="section"
      className="scroll-mt-24 py-16 md:py-28"
      id="connections"
    >
      <div
        className="grid gap-6 md:grid-cols-2 md:items-end md:gap-20"
        data-landing-reveal
      >
        <h2 className="max-w-xl text-3xl md:text-5xl" id="connections-title">
          Your tools, connected.
        </h2>
        <p className="text-text-description max-w-lg leading-relaxed">
          Bring conversations, code, calendars, and source context into the same
          plan. Explore the connections your team already uses.
        </p>
      </div>
      <ul className="border-border/60 mt-12 grid overflow-hidden border-y sm:grid-cols-2 lg:grid-cols-3">
        {CONNECTIONS.map(({ name, description }) => (
          <li
            className="before:bg-border/60 after:bg-border/60 relative scroll-mt-28 px-6 py-8 before:absolute before:inset-x-0 before:-top-px before:h-px after:absolute after:inset-y-0 after:-left-px after:w-px md:px-8 md:py-10"
            id={name.toLowerCase().replaceAll(" ", "-")}
            key={name}
          >
            <IntegrationBrand name={name} size={36} />
            <h3 className="mt-6 text-lg font-semibold">{name}</h3>
            <p className="text-text-description mt-2 max-w-xs text-base leading-relaxed">
              {description}
            </p>
          </li>
        ))}
      </ul>
    </Container>
  );
}
