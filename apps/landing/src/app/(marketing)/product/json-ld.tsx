import type { WithContext, SoftwareApplication } from "schema-dts";

const productPage: WithContext<SoftwareApplication> = {
  "@context": "https://schema.org",
  "@type": "SoftwareApplication",
  name: "Complexus Project Management Platform",
  applicationCategory: "ProjectManagementApplication",
  operatingSystem: "Web",
  description:
    "A comprehensive project management platform featuring OKR tracking, sprint planning, and team collaboration tools.",
  featureList: [
    "OKR Management - Set and track Objectives and Key Results (OKRs) with powerful tools for measuring progress and aligning teams.",
    "Strategic Roadmapping - Plan and visualize your organization's journey with interactive roadmaps and milestone tracking.",
    "Team Collaboration - Foster cross-functional teamwork with shared objectives, real-time updates, and integrated team workspaces.",
    "Progress Tracking - Monitor objective health, track key results, and get insights into team performance with comprehensive analytics.",
    "Sprint Planning - Plan and execute sprints with confidence. Track velocity and deliver predictably.",
    "Kanban Boards - Visualize work in progress and optimize your team's flow with intuitive drag-and-drop interface.",
    "Analytics & Insights - Make informed decisions with real-time metrics and customizable dashboards.",
    "User Stories - Create, organize, and prioritize work items with clarity and purpose.",
  ],
  offers: {
    "@type": "AggregateOffer",
    priceCurrency: "USD",
    lowPrice: "0",
    highPrice: "12",
    offerCount: "3",
  },
};

export const ProductJsonLd = () => {
  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(productPage) }}
      type="application/ld+json"
    />
  );
};
