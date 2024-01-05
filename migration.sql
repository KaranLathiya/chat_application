-- public."users" definition

-- Drop table

-- DROP TABLE public."users";

CREATE TABLE public.users (
	id INT8 NOT NULL DEFAULT unique_rowid(),
	fullname VARCHAR(50) NOT NULL,
	email VARCHAR(320) NOT NULL,
	gender VARCHAR(6) NULL,
	ip_address VARCHAR(15) NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (id ASC)
);


-- public."groups" definition

-- Drop table

-- DROP TABLE public."groups";

CREATE TABLE public.groups (
	id INT8 NOT NULL DEFAULT unique_rowid(),
	group_name VARCHAR(30) NOT NULL,
	admin_id INT8 NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	CONSTRAINT groups_pk PRIMARY KEY (id ASC),
	CONSTRAINT groups_fk_admin_id FOREIGN KEY (admin_id) REFERENCES public.users(id)
);


-- public.personal_conversation definition

-- Drop table

-- DROP TABLE public.personal_conversation;

CREATE TABLE public.personal_conversations (
	sender_id INT8 NOT NULL,
	receiver_id INT8 NOT NULL,
	content VARCHAR(4096) NOT NULL,
	id INT8 NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	CONSTRAINT personal_conversation_pk PRIMARY KEY (id ASC),
	CONSTRAINT personal_conversation_fk_sender_id FOREIGN KEY (sender_id) REFERENCES public.users(id),
	CONSTRAINT personal_conversation_fk_receiver_id FOREIGN KEY (receiver_id) REFERENCES public.users(id)
);


-- public.group_members definition

-- Drop table

-- DROP TABLE public.group_members;

CREATE TABLE public.group_members (
	group_id INT8 NOT NULL,
	member_id INT8 NOT NULL,
	CONSTRAINT group_members_pk PRIMARY KEY (group_id ASC, member_id ASC),
	CONSTRAINT group_members_fk_member_id FOREIGN KEY (member_id) REFERENCES public.users(id),
	CONSTRAINT group_members_fk_group_id FOREIGN KEY (group_id) REFERENCES public.groups(id)
);


-- public.group_conversation definition

-- Drop table

-- DROP TABLE public.group_conversation;

CREATE TABLE public.group_conversations (
	group_id INT8 NOT NULL,
	sender_id INT8 NOT NULL,
	created_at TIMESTAMPTZ NOT NULL,
	content VARCHAR(4096) NOT NULL,
	id INT8 NOT NULL,
	CONSTRAINT group_conversation_pk PRIMARY KEY (id ASC),
	CONSTRAINT group_conversation_fk FOREIGN KEY (group_id, sender_id) REFERENCES public.group_members(group_id, member_id)
);