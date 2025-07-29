import { Box } from "ui";
import { Hero } from "./components/hero";
import { Support } from "./components/support";
import { Stories } from "./components/stories";
import { Purpose } from "./components/purpose";
import { FounderReel } from "./components/founder-reel";

export const HomePage = () => {
  return (
    <Box>
      <Hero />
      <Support />
      <Stories />
      <Purpose />
      <FounderReel />
    </Box>
  );
};
