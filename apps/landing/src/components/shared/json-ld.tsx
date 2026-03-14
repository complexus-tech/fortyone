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
        text: "Maya is your AI project manager — not a chatbot bolted onto a to-do list. She drafts tasks from plain language, proposes sprint scope from your backlog, writes OKRs tied to your roadmap, and flags blockers before they cost you a sprint. She doesn't replace your team's judgment — she handles the coordination work so your team can apply it to better things. The longer she works with your setup, the more useful she gets.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne connect goals to daily work?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Goals aren't a separate module — they're built into the structure of every sprint and task. Link work directly to key results and watch objective progress update automatically as tasks close. Your team sees exactly how their work moves the quarter. Leadership gets a live, trustworthy view without sending a single 'can you send me an update?' email.",
      },
    },
    {
      "@type": "Question",
      name: "Is the free plan actually free?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes — no credit card, no trial expiry, no watered-down version. The Hobby plan supports one team and up to five members, which is enough to run a real sprint and decide whether FortyOne is worth scaling. When you're ready to grow, paid plans start at $5.60 per user per month, billed annually.",
      },
    },
    {
      "@type": "Question",
      name: "How does FortyOne handle security?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Encryption in transit and at rest, SSO with Google, role-based permissions, and private teams. For organizations with stricter requirements — compliance, on-premise, custom data residency — the Enterprise plan covers it with dedicated onboarding and a named account manager.",
      },
    },
    {
      "@type": "Question",
      name: "Can we make it fit the way our team works?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "That's the point. Customize statuses, workflows, terminology, and permissions to match your org — not a generic process template someone wrote in 2018. Structure your backlog and boards around how you plan and ship, and adjust as you grow without breaking historical context.",
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
