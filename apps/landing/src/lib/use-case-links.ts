export const useCaseLinks = [
  { href: "/use-cases/operations", label: "Operations", slug: "operations" },
  { href: "/use-cases/product", label: "Product", slug: "product" },
  {
    href: "/use-cases/customer-support",
    label: "Support",
    slug: "customer-support",
  },
  { href: "/use-cases/developers", label: "Developers", slug: "developers" },
  {
    href: "/use-cases/field-crews",
    label: "Field Crews",
    slug: "field-crews",
  },
  { href: "/use-cases/marketing", label: "Marketing", slug: "marketing" },
  { href: "/use-cases/leadership", label: "Leadership", slug: "leadership" },
] as const;

export const useCaseLabels = Object.fromEntries(
  useCaseLinks.map(({ label, slug }) => [slug, label]),
) as Record<(typeof useCaseLinks)[number]["slug"], string>;

export const primaryUseCaseLinks = useCaseLinks.filter(({ slug }) =>
  [
    "operations",
    "product",
    "customer-support",
    "developers",
    "marketing",
  ].includes(slug),
);
