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
  name: "FortyOne",
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web",
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "10",
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
        price: "7",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Business",
        price: "10",
        priceCurrency: "USD",
      },
    ],
  },
  featureList: [
    "Project Management",
    "Team Collaboration Platform",
    "OKR Software",
    "Sprint Planning Tool",
    "OKR Management",
    "Strategic Roadmapping",
    "Team Collaboration",
    "Progress Tracking",
    "Sprint Planning",
    "Kanban Boards",
    "Analytics & Insights",
    "Tasks",
  ],
  description:
    "FortyOne is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
};

const organization: WithContext<Organization> = {
  "@context": "https://schema.org",
  "@type": "Organization",
  name: "FortyOne",
  url: "https://www.fortyone.app",
  logo: "https://www.fortyone.app/images/logo.png",
  image: "https://www.fortyone.app/images/logo.png",
  sameAs: [
    "https://x.com/fortyoneapp",
    "https://www.linkedin.com/company/complexus-app/",
    "https://github.com/complexus-tech/fortyone",
  ],
  contactPoint: {
    "@type": "ContactPoint",
    contactType: "customer support",
    email: "hello@complexus.tech",
  },
};

const product: WithContext<Product> = {
  "@context": "https://schema.org",
  "@type": "Product",
  name: "Meet FortyOne - AI-powered all-in-one Projects & OKRs platform",
  description:
    "FortyOne is an AI-powered alternative to Jira, Notion, and Monday built to align teams on Projects & OKRs, track progress, and deliver faster. Try it for free.",
  category: "Software",
  brand: {
    "@type": "Brand",
    name: "FortyOne",
    logo: "https://www.fortyone.app/images/logo.png",
  },
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "10",
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
        price: "7",
        priceCurrency: "USD",
      },
      {
        "@type": "Offer",
        name: "Business",
        price: "10",
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
      name: "How does AI improve project management workflows?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Maya accelerates planning and execution by drafting tasks from plain text, proposing sprint scope from backlog context, and suggesting objectives and key results that map to your roadmap. During execution, Maya surfaces risks, highlights stuck work, and recommends follow ups based on ownership and recent activity.",
      },
    },
    {
      "@type": "Question",
      name: "How does your project management platform connect OKRs to daily tasks?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Objectives and key results are first class. You can link tasks and sprints directly to key results so progress rolls up automatically without spreadsheet wrangling or end of quarter scrambles. Contributors see exactly how their tasks drive outcomes, while leaders get a live, trustworthy view of progress and confidence levels.",
      },
    },
    {
      "@type": "Question",
      name: "How is FortyOne priced? Is there a free plan?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. The Hobby plan is free with no credit card required and supports up to 1 team and 5 members to get started quickly. Paid plans are per user with monthly or annual billing, and annual saves 20%. Higher tiers unlock advanced workflows, custom terminology and permissions, unlimited teams, and priority support. You can upgrade, downgrade, or cancel at any time.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne handle security and privacy? Is private cloud available?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "We use industry standard encryption in transit and at rest, support SSO with Google, and provide role based permissions and private teams to control access. Audit friendly activity history and fine grained visibility help maintain compliance practices. For organizations with stricter controls, the Enterprise plan offers private cloud or on premise deployment with tailored onboarding.",
      },
    },
    {
      "@type": "Question",
      name: "Can FortyOne adapt to our workflow?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. Teams can customize statuses and workflows, define their own terminology, and set granular permissions by role or team. Automations help reduce repetitive work, and you can structure backlogs and boards to mirror how your org plans and executes. As needs evolve, you can adjust configurations without breaking historical data or reports.",
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
