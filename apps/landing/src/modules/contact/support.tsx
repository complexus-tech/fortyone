import { Box, Text } from "ui";
import { Container } from "@/components/ui";

const contactRoutes = [
  {
    heading: "Product and pricing",
    email: "hello@complexus.tech",
    description:
      "For demos, pricing questions, procurement, or deciding which plan fits your team.",
  },
  {
    heading: "Implementation support",
    email: "hello@complexus.tech",
    description:
      "For setup, integrations, workspace planning, or help rolling FortyOne out to a team.",
  },
];

const contactNotes = [
  "What kind of work your team manages",
  "How many people or teams will use FortyOne",
  "Which tools you want to connect",
  "Any deadline, rollout, or security requirements",
];

export const Support = () => {
  return (
    <Container className="pb-20 md:pb-24">
      <Box className="grid max-w-5xl gap-6 md:grid-cols-[1.15fr_0.85fr]">
        <Box className="grid gap-4">
          {contactRoutes.map(({ heading, description, email }) => (
            <Box
              className="border-border bg-surface rounded-2xl border p-6"
              key={heading}
            >
              <Text as="h2" className="text-foreground text-xl font-medium">
                {heading}
              </Text>
              <Text className="text-text-muted mt-3 leading-7">
                {description}
              </Text>
              <a
                className="bg-background-inverse text-foreground-inverse mt-5 inline-flex rounded-lg px-4 py-2.5 transition-opacity hover:opacity-90"
                href={`mailto:${email}`}
              >
                {email}
              </a>
            </Box>
          ))}
        </Box>

        <Box className="bg-surface-muted rounded-2xl p-6 dark:bg-white/[0.06]">
          <Text as="h2" className="text-foreground text-xl font-medium">
            What helps us respond faster
          </Text>
          <Text className="text-text-muted mt-3 leading-7">
            A little context makes it easier to route your note to the right
            person and give you a useful answer.
          </Text>
          <Box className="mt-6 grid gap-3">
            {contactNotes.map((note) => (
              <Box className="flex gap-3" key={note}>
                <span className="bg-foreground mt-2 size-1.5 shrink-0 rounded-full" />
                <Text className="text-text-muted">{note}</Text>
              </Box>
            ))}
          </Box>
        </Box>
      </Box>
    </Container>
  );
};
