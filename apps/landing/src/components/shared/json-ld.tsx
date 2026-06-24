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
    "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
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
    "FortyOne is AI project management software for modern teams. Plan projects, assign tasks, track goals, and let AI find the right owner, estimate, schedule, and next step.",
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
      name: "What makes FortyOne an AI project manager?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "FortyOne combines project management software with an AI assistant that can turn project context into tasks, suggest owners, add estimates, plan timing, and surface delivery risks before work slips.",
      },
    },
    {
      "@type": "Question",
      name: "What happens when I assign work to AI?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "The AI reviews the task, team context, workload, estimates, and availability, then helps find the right owner, schedule, and next action. Admins can review important AI actions before they are applied.",
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
      name: "Can FortyOne plan around my team's calendar?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. Google Calendar integration lets FortyOne sync availability so AI can recommend better schedules and work windows without storing private event details unnecessarily.",
      },
    },
    {
      "@type": "Question",
      name: "Can we create tasks from Slack?",
      acceptedAnswer: {
        "@type": "Answer",
        text: "Yes. Slack support lets teams create tasks from Slack and ask the AI assistant for help where conversations already happen, while keeping the project plan in FortyOne.",
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
