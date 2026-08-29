ALTER TABLE public.objectives
    ADD COLUMN color varchar(7);

UPDATE public.objectives
SET color = (ARRAY[
    '#4A90E2',
    '#686DE0',
    '#9B59B6',
    '#4ECDC4',
    '#E67E22',
    '#6BCB77',
    '#FF6B6B',
    '#95A5A6'
])[MOD(sequence_id - 1, 8) + 1];

ALTER TABLE public.objectives
    ALTER COLUMN color SET DEFAULT '#4A90E2',
    ALTER COLUMN color SET NOT NULL,
    ADD CONSTRAINT objectives_color_hex_check
        CHECK (color ~ '^#[0-9A-Fa-f]{6}$');

