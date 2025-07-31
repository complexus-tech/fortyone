import { Box } from "ui";
import { Hero } from "./components/hero";
import { Support } from "./components/support";
import { Stories } from "./components/stories";
import { Purpose } from "./components/purpose";
import { FounderReel } from "./components/founder-reel";
import { WhatWeDo } from "./components/what-we-do";

export const HomePage = () => {
  return (
    <Box>
      <Hero />
      <Support />
      <WhatWeDo />
      <Stories />
      <Purpose />
      <FounderReel />
    </Box>
  );
};
