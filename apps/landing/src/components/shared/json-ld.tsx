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
    "FortyOne is the AI project management platform where every task rolls up to a goal and Maya keeps work moving from planning through delivery.",
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
  name: "FortyOne - AI project management platform for teams",
  description:
    "FortyOne is the AI project management platform where every task rolls up to a goal and Maya keeps work moving from planning through delivery.",
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
      name: "What does Maya actually do?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Maya is your AI project manager - not just a chatbot bolted onto a to-do list. She drafts tasks from plain text, proposes sprint scope from your backlog, writes goals and key results that connect to your roadmap, and flags blockers in real time based on ownership and activity.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne connect goals to daily work?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Goals are built into the structure of every sprint and task. Link work directly to key results and watch progress roll up automatically. Team members see how their work drives outcomes, and leaders get a live, trustworthy view of where things stand without spreadsheet wrangling.",
      },
    },
    {
      "@type": "Question",
      name: "Is there really a free plan?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. The Hobby plan is free with no credit card required and supports up to 1 team and 5 members, which is enough to get real work done and decide if FortyOne is right for you. Paid plans are per user, with annual billing saving 20%. You can upgrade, downgrade, or cancel at any time.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne handle security?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "We use industry-standard encryption in transit and at rest, support SSO with Google, and provide role-based permissions and private teams to control access. For organizations with stricter requirements, the Enterprise plan includes private cloud or on-premise deployment with tailored onboarding.",
      },
    },
    {
      "@type": "Question",
      name: "Can we make it fit the way our team works?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. Customize statuses, workflows, terminology, and permissions to match your org - not a generic process template. Structure your backlog and boards around the way your team plans and executes, and adjust things as you grow without breaking historical context.",
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
