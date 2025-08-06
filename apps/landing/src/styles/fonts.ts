import {
  // Instrument_Sans as InstrumentSans,
  Englebert,
  Inter,
} from "next/font/google";

export const instrumentSans = Inter({
  variable: "--font-body",
  subsets: ["latin"],
  display: "swap",
  weight: "variable",
});

export const heading = Englebert({
  variable: "--font-heading",
  display: "swap",
  weight: "400",
  style: "normal",
});
