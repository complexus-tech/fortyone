import { Button } from "ui";
import { MarketingHero } from "@/components/shared/marketing-surface";

const CONTACT_EMAIL = "info@fortyone.app";

function contactHref(subject: string, body: string) {
  return `mailto:${CONTACT_EMAIL}?subject=${encodeURIComponent(subject)}&body=${encodeURIComponent(body)}`;
}

const salesHref = contactHref(
  "FortyOne — product and pricing",
  "Hi FortyOne team,\n\nI'd like to learn more about FortyOne.\n\nThe work our team manages:\nTeam size:\nTools we want to connect:\nRollout timeline or requirements:\n",
);

const supportHref = contactHref(
  "FortyOne — support request",
  "Hi FortyOne team,\n\nI need help with:\n\nWorkspace:\nWhat I was trying to do:\nWhat happened:\n",
);

export const Hero = () => (
  <MarketingHero
    description="Ask about pricing, implementation, integrations, support, or whether FortyOne is the right fit for the way your team plans and tracks work."
    eyebrow="Contact us"
    id="contact-title"
    title="Talk to the team behind FortyOne."
  >
    <div className="mt-8 flex flex-col items-center justify-center gap-3 sm:flex-row">
      <Button
        className="focus-visible:outline-ring w-full justify-center focus-visible:outline-2 focus-visible:outline-offset-4 sm:w-auto"
        color="invert"
        href={salesHref}
        rounded="md"
        size="lg"
      >
        Contact sales
      </Button>
      <Button
        className="focus-visible:outline-ring w-full justify-center focus-visible:outline-2 focus-visible:outline-offset-4 sm:w-auto"
        color="invert"
        href={supportHref}
        rounded="md"
        size="lg"
        variant="outline"
      >
        Contact support
      </Button>
    </div>
    <p className="text-text-muted mt-6 text-sm">
      We respond to support requests within two business days.
    </p>
  </MarketingHero>
);
