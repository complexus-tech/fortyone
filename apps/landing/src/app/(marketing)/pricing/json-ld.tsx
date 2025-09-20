import type { WithContext, Product } from "schema-dts";

const pricingPage: WithContext<Product> = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "FortyOne Project Management Platform",
  description:
    "A comprehensive project management solution featuring OKR tracking, sprint planning, and team collaboration tools.",
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "8",
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
            "1 team",
            "Up to 5 members",
            "Up to 200 stories",
            "Single Sign-On (SSO)",
            "Kanban & list views",
            "Email support",
          ],
        },
      },
      {
        "@type": "Offer",
        name: "Professional",
        description: "For small teams",
        price: "5.60",
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        itemOffered: {
          "@type": "SoftwareApplication",
          featureList: [
            "Up to 3 teams",
            "Up to 20 objectives",
            "Track OKRs",
            "Unlimited stories",
            "Unlimited guests",
            "Custom workflows",
          ],
        },
      },
      {
        "@type": "Offer",
        name: "Business",
        description: "For mid-sized teams",
        price: "8",
        priceCurrency: "USD",
        availability: "https://schema.org/InStock",
        itemOffered: {
          "@type": "SoftwareApplication",
          featureList: [
            "Unlimited teams",
            "Unlimited objectives",
            "Unlimited everything",
            "Custom terminology",
            "Private teams",
            "Priority support",
          ],
        },
      },
    ],
  },
  brand: {
    "@type": "Brand",
    name: "FortyOne",
    logo: "https://fortyone.app/images/logo.png",
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
