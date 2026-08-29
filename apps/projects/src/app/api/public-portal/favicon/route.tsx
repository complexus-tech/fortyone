import { ImageResponse } from "next/og";

const ICON_SIZE = 64;
const DOT_SIZE = 44;
const PRIMARY_COLOR = "#ea6060";

export const GET = () =>
  new ImageResponse(
    (
      <div
        style={{
          alignItems: "center",
          background: "transparent",
          display: "flex",
          height: "100%",
          justifyContent: "center",
          width: "100%",
        }}
      >
        <div
          style={{
            background: PRIMARY_COLOR,
            borderRadius: "999px",
            height: `${DOT_SIZE}px`,
            width: `${DOT_SIZE}px`,
          }}
        />
      </div>
    ),
    {
      height: ICON_SIZE,
      headers: {
        "Cache-Control": "public, max-age=31536000, immutable",
      },
      width: ICON_SIZE,
    },
  );
