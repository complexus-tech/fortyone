import { Box } from "ui";
import { Container } from "@/components/ui";
import { VoiceSection } from "./components/voice-section";
import { ContactForm } from "./components/contact-form";
import { NewsletterForm } from "./components/newsletter-form";

export const ContactPage = () => {
  return (
    <Box>
      <Container className="max-w-7xl">
        <VoiceSection />
        <Box className="grid grid-cols-1 gap-16 pb-20 md:grid-cols-2">
          <ContactForm />
          <NewsletterForm />
        </Box>
      </Container>
    </Box>
  );
};
