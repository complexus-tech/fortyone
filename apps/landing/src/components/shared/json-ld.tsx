import type {
  WebApplication,
  WithContext,
  Organization,
  Product,
  FAQPage,
} from "schema-dts";

const softwareApplication: WithContext<WebApplication> = {
  "@context": "https://schema.org",
  "@type": "WebApplication",
  name: "Complexus",
  applicationCategory: "ProjectManagementApplication",
  operatingSystem: "Web",
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
        price: "0",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Professional",
        price: "9",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Business",
        price: "12",
        priceCurrency: "USD",
      },
    ],
  },
  featureList: [
    "OKR Management",
    "Strategic Roadmapping",
    "Team Collaboration",
    "Progress Tracking",
    "Sprint Planning",
    "Kanban Boards",
    "Analytics & Insights",
    "User Stories",
  ],
  description:
    "Complexus is a comprehensive project management platform that helps teams set and achieve objectives through OKR tracking, sprint planning, and collaborative tools.",
};

const organization: WithContext<Organization> = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "Complexus",
  url: "https://complexus.app",
  sameAs: ["https://x.com/complexus_app"],
  contactPoint: {
    "@type": "ContactPoint",
    contactType: "customer support",
    email: "support@complexus.app",
  },
};

const product: WithContext<Product> = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Complexus Project Management Platform",
  description:
    "A comprehensive project management solution featuring OKR tracking, sprint planning, and team collaboration tools.",
  category: "Software",
  brand: {
    "@type": "Brand",
    name: "Complexus",
  },
};

const faq: WithContext<FAQPage> = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "What is Complexus?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Complexus is a modern project management platform that helps teams set, track, and achieve objectives through OKR management, sprint planning, and collaborative tools.",
      },
    },
    {
      "@type": "Question",
      name: "What features does Complexus offer?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Complexus offers comprehensive features including OKR management, sprint planning, kanban boards, user stories, strategic roadmapping, team collaboration tools, and analytics & insights.",
      },
    },
    {
      "@type": "Question",
      name: "How does Complexus help with team alignment?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Complexus helps teams align through shared objectives, real-time updates, integrated team workspaces, and comprehensive OKR tracking tools that ensure everyone is working towards common goals.",
      },
    },
  ],
};

export const JsonLd = () => {
  return (
    <>
      <script
        dangerouslySetInnerHTML={{
          __html: JSON.stringify(softwareApplication),
        }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(organization) }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(product) }}
        type="application/ld+json"
      />
      <script
        dangerouslySetInnerHTML={{ __html: JSON.stringify(faq) }}
        type="application/ld+json"
      />
    </>
  );
};
