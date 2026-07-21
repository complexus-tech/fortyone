import type {
  WebApplication,
  WithContext,
  Organization,
  Product,
  FAQPage,
} from "schema-dts";
import { homeFaqs } from "@/lib/home-faqs";
import { getCanonicalUrl } from "@/lib/seo";

const softwareApplication: WithContext<WebApplication> = {
  "@context": "https://schema.org",
  "@type": "WebApplication",
  name: "FortyOne",
  applicationCategory: "BusinessApplication",
  operatingSystem: "Web",
  url: getCanonicalUrl("/"),
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
    "Customer Feedback Management",
    "Feedback Portals",
    "Public Product Roadmaps",
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
  mainEntity: homeFaqs.map(({ question, answer }) => ({
    "@type": "Question",
    name: question,
    acceptedAnswer: {
      "@type": "Answer",
      text: answer,
    },
  })),
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
