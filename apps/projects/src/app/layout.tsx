import type { Metadata } from "next";
import { Inter } from "next/font/google";
import type { ReactNode } from "react";
import "../styles/global.css";
import { Wrapper } from "./wrapper";

const inter = Inter({
  subsets: ["latin"],
  display: "swap",
  weight: ["400", "500", "600", "700"],
});

export const metadata: Metadata = {
  title: "Complexus Projects",
  description: "Complexus Projects",
};

export default function RootLayout({
  children,
}: {
  children: ReactNode;
}): JSX.Element {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <Wrapper>{children}</Wrapper>
      </body>
    </html>
  );
}
