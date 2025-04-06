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
    "Complexus is an advanced project management platform that integrates OKR frameworks with agile methodologies, helping teams align objectives, track progress, and deliver consistent results through collaborative workflows.",
};

const organization: WithContext<Organization> = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "Complexus",
  url: "https://complexus.app",
  logo: "https://complexus.app/images/logo.png",
  image: "https://complexus.app/images/logo.png",
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
  name: "Project Management & OKR Software for Teams",
  description:
    "An integrated project management solution that combines OKR frameworks, sprint planning, and collaborative tools to help teams set clear objectives, execute efficiently, and deliver measurable outcomes.",
  category: "Software",
  brand: {
    "@type": "Brand",
    name: "Complexus",
    logo: "https://complexus.app/images/logo.png",
  },
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
        text: "Complexus is a comprehensive project management platform designed for modern teams. It combines OKR tracking, sprint planning, and collaborative workflows to help teams align strategic objectives with day-to-day execution, ensuring measurable results and improved team productivity.",
      },
    },
    {
      "@type": "Question",
      name: "What features does Complexus offer?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Complexus offers a complete suite of project management features including OKR management, sprint planning, kanban boards, user stories, strategic roadmapping, real-time team collaboration tools, customizable dashboards, progress tracking, and in-depth analytics & insights to drive team performance.",
      },
    },
    {
      "@type": "Question",
      name: "How does Complexus help with team alignment?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Complexus helps teams align by connecting high-level objectives with tactical work through its integrated OKR framework. Teams benefit from shared objectives, real-time progress updates, integrated workspaces, and visual tracking tools that ensure everyone understands how their work contributes to organizational goals.",
      },
    },
    {
      "@type": "Question",
      name: "Can Complexus integrate with other tools my team uses?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes, Complexus integrates seamlessly with popular collaboration, development, and productivity tools including Slack, Microsoft Teams, GitHub, GitLab, Jira, Figma, and Google Workspace, allowing teams to maintain their existing workflows while gaining the benefits of centralized project management.",
      },
    },
    {
      "@type": "Question",
      name: "Is Complexus suitable for remote or distributed teams?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Absolutely. Complexus is built with remote and distributed teams in mind, offering real-time collaboration features, asynchronous communication tools, transparent goal tracking, and centralized documentation that keeps everyone aligned regardless of time zone or location.",
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
