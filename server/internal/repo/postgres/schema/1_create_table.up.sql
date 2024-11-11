CREATE TABLE space (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    UNIQUE (name)
);

CREATE TABLE "user"(
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    email VARCHAR(128) NOT NULL,
    pub_key TEXT NOT NULL,
    ban BOOLEAN,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    UNIQUE (email)
);

CREATE TABLE user_cert(
    id SERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    cert TEXT NOT NULL,
    expires_at BIGINT NOT NULL,
    is_revoked BOOLEAN NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT
);

CREATE INDEX idx_user_cert_user_id ON user_cert(user_id);

CREATE TABLE node (
    id SERIAL PRIMARY KEY,
    unique_id VARCHAR(64) NOT NULL,
    secret VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    space_id BIGINT NOT NULL,
    ip INET NOT NULL,
    last_heartbeat BIGINT,
    accounts TEXT[] NOT NULL,
    created_at BIGINT NOT NULL,
    updated_at BIGINT,
    UNIQUE (unique_id)
);

CREATE INDEX idx_node_space_id ON node(space_id);

CREATE TABLE role (
    id SERIAL PRIMARY KEY,
    space_id BIGINT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL,
    UNIQUE (name)
);

CREATE TABLE role_node (
    id SERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    node_id BIGINT NOT NULL,
    account VARCHAR(120) NOT NULL,
    created_at BIGINT NOT NULL,
    UNIQUE (role_id, node_id)
);

CREATE TABLE role_user(
    id SERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    created_at BIGINT NOT NULL,
    UNIQUE (role_id, user_id)
);



-- 添加注释
COMMENT ON COLUMN node.unique_id IS 'Unique identifier for the node';
COMMENT ON COLUMN node.secret IS 'Secret for the node';
COMMENT ON COLUMN node.name IS 'Name of the node';
COMMENT ON COLUMN node.description IS 'Description of the node';
COMMENT ON COLUMN node.space_id IS 'Space ID';
COMMENT ON COLUMN node.ip IS 'IP address of the node';
COMMENT ON COLUMN node.last_heartbeat IS 'Last heartbeat of the node';
COMMENT ON COLUMN node.account IS 'Account of the node, use "," to separate multiple accounts';
COMMENT ON COLUMN node.created_at IS 'Creation time';
COMMENT ON COLUMN node.updated_at IS 'Last update time';

COMMENT ON COLUMN role.space_id IS 'Space ID';
COMMENT ON COLUMN role.name IS 'Name of the role';
COMMENT ON COLUMN role.description IS 'Description of the role';
COMMENT ON COLUMN role.created_at IS 'Creation time';

COMMENT ON COLUMN role_node.role_id IS 'Role ID';
COMMENT ON COLUMN role_node.node_id IS 'Node ID';
COMMENT ON COLUMN role_node.account IS 'Account of the node';
COMMENT ON COLUMN role_node.created_at IS 'Creation time';

COMMENT ON COLUMN role_user.role_id IS 'Role ID';
COMMENT ON COLUMN role_user.user_id IS 'User ID';
COMMENT ON COLUMN role_user.created_at IS 'Creation time';

COMMENT ON COLUMN `user`.username IS 'Username';
COMMENT ON COLUMN `user`.email IS 'Email';
COMMENT ON COLUMN `user`.pub_key IS 'Public key';
COMMENT ON COLUMN `user`.created_at IS 'Creation time';
COMMENT ON COLUMN `user`.updated_at IS 'Last update time';