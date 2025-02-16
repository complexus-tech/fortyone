import type { ContactPage, WithContext } from "schema-dts";

const contactPage: WithContext<ContactPage> = {
  "@context": "https://schema.org",
  "@type": "ContactPage",
  name: "Contact Complexus",
  description:
    "Contact page for Complexus - Modern OKR & Project Management Platform",
  mainEntity: {
    "@type": "Organization",
    name: "Complexus",
    url: "https://complexus.app",
    logo: "https://complexus.app/images/logo.png",
    contactPoint: {
      "@type": "ContactPoint",
      contactType: "customer support",
      email: "support@complexus.app",
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
