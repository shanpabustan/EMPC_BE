--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2026-04-18 20:44:18

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 237 (class 1255 OID 25171)
-- Name: get_nested_navigation(integer); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.get_nested_navigation(p_role_id integer) RETURNS jsonb
    LANGUAGE plpgsql
    AS $$
DECLARE
    result jsonb;
BEGIN
    WITH RECURSIVE nav_ids AS (
        -- 1. Start with nodes the user has explicit 'view' access to
        SELECT navigation_id AS id 
        FROM role_navigation_access 
        WHERE role_id = p_role_id AND can_view = true
        
        UNION
        
        -- 2. Recursively find all parent nodes so the tree is complete
        SELECT n.parent_id 
        FROM sys_navigation n
        JOIN nav_ids ON n.id = nav_ids.id
        WHERE n.parent_id IS NOT NULL
    ),
    allowed_nodes AS (
        -- 3. Combine the navigation details with the specific role permissions
        SELECT 
            n.id, n.parent_id, n.label, n.slug, n.sort_order,
            COALESCE(ra.can_view, true) as can_view,
            COALESCE(ra.can_add, false) as can_add,
            COALESCE(ra.can_edit, false) as can_edit,
            COALESCE(ra.can_delete, false) as can_delete,
            COALESCE(ra.can_override, false) as can_override
        FROM sys_navigation n
        LEFT JOIN role_navigation_access ra ON n.id = ra.navigation_id AND ra.role_id = p_role_id
        WHERE n.id IN (SELECT id FROM nav_ids)
    )
    -- 4. Build the JSON structure
    SELECT jsonb_agg(l1_nodes) INTO result FROM (
        SELECT jsonb_build_object(
            'label', l1.label,
            'slug', l1.slug,
            'children', COALESCE((
                SELECT jsonb_agg(l2_nodes) FROM (
                    SELECT jsonb_build_object(
                        'label', l2.label,
                        'slug', l2.slug,
                        'can_add', l2.can_add,
                        'can_edit', l2.can_edit,
                        'can_override', l2.can_override,
                        'children', COALESCE((
                            SELECT jsonb_agg(l3_nodes) FROM (
                                SELECT jsonb_build_object(
                                    'label', l3.label,
                                    'slug', l3.slug,
                                    'can_add', l3.can_add,
                                    'can_edit', l3.can_edit,
                                    'can_override', l3.can_override
                                ) AS l3_nodes
                                FROM allowed_nodes l3 
                                WHERE l3.parent_id = l2.id 
                                ORDER BY l3.sort_order
                            )
                        ), '[]'::jsonb)
                    ) AS l2_nodes
                    FROM allowed_nodes l2 
                    WHERE l2.parent_id = l1.id 
                    ORDER BY l2.sort_order
                )
            ), '[]'::jsonb)
        ) AS l1_nodes
        FROM allowed_nodes l1 
        WHERE l1.parent_id IS NULL 
        ORDER BY l1.sort_order
    ) sub;

    RETURN COALESCE(result, '[]'::jsonb);
END;
$$;


ALTER FUNCTION public.get_nested_navigation(p_role_id integer) OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 221 (class 1259 OID 25114)
-- Name: role_navigation_access; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.role_navigation_access (
    role_id integer NOT NULL,
    navigation_id integer NOT NULL,
    can_view boolean DEFAULT false,
    can_add boolean DEFAULT false,
    can_edit boolean DEFAULT false,
    can_delete boolean DEFAULT false,
    can_override boolean DEFAULT false
);


ALTER TABLE public.role_navigation_access OWNER TO postgres;

--
-- TOC entry 220 (class 1259 OID 25101)
-- Name: sys_module; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sys_module (
    id integer NOT NULL,
    parent_id integer,
    label character varying(100) NOT NULL,
    slug character varying(100),
    sort_order integer DEFAULT 0
);


ALTER TABLE public.sys_module OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 25100)
-- Name: sys_module_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sys_module_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sys_module_id_seq OWNER TO postgres;

--
-- TOC entry 4858 (class 0 OID 0)
-- Dependencies: 219
-- Name: sys_module_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sys_module_id_seq OWNED BY public.sys_module.id;


--
-- TOC entry 218 (class 1259 OID 25088)
-- Name: sys_navigation; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sys_navigation (
    id integer NOT NULL,
    parent_id integer,
    label character varying(100) NOT NULL,
    slug character varying(100),
    sort_order integer DEFAULT 0
);


ALTER TABLE public.sys_navigation OWNER TO postgres;

--
-- TOC entry 217 (class 1259 OID 25087)
-- Name: sys_navigation_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sys_navigation_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sys_navigation_id_seq OWNER TO postgres;

--
-- TOC entry 4859 (class 0 OID 0)
-- Dependencies: 217
-- Name: sys_navigation_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sys_navigation_id_seq OWNED BY public.sys_navigation.id;


--
-- TOC entry 223 (class 1259 OID 25131)
-- Name: sys_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.sys_roles (
    id integer NOT NULL,
    role_name character varying(100) NOT NULL,
    description text,
    created_at timestamp without time zone DEFAULT now()
);


ALTER TABLE public.sys_roles OWNER TO postgres;

--
-- TOC entry 222 (class 1259 OID 25130)
-- Name: sys_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.sys_roles_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.sys_roles_id_seq OWNER TO postgres;

--
-- TOC entry 4860 (class 0 OID 0)
-- Dependencies: 222
-- Name: sys_roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.sys_roles_id_seq OWNED BY public.sys_roles.id;


--
-- TOC entry 225 (class 1259 OID 25143)
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id integer NOT NULL,
    username character varying(255),
    staff_id character varying(50) NOT NULL,
    first_name character varying(100) NOT NULL,
    middle_name character varying(100),
    last_name character varying(100) NOT NULL,
    email character varying(255),
    phone_no character varying(20),
    birthdate date,
    password character varying(255),
    institution_id integer NOT NULL,
    institution_code character varying(50) NOT NULL,
    institution_name character varying(255),
    is_active boolean DEFAULT false,
    requires_password_reset boolean DEFAULT true,
    last_login timestamp without time zone,
    last_password_reset timestamp without time zone,
    role_id integer,
    created_at timestamp without time zone DEFAULT now() NOT NULL,
    updated_at timestamp without time zone DEFAULT now() NOT NULL,
    deleted_at timestamp without time zone
);


ALTER TABLE public.users OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 25142)
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.users_id_seq OWNER TO postgres;

--
-- TOC entry 4861 (class 0 OID 0)
-- Dependencies: 224
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- TOC entry 4663 (class 2604 OID 25104)
-- Name: sys_module id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_module ALTER COLUMN id SET DEFAULT nextval('public.sys_module_id_seq'::regclass);


--
-- TOC entry 4661 (class 2604 OID 25091)
-- Name: sys_navigation id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_navigation ALTER COLUMN id SET DEFAULT nextval('public.sys_navigation_id_seq'::regclass);


--
-- TOC entry 4670 (class 2604 OID 25134)
-- Name: sys_roles id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_roles ALTER COLUMN id SET DEFAULT nextval('public.sys_roles_id_seq'::regclass);


--
-- TOC entry 4672 (class 2604 OID 25146)
-- Name: users id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- TOC entry 4848 (class 0 OID 25114)
-- Dependencies: 221
-- Data for Name: role_navigation_access; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.role_navigation_access (role_id, navigation_id, can_view, can_add, can_edit, can_delete, can_override) FROM stdin;
3	11	t	t	t	f	f
3	12	t	t	t	f	f
3	14	t	t	t	f	f
3	21	t	t	t	f	f
3	421	t	t	t	f	f
3	51	t	f	f	f	f
3	52	t	f	f	f	f
3	53	t	f	f	f	f
3	54	t	f	f	f	f
3	1	t	f	f	f	f
3	2	t	f	f	f	f
3	4	t	f	f	f	f
3	42	t	f	f	f	f
3	5	t	f	f	f	f
1	1	t	t	t	t	t
1	2	t	t	t	t	t
1	3	t	t	t	t	t
1	4	t	t	t	t	t
1	5	t	t	t	t	t
1	11	t	t	t	t	t
1	12	t	t	t	t	t
1	13	t	t	t	t	t
1	14	t	t	t	t	t
1	21	t	t	t	t	t
1	22	t	t	t	t	t
1	23	t	t	t	t	t
1	31	t	t	t	t	t
1	32	t	t	t	t	t
1	33	t	t	t	t	t
1	34	t	t	t	t	t
1	41	t	t	t	t	t
1	42	t	t	t	t	t
1	43	t	t	t	t	t
1	51	t	t	t	t	t
1	52	t	t	t	t	t
1	53	t	t	t	t	t
1	54	t	t	t	t	t
1	55	t	t	t	t	t
1	421	t	t	t	t	t
1	422	t	t	t	t	t
1	423	t	t	t	t	t
\.


--
-- TOC entry 4847 (class 0 OID 25101)
-- Dependencies: 220
-- Data for Name: sys_module; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sys_module (id, parent_id, label, slug, sort_order) FROM stdin;
\.


--
-- TOC entry 4845 (class 0 OID 25088)
-- Dependencies: 218
-- Data for Name: sys_navigation; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sys_navigation (id, parent_id, label, slug, sort_order) FROM stdin;
1	\N	FILE	\N	1
2	\N	PAYMENT	\N	2
3	\N	UTILITIES	\N	3
4	\N	MAINTENANCE	\N	4
5	\N	REPORTS	\N	5
11	1	CLIENT	client	1
12	1	LOAN	loan	2
13	1	HEALTH	health	3
14	1	MSP	msp	4
21	2	SINGLE PAYMENT	single-payment	1
22	2	PEOPLECORE REMITTANCE	peoplecore-remittance	2
23	2	STAFF HEALTH PREMIUM	staff-health-premium	3
31	3	ADMINISTRATIVE TOOL	admin-tool	1
32	3	AUDIT TRAIL	audit-trail	2
33	3	CSP CHARGE RATE	csp-charge-rate	3
34	3	HMO MANAGEMENT	hmo-management	4
41	4	CHANGE PASSWORD	change-password	1
42	4	PARAMETERS	\N	2
43	4	USER ACCESS	user-access	3
51	5	CLIENTS	reports-clients	1
52	5	LOAN RELATED	reports-loan	2
53	5	HEALTH RELATED	reports-health	3
54	5	MSP	reports-msp	4
55	5	OTHER REPORTS	reports-other	5
421	42	SAVINGS ACCOUNT	savings-account	1
422	42	LOAN ACCOUNTS	loan-accounts	2
423	42	PROCESSING FEE	processing-fee	3
\.


--
-- TOC entry 4850 (class 0 OID 25131)
-- Dependencies: 223
-- Data for Name: sys_roles; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.sys_roles (id, role_name, description, created_at) FROM stdin;
1	System Admin	Full Administrative Access	2026-04-18 19:48:27.97749
2	General Manager	Overall Management with Overrides	2026-04-18 19:48:27.97749
3	Asst. General Manager	Compliance and Management	2026-04-18 19:48:27.97749
4	Finance Manager-MERP	MERP Financial Oversight	2026-04-18 19:48:27.97749
5	Finance Officer-MERP	MERP Financial Operations	2026-04-18 19:48:27.97749
6	Sr. Finance Officer - Taxation	Finance and Taxation Specialist	2026-04-18 19:48:27.97749
7	Finance Officer	Standard Finance Operations	2026-04-18 19:48:27.97749
8	Bookkeeper	General Ledger and Admin Tools	2026-04-18 19:48:27.97749
9	Finance Asst. 1	Finance Support Level 1	2026-04-18 19:48:27.97749
10	Finance Asst. 2	Finance Support Level 2	2026-04-18 19:48:27.97749
11	Loan Manager	Loan Operations and Assessment	2026-04-18 19:48:27.97749
12	Admin Manager-Health	Health Department Administration	2026-04-18 19:48:27.97749
13	Admin Officer-Health	Health Department Operations	2026-04-18 19:48:27.97749
14	Admin Manager-Compliance	Compliance Department Management	2026-04-18 19:48:27.97749
15	Loan Manager 1	Finance Department Loan Operations	2026-04-18 19:48:27.97749
16	Loan Manager 2	Operations Department Loan Assessment	2026-04-18 19:48:27.97749
\.


--
-- TOC entry 4852 (class 0 OID 25143)
-- Dependencies: 225
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, username, staff_id, first_name, middle_name, last_name, email, phone_no, birthdate, password, institution_id, institution_code, institution_name, is_active, requires_password_reset, last_login, last_password_reset, role_id, created_at, updated_at, deleted_at) FROM stdin;
1	sr_finance_tax	EMPC-TAX-001	Juan	\N	Dela Cruz	\N	\N	\N	\N	101	HO	\N	t	t	\N	\N	6	2026-04-18 19:49:08.199861	2026-04-18 19:49:08.199861	\N
\.


--
-- TOC entry 4862 (class 0 OID 0)
-- Dependencies: 219
-- Name: sys_module_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sys_module_id_seq', 1, false);


--
-- TOC entry 4863 (class 0 OID 0)
-- Dependencies: 217
-- Name: sys_navigation_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sys_navigation_id_seq', 1, false);


--
-- TOC entry 4864 (class 0 OID 0)
-- Dependencies: 222
-- Name: sys_roles_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.sys_roles_id_seq', 17, false);


--
-- TOC entry 4865 (class 0 OID 0)
-- Dependencies: 224
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.users_id_seq', 1, true);


--
-- TOC entry 4682 (class 2606 OID 25123)
-- Name: role_navigation_access role_navigation_access_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_navigation_access
    ADD CONSTRAINT role_navigation_access_pkey PRIMARY KEY (role_id, navigation_id);


--
-- TOC entry 4680 (class 2606 OID 25107)
-- Name: sys_module sys_module_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_module
    ADD CONSTRAINT sys_module_pkey PRIMARY KEY (id);


--
-- TOC entry 4678 (class 2606 OID 25094)
-- Name: sys_navigation sys_navigation_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_navigation
    ADD CONSTRAINT sys_navigation_pkey PRIMARY KEY (id);


--
-- TOC entry 4684 (class 2606 OID 25139)
-- Name: sys_roles sys_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_roles
    ADD CONSTRAINT sys_roles_pkey PRIMARY KEY (id);


--
-- TOC entry 4686 (class 2606 OID 25141)
-- Name: sys_roles sys_roles_role_name_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_roles
    ADD CONSTRAINT sys_roles_role_name_key UNIQUE (role_name);


--
-- TOC entry 4688 (class 2606 OID 25160)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4690 (class 2606 OID 25154)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- TOC entry 4692 (class 2606 OID 25158)
-- Name: users users_staff_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_staff_id_key UNIQUE (staff_id);


--
-- TOC entry 4694 (class 2606 OID 25156)
-- Name: users users_username_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_username_key UNIQUE (username);


--
-- TOC entry 4697 (class 2606 OID 25124)
-- Name: role_navigation_access role_navigation_access_navigation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.role_navigation_access
    ADD CONSTRAINT role_navigation_access_navigation_id_fkey FOREIGN KEY (navigation_id) REFERENCES public.sys_navigation(id) ON DELETE CASCADE;


--
-- TOC entry 4696 (class 2606 OID 25108)
-- Name: sys_module sys_module_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_module
    ADD CONSTRAINT sys_module_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.sys_module(id) ON DELETE CASCADE;


--
-- TOC entry 4695 (class 2606 OID 25095)
-- Name: sys_navigation sys_navigation_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.sys_navigation
    ADD CONSTRAINT sys_navigation_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.sys_navigation(id) ON DELETE CASCADE;


--
-- TOC entry 4698 (class 2606 OID 25161)
-- Name: users users_role_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.sys_roles(id) ON DELETE SET NULL;


-- Completed on 2026-04-18 20:44:18

--
-- PostgreSQL database dump complete
--

