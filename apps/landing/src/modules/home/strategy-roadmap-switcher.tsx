"use client";

import { RoadmapIcon, StrategyIcon } from "icons";
import { Tabs } from "ui";
import roadmapImageDark from "../../../public/images/product/roadmap-timeline-dark.webp";
import roadmapImageLight from "../../../public/images/product/roadmap-timeline-light.webp";
import strategyImageDark from "../../../public/images/product/strategy-map-dark.webp";
import strategyImageLight from "../../../public/images/product/strategy-map-light.webp";
import { ProductScreenshot } from "./product-screenshot";

const tabClassName =
  "text-text-muted hover:text-foreground focus-visible:outline-foreground h-10 min-h-0 rounded-full border-0 px-3 py-0 text-[0.9rem] font-medium transition-[padding,color,background-color,transform] duration-200 ease-out focus-visible:outline-2 focus-visible:outline-offset-1 active:scale-[0.97] data-[state=active]:border-0 data-[state=active]:bg-state-active data-[state=active]:px-4 data-[state=active]:text-foreground";

export const StrategyRoadmapSwitcher = () => {
  return (
    <Tabs className="mt-8 md:mt-10" defaultValue="strategy-map">
      <Tabs.List
        aria-label="Choose a strategy planning view"
        className="bg-state-hover text-text-muted mx-5 w-fit gap-0.5 rounded-full p-1 md:mx-auto"
      >
        <Tabs.Tab
          className={tabClassName}
          leftIcon={
            <StrategyIcon
              aria-hidden="true"
              className="text-current"
              strokeWidth={1.8}
            />
          }
          value="strategy-map"
        >
          Strategy Map
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
          value="roadmap"
        >
          Roadmap
        </Tabs.Tab>
      </Tabs.List>

      <Tabs.Panel className="outline-none" value="strategy-map">
        <ProductScreenshot
          alt="FortyOne Strategy Map connecting an ultimate goal to strategic pillars, objectives, and key results"
          containerClassName="mt-6 md:mt-10"
          darkImage={strategyImageDark}
          lightImage={strategyImageLight}
          url="https://fortyone.app/strategy"
        />
      </Tabs.Panel>

      <Tabs.Panel className="outline-none" value="roadmap">
        <ProductScreenshot
          alt="FortyOne roadmap timeline showing objective ownership, health, dates, progress, and key results"
          containerClassName="mt-6 md:mt-10"
          darkImage={roadmapImageDark}
          lightImage={roadmapImageLight}
          url="https://fortyone.app/roadmap"
        />
      </Tabs.Panel>
    </Tabs>
  );
};
