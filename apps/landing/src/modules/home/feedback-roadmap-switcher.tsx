"use client";

import { RequestsIcon, RoadmapIcon } from "icons";
import { Tabs } from "ui";
import { Container } from "@/components/ui";
import feedbackImageDark from "../../../public/images/product/feedback-portal-dark.webp";
import feedbackImageLight from "../../../public/images/product/feedback-portal-light.webp";
import publicRoadmapImageDark from "../../../public/images/product/public-feedback-roadmap-dark.webp";
import publicRoadmapImageLight from "../../../public/images/product/public-feedback-roadmap-light.webp";
import { ProductScreenshot } from "./product-screenshot";

const tabClassName =
  "text-text-muted hover:text-foreground focus-visible:outline-foreground h-10 min-h-0 rounded-full border-0 px-3 py-0 text-[0.9rem] font-medium transition-[padding,color,background-color,transform] duration-200 ease-out focus-visible:outline-2 focus-visible:outline-offset-1 active:scale-[0.97] data-[state=active]:border-0 data-[state=active]:bg-state-active data-[state=active]:px-4 data-[state=active]:text-foreground";

export const FeedbackRoadmapSwitcher = () => {
  return (
    <Tabs className="mt-8 md:mt-10" defaultValue="customer-feedback">
      <Container>
        <Tabs.List
          aria-label="Choose a customer feedback view"
          className="bg-state-hover text-text-muted mx-0 w-fit gap-0.5 rounded-full p-1 md:mx-0"
        >
          <Tabs.Tab
            className={tabClassName}
            leftIcon={
              <RequestsIcon
                aria-hidden="true"
                className="text-current"
                strokeWidth={1.8}
              />
            }
            value="customer-feedback"
          >
            Collect Feedback
          </Tabs.Tab>
          <Tabs.Tab
            className={tabClassName}
            leftIcon={
              <RoadmapIcon
                aria-hidden="true"
                className="text-current"
                strokeWidth={1.8}
              />
            }
            value="public-roadmap"
          >
            Public Roadmap
          </Tabs.Tab>
        </Tabs.List>
      </Container>

      <Tabs.Panel className="outline-none" value="customer-feedback">
        <ProductScreenshot
          alt="Public FortyOne feedback portal showing customer requests, votes, boards, and delivery statuses"
          containerClassName="mt-6 md:mt-10"
          darkImage={feedbackImageDark}
          lightImage={feedbackImageLight}
          url="https://fortyone.app/feedback"
        />
      </Tabs.Panel>

      <Tabs.Panel className="outline-none" value="public-roadmap">
        <ProductScreenshot
          alt="Public FortyOne roadmap showing customer requests grouped into Planned, In Progress, and Done"
          containerClassName="mt-6 md:mt-10"
          darkImage={publicRoadmapImageDark}
          lightImage={publicRoadmapImageLight}
          url="https://fortyone.app/roadmap"
        />
      </Tabs.Panel>
    </Tabs>
  );
};
