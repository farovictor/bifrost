CREATE TABLE IF NOT EXISTS org_memberships (
    user_id VARCHAR(255) NOT NULL,
    org_id VARCHAR(255) NOT NULL REFERENCES organizations(id),
    role VARCHAR(255) NOT NULL,
    PRIMARY KEY (user_id, org_id)
);
