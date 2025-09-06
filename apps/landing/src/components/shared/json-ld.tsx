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
  name: "Forty One",
  applicationCategory: "ProjectManagementApplication",
  operatingSystem: "Web",
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
        price: "0",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Professional",
        price: "5.60",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Business",
        price: "8",
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
    "Complexus is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
};

const organization: WithContext<Organization> = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "Forty One",
  url: "https://complexus.app",
  logo: "https://complexus.app/images/logo.png",
  image: "https://complexus.app/images/logo.png",
  sameAs: ["https://x.com/fortyoneapp"],
  contactPoint: {
    "@type": "ContactPoint",
    contactType: "customer support",
    email: "support@complexus.app",
  },
};

const product: WithContext<Product> = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Meet Forty One - AI-powered all-in-one Projects & OKRs platform",
  description:
    "Forty One is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
  category: "Software",
  brand: {
    "@type": "Brand",
    name: "Forty One",
    logo: "https://fortyone.app/images/logo.png",
  },
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
        price: "0",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Professional",
        price: "5.60",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Business",
        price: "8",
        priceCurrency: "USD",
      },
    ],
  },
};

const faq: WithContext<FAQPage> = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  mainEntity: [
    {
      "@type": "Question",
      name: "What is Forty One?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Forty One is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
      },
    },
    {
      "@type": "Question",
      name: "What features does Forty One offer?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Forty One offers a complete suite of project management features including OKR management, sprint planning, kanban boards, user stories, strategic roadmapping, real-time team collaboration tools, customizable dashboards, progress tracking, and in-depth analytics & insights to drive team performance.",
      },
    },
    {
      "@type": "Question",
      name: "How does Forty One help with team alignment?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Forty One helps teams align by connecting high-level objectives with tactical work through its integrated OKR framework. Teams benefit from shared objectives, real-time progress updates, integrated workspaces, and visual tracking tools that ensure everyone understands how their work contributes to organizational goals.",
      },
    },
    {
      "@type": "Question",
      name: "Can Forty One integrate with other tools my team uses?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes, Forty One integrates seamlessly with popular collaboration, development, and productivity tools including Slack, Microsoft Teams, GitHub, GitLab, Jira, Figma, and Google Workspace, allowing teams to maintain their existing workflows while gaining the benefits of centralized project management.",
      },
    },
    {
      "@type": "Question",
      name: "Is Forty One suitable for remote or distributed teams?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Absolutely. Forty One is built with remote and distributed teams in mind, offering real-time collaboration features, asynchronous communication tools, transparent goal tracking, and centralized documentation that keeps everyone aligned regardless of time zone or location.",
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
