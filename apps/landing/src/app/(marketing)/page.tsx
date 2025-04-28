import { Box } from "ui";
import { CallToAction } from "@/components/shared";
import { Pricing } from "@/components/ui";
import {
  Hero,
  HeroCards,
  SampleClients,
  Features,
  Integrations,
  Testimonials,
  Transform,
} from "@/modules/home";

export default function Page() {
  return (
    <>
      <Hero />
      <HeroCards />
      <SampleClients />
      <Testimonials />
      <Features />
      <Transform />
      <Box className="pb-16 md:pb-28">
        <Pricing />
      </Box>
      <Integrations />
      <CallToAction />
    </>
  );
}
