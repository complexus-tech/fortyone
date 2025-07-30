import { Box } from "ui";
import { Hero } from "./components/hero";
import { GivingBack } from "./components/giving-back";
import { Welcome } from "./components/welcome";
import { Together } from "./components/together";
import { Future } from "./components/future";
import { Compliance } from "./components/compliance";
import { TrustTransparency } from "./components/trust-transparency";

export const AboutPage = () => {
  return (
    <Box>
      <Hero />
      <Together />
      <Future />
      <Welcome />
      <GivingBack />
      <TrustTransparency />
      <Compliance />
    </Box>
  );
};
