import "./global.css";
import { RootProvider } from "fumadocs-ui/provider";
import { cn } from "lib";
import {
  Inter,
  Bricolage_Grotesque as BricolageGrotesque,
} from "next/font/google";
import type { ReactNode } from "react";
import Script from "next/script";

const inter = Inter({
  subsets: ["latin"],
});

const heading = BricolageGrotesque({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <html
      lang="en"
      className={cn(inter.className, heading.variable)}
      suppressHydrationWarning
    >
      <Script id="brevo-conversations" strategy="afterInteractive">
        {`
        (function(d, w, c) {
            w.BrevoConversationsID = '6834856b58b6d2f7800e0e5e';
            w[c] = w[c] || function() {
                (w[c].q = w[c].q || []).push(arguments);
            };
            var s = d.createElement('script');
            s.async = true;
            s.src = 'https://conversations-widget.brevo.com/brevo-conversations.js';
            if (d.head) d.head.appendChild(s);
        })(document, window, 'BrevoConversations');
      `}
      </Script>
      <body className="flex flex-col min-h-screen antialiased">
        <RootProvider>{children}</RootProvider>
      </body>
    </html>
  );
}
