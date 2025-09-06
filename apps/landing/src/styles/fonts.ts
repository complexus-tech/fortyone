import {
  Inter_Tight as Inter,
  Inconsolata,
  Instrument_Sans as InstrumentSans,
} from "next/font/google";

export const body = Inter({
  variable: "--font-body",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const mono = Inconsolata({
  variable: "--font-mono",
  display: "swap",
  subsets: ["latin"],
});

export const heading = InstrumentSans({
  variable: "--font-heading",
  display: "swap",
  subsets: ["latin"],
  weight: "variable",
});
