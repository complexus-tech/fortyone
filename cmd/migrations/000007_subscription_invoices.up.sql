-- 000007_subscription_invoices.up.sql
CREATE TABLE public.subscription_invoices (
    invoice_id integer NOT NULL,
    workspace_id uuid NOT NULL,
    stripe_invoice_id text NOT NULL,
    amount_paid numeric(10,2) NOT NULL,
    invoice_date timestamp with time zone NOT NULL,
    status text NOT NULL,
    seats_count integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    hosted_url character varying DEFAULT ''::character varying,
    customer_name character varying
);

CREATE SEQUENCE public.subscription_invoices_invoice_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER TABLE ONLY public.subscription_invoices ADD CONSTRAINT subscription_invoices_pkey PRIMARY KEY (invoice_id);

ALTER TABLE ONLY public.subscription_invoices 
    ADD CONSTRAINT subscription_invoices_workspace_id_fkey FOREIGN KEY (workspace_id) REFERENCES public.workspaces(workspace_id);

ALTER TABLE ONLY public.subscription_invoices ALTER COLUMN invoice_id SET DEFAULT nextval('public.subscription_invoices_invoice_id_seq'::regclass);

CREATE INDEX idx_subscription_invoices_workspace ON public.subscription_invoices USING btree (workspace_id);
