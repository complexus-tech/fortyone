import { Box } from "ui";
import { Hero } from "./components/hero";
import { Support } from "./components/support";
import { Stories } from "./components/stories";

export const HomePage = () => {
  return (
    <Box>
      <Hero />
      <Support />
      <Stories />
    </Box>
  );
};
