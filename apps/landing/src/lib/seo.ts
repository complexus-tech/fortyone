export const SITE_URL = "https://www.fortyone.app";

export const HOME_METADATA_TITLE =
  "Strategy, Feedback & AI Project Management | FortyOne";
export const HOME_METADATA_DESCRIPTION =
  "Connect company strategy and customer feedback to project plans your team can deliver, with AI support for ownership, scheduling, and delivery risk.";

export const DEFAULT_SOCIAL_IMAGE = {
  url: "/opengraph-image.png",
  width: 1200,
  height: 630,
  alt: "FortyOne AI project management platform for strategy, feedback, and delivery",
};

export const DEFAULT_TWITTER_IMAGE = {
  url: "/twitter-image.png",
  width: 1200,
  height: 630,
  alt: "FortyOne AI project management platform for strategy, feedback, and delivery",
};

export const getCanonicalUrl = (pathname: string) =>
  new URL(pathname, SITE_URL).toString();
