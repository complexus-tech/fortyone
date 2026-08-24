export const featureLinks = [
  {
    href: "/features/customer-feedback",
    label: "Customer Feedback",
    slug: "customer-feedback",
  },
  { href: "/features/goals", label: "Goals", slug: "goals" },
  { href: "/features/tasks", label: "Tasks", slug: "tasks" },
  {
    href: "/features/ai-planning",
    label: "AI Planning",
    slug: "ai-planning",
  },
  { href: "/features/roadmaps", label: "Roadmaps", slug: "roadmaps" },
  {
    href: "/features/integrations",
    label: "Integrations",
    slug: "integrations",
  },
] as const;

export const featureLabels = Object.fromEntries(
  featureLinks.map(({ label, slug }) => [slug, label]),
) as Record<(typeof featureLinks)[number]["slug"], string>;
