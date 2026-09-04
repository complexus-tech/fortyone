import { EmptyStateIllustrationFrame } from "./empty-state-illustration-frame";

const WindowHeader = () => (
  <>
    <path d="M46 61H234" stroke="currentColor" strokeOpacity="0.18" />
    <circle cx="61" cy="48" fill="currentColor" opacity="0.45" r="3" />
    <circle cx="72" cy="48" fill="currentColor" opacity="0.22" r="3" />
    <circle cx="83" cy="48" fill="currentColor" opacity="0.12" r="3" />
  </>
);

const Sparkles = () => (
  <>
    <path
      d="M236 25L238.2 30.8L244 33L238.2 35.2L236 41L233.8 35.2L228 33L233.8 30.8L236 25Z"
      fill="currentColor"
      opacity="0.55"
    />
    <path
      d="M37 119L38.5 123L42.5 124.5L38.5 126L37 130L35.5 126L31.5 124.5L35.5 123L37 119Z"
      fill="currentColor"
      opacity="0.28"
    />
  </>
);

export const WorkEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    {[76, 101, 126].map((y, index) => (
      <g key={y} opacity={1 - index * 0.22}>
        <rect
          fill="currentColor"
          fillOpacity="0.035"
          height="19"
          rx="5"
          stroke="currentColor"
          strokeOpacity="0.2"
          width="148"
          x="66"
          y={y}
        />
        <rect
          height="7"
          rx="2"
          stroke="currentColor"
          strokeOpacity="0.4"
          width="7"
          x="76"
          y={y + 6}
        />
        <path
          d={`M94 ${y + 9.5}H${158 - index * 12}`}
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.3"
          strokeWidth="2"
        />
        <circle
          cx="201"
          cy={y + 9.5}
          fill="currentColor"
          opacity="0.18"
          r="3"
        />
      </g>
    ))}
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const ActivityEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M76 83V135"
      stroke="currentColor"
      strokeDasharray="3 5"
      strokeOpacity="0.2"
      strokeWidth="1.5"
    />
    {[82, 107, 132].map((y, index) => (
      <g key={y} opacity={1 - index * 0.22}>
        <circle cx="76" cy={y} fill="currentColor" opacity="0.12" r="7" />
        <circle cx="76" cy={y} fill="currentColor" opacity="0.4" r="2.5" />
        <path
          d={`M92 ${y - 3}H${145 - index * 7}M92 ${y + 4}H${127 - index * 7}`}
          stroke="currentColor"
          strokeLinecap="round"
          strokeOpacity="0.28"
          strokeWidth="2"
        />
      </g>
    ))}
    <circle
      cx="189"
      cy="108"
      fill="currentColor"
      fillOpacity="0.05"
      r="25"
      stroke="currentColor"
      strokeOpacity="0.32"
      strokeWidth="1.5"
    />
    <path
      d="M189 88V91M209 108H206M189 128V125M169 108H172"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.2"
      strokeWidth="1.5"
    />
    <path
      d="M189 95V108L198 114"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.5"
      strokeWidth="2"
    />
    <circle cx="189" cy="108" fill="currentColor" opacity="0.5" r="2" />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const RoadmapEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M67 127C91 106 108 111 128 91C148 71 171 75 207 86"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.32"
      strokeWidth="1.5"
    />
    <path
      d="M68 76V137M116 76V137M164 76V137M212 76V137"
      stroke="currentColor"
      strokeDasharray="3 5"
      strokeOpacity="0.11"
    />
    <circle cx="82" cy="116" fill="currentColor" opacity="0.16" r="7" />
    <circle cx="128" cy="91" fill="currentColor" opacity="0.2" r="8" />
    <circle cx="177" cy="77" fill="currentColor" opacity="0.14" r="6" />
    <path
      d="M128 84V98M121 91H135"
      stroke="currentColor"
      strokeLinecap="round"
      strokeWidth="1.75"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const NotificationsEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    {[80, 104, 128].map((y, index) => (
      <g key={y} opacity={0.34 - index * 0.08}>
        <circle cx="70" cy={y} fill="currentColor" r="5" />
        <path
          d={`M82 ${y - 3}H151M82 ${y + 3}H126`}
          stroke="currentColor"
          strokeLinecap="round"
          strokeWidth="2"
        />
      </g>
    ))}
    <path
      d="M185 111V96C185 86.6 192.6 79 202 79C211.4 79 219 86.6 219 96V111L224 119H180L185 111Z"
      fill="currentColor"
      fillOpacity="0.08"
      stroke="currentColor"
      strokeOpacity="0.38"
      strokeWidth="1.5"
    />
    <path
      d="M196 124C197.5 127 200 128.5 202 128.5C204 128.5 206.5 127 208 124"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.38"
      strokeWidth="1.5"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const DocumentsEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M91 82H151L167 98V142H91V82Z"
      fill="currentColor"
      fillOpacity="0.05"
      stroke="currentColor"
      strokeOpacity="0.3"
      strokeWidth="1.5"
    />
    <path
      d="M151 82V98H167"
      stroke="currentColor"
      strokeOpacity="0.3"
      strokeWidth="1.5"
    />
    <path
      d="M105 108H151M105 119H145M105 130H134"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.3"
      strokeWidth="2"
    />
    <path
      d="M175 75H201L210 84V120H179"
      stroke="currentColor"
      strokeDasharray="4 5"
      strokeOpacity="0.16"
      strokeWidth="1.5"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const SprintsEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M99 93C107 80 122 72 140 72C163 72 182 88 186 109"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.34"
      strokeWidth="2"
    />
    <path
      d="M181 101L186 109L192 101"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.34"
      strokeWidth="2"
    />
    <path
      d="M181 123C172 136 158 143 140 143C117 143 98 128 94 107"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.2"
      strokeWidth="2"
    />
    <path
      d="M99 115L94 107L88 115"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.2"
      strokeWidth="2"
    />
    <path
      d="M119 98H161V122H119V98Z"
      fill="currentColor"
      fillOpacity="0.08"
      stroke="currentColor"
      strokeOpacity="0.28"
      strokeWidth="1.5"
    />
    <path
      d="M130 98V122M150 98V122"
      stroke="currentColor"
      strokeOpacity="0.14"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const SearchEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M74 79H149M74 96H130M74 113H118"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.18"
      strokeWidth="3"
    />
    <circle
      cx="178"
      cy="105"
      fill="currentColor"
      fillOpacity="0.06"
      r="24"
      stroke="currentColor"
      strokeOpacity="0.36"
      strokeWidth="1.5"
    />
    <path
      d="M196 123L211 138"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.42"
      strokeWidth="3"
    />
    <path
      d="M170 105H186"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.36"
      strokeWidth="2"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const InboxEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M86 91H194L207 136H73L86 91Z"
      fill="currentColor"
      fillOpacity="0.05"
      stroke="currentColor"
      strokeLinejoin="round"
      strokeOpacity="0.3"
      strokeWidth="1.5"
    />
    <path
      d="M75 120H112L120 130H160L168 120H205"
      stroke="currentColor"
      strokeOpacity="0.3"
      strokeWidth="1.5"
    />
    <path
      d="M105 78H175M117 68H163"
      stroke="currentColor"
      strokeDasharray="4 5"
      strokeLinecap="round"
      strokeOpacity="0.18"
      strokeWidth="2"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const TeamEmptyIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M104 126C104 111 116 100 131 100H149C164 100 176 111 176 126"
      fill="currentColor"
      fillOpacity="0.05"
      stroke="currentColor"
      strokeOpacity="0.3"
      strokeWidth="1.5"
    />
    <circle
      cx="140"
      cy="84"
      fill="currentColor"
      fillOpacity="0.08"
      r="15"
      stroke="currentColor"
      strokeOpacity="0.34"
      strokeWidth="1.5"
    />
    <circle cx="95" cy="98" fill="currentColor" fillOpacity="0.12" r="9" />
    <circle cx="185" cy="98" fill="currentColor" fillOpacity="0.12" r="9" />
    <path
      d="M75 129C76 117 84 110 95 110C103 110 109 114 112 120M205 129C204 117 196 110 185 110C177 110 171 114 168 120"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.2"
      strokeWidth="1.5"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const NotFoundIllustration = ({ className }: { className?: string }) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M78 125L105 91L130 112L158 78L202 125"
      stroke="currentColor"
      strokeDasharray="5 6"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeOpacity="0.28"
      strokeWidth="2"
    />
    <circle cx="158" cy="78" fill="currentColor" opacity="0.18" r="8" />
    <path
      d="M151 71L165 85M165 71L151 85"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.56"
      strokeWidth="2"
    />
    <path
      d="M78 125H202"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.14"
      strokeWidth="1.5"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);

export const UnauthorizedIllustration = ({
  className,
}: {
  className?: string;
}) => (
  <EmptyStateIllustrationFrame className={className}>
    <WindowHeader />
    <path
      d="M112 99V89C112 73.5 124.5 61 140 61C155.5 61 168 73.5 168 89V99"
      stroke="currentColor"
      strokeOpacity="0.34"
      strokeWidth="2"
    />
    <path
      d="M101 99H179V141H101V99Z"
      fill="currentColor"
      fillOpacity="0.06"
      stroke="currentColor"
      strokeOpacity="0.34"
      strokeWidth="1.5"
    />
    <circle cx="140" cy="118" fill="currentColor" opacity="0.34" r="4" />
    <path
      d="M140 122V130"
      stroke="currentColor"
      strokeLinecap="round"
      strokeOpacity="0.34"
      strokeWidth="2"
    />
    <Sparkles />
  </EmptyStateIllustrationFrame>
);
