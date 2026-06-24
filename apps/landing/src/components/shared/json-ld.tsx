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
    "FortyOne connects goals, sprint plans, tasks, and Maya in one AI project management workspace, so teams can plan faster and spot delivery risks early.",
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
    "FortyOne connects goals, sprint plans, tasks, and Maya in one AI project management workspace, so teams can plan faster and spot delivery risks early.",
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
        text: "Maya is the AI project manager inside FortyOne. She turns rough outcomes into tasks, proposes sprint scope from your backlog and capacity, connects work to goals, and flags blockers before they become end-of-sprint surprises.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne connect goals to daily work?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Goals are not a reporting layer pasted on top of tasks. Work can be linked directly to objectives and key results, so progress updates as tasks move and leaders can see what changed without asking for another status report.",
      },
    },
    {
      "@type": "Question",
      name: "Is the free plan actually free?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. There is no credit card and no trial expiry. The Hobby plan supports one team and up to five members, enough to run a real sprint and decide whether FortyOne should scale with you.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne handle security?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "FortyOne supports encryption in transit and at rest, Google SSO, role-based permissions, and private teams. Enterprise teams can work with us on stricter requirements such as dedicated onboarding, private cloud, and deployment preferences.",
      },
    },
    {
      "@type": "Question",
      name: "Can we make it fit the way our team works?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. Customize statuses, workflows, terminology, permissions, teams, and planning rules around how you already ship. The structure can evolve as the team grows without throwing away the history behind prior work.",
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
