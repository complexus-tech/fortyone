import { type ReactNode } from "react";
import { cn } from "lib";
import { ibmplexsans } from "@/styles/fonts";
import "../styles/global.css";
import { Navigation } from "@/components/shared/navigation";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={cn(ibmplexsans.className)}>
        <Navigation />
        {children}
      </body>
    </html>
  );
}
