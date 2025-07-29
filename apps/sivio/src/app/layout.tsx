import { type ReactNode } from "react";
import { cn } from "lib";
import { instrumentSans } from "@/styles/fonts";
import "../styles/global.css";

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={cn(instrumentSans.className)}>{children}</body>
    </html>
  );
}
