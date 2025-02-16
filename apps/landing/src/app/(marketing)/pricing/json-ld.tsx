import type { WithContext, Product } from "schema-dts";

const pricingPage: WithContext<Product> = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Complexus Project Management Platform",
  description:
    "A comprehensive project management solution featuring OKR tracking, sprint planning, and team collaboration tools.",
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "12",
    offerCount: "3",
    offers: [
      {
        "@type": "Offer",
        name: "Hobby",
        description: "Track your personal objectives",
        price: "0",
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        itemOffered: {
          "@type": "SoftwareApplication",
          featureList: [
            "Use Kanban boards & lists",
            "All Integrations & API",
            "Google OAuth login",
            "Import & Export",
            "200 stories",
            "Importers",
          ],
        },
      },
      {
        "@type": "Offer",
        name: "Professional",
        description: "For small teams",
        price: "9",
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        itemOffered: {
          "@type": "SoftwareApplication",
          featureList: [
            "See product vision with Roadmaps",
            "Track OKRs",
            "Roadmaps",
            "Analytics & Reporting",
            "Custom fields",
          ],
        },
      },
      {
        "@type": "Offer",
        name: "Business",
        description: "For mid-sized teams",
        price: "12",
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        itemOffered: {
          "@type": "SoftwareApplication",
          featureList: [
            "Unlimited everything",
            "Unlimited custom workflows",
            "Single Sign-On (SSO)",
            "Priority support",
            "Discussions & Whiteboards",
            "Advanced security",
          ],
        },
      },
    ],
  },
  brand: {
    "@type": "Brand",
    name: "Complexus",
    logo: "https://complexus.app/images/logo.png",
  },
};

export const PricingJsonLd = () => {
  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(pricingPage) }}
      type="application/ld+json"
    />
  );
};
