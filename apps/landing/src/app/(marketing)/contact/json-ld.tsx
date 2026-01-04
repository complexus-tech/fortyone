import type { ContactPage, WithContext } from "schema-dts";

const contactPage: WithContext<ContactPage> = {
  "@context": "https://schema.org",
  "@type": "ContactPage",
  name: "Contact FortyOne",
  description:
    "Contact page for FortyOne - Modern OKR & Project Management Platform",
  mainEntity: {
    "@type": "Organization",
    name: "FortyOne",
    url: "https://fortyone.app",
    logo: "https://fortyone.app/images/logo.png",
    contactPoint: {
      "@type": "ContactPoint",
      contactType: "customer support",
      email: "hello@complexus.tech",
      availableLanguage: ["English"],
      contactOption: "TollFree",
      areaServed: "Worldwide",
    },
  },
};

export const ContactJsonLd = () => {
  return (
    <script
      dangerouslySetInnerHTML={{ __html: JSON.stringify(contactPage) }}
      type="application/ld+json"
    />
  );
};
