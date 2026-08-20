ALTER TABLE public.objectives
    DROP CONSTRAINT IF EXISTS objectives_color_hex_check,
    DROP COLUMN IF EXISTS color;

