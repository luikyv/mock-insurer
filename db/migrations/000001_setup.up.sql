CREATE TABLE mock_users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL,
    name TEXT NOT NULL,
    cpf TEXT NOT NULL,
    cnpj TEXT,
    description TEXT,

    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
CREATE INDEX idx_mock_users_org_id ON mock_users (org_id);
CREATE UNIQUE INDEX idx_mock_users_org_id_cpf ON mock_users (org_id, cpf);
CREATE UNIQUE INDEX idx_mock_users_org_id_cnpj ON mock_users (org_id, cnpj);
CREATE UNIQUE INDEX idx_mock_users_org_id_username ON mock_users (org_id, username);

-- mock_user_business associates individual users with business users (i.e., users that own a CNPJ).
CREATE TABLE mock_user_business (
    user_id UUID NOT NULL REFERENCES mock_users(id) ON DELETE CASCADE,
    business_user_id UUID NOT NULL REFERENCES mock_users(id) ON DELETE CASCADE,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    PRIMARY KEY (user_id, business_user_id)
);
CREATE INDEX idx_mock_user_business_org_id ON mock_user_business (org_id);

CREATE TABLE oauth_clients (
    id TEXT PRIMARY KEY,
    data JSONB NOT NULL,
    name TEXT,
    webhook_uris JSONB,
    origin_uris JSONB,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
CREATE INDEX idx_oauth_clients_org_id ON oauth_clients (org_id);

CREATE TABLE oauth_sessions (
    id TEXT PRIMARY KEY,
    callback_id TEXT,
    auth_code TEXT,
    pushed_auth_req_id TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    data JSONB NOT NULL,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
CREATE INDEX idx_oauth_sessions_org_id ON oauth_sessions (org_id);
CREATE INDEX idx_oauth_sessions_callback_id ON oauth_sessions (callback_id);
CREATE INDEX idx_oauth_sessions_auth_code ON oauth_sessions (auth_code);
CREATE INDEX idx_oauth_sessions_pushed_auth_req_id ON oauth_sessions (pushed_auth_req_id);

CREATE TABLE oauth_grants (
    id TEXT PRIMARY KEY,
    token_id TEXT NOT NULL,
    refresh_token TEXT,
    auth_code TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    data JSONB,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
CREATE INDEX idx_oauth_grants_org_id ON oauth_grants (org_id);
CREATE INDEX idx_oauth_grants_token_id ON oauth_grants (token_id);
CREATE INDEX idx_oauth_grants_refresh_token ON oauth_grants (refresh_token);
CREATE INDEX idx_oauth_grants_auth_code ON oauth_grants (auth_code);

CREATE TABLE consents (
    id UUID PRIMARY KEY,
    status TEXT NOT NULL,
    permissions JSONB NOT NULL,
    status_updated_at TIMESTAMPTZ DEFAULT now(),
    expires_at TIMESTAMPTZ,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    user_identification TEXT NOT NULL,
	user_rel TEXT NOT NULL,
    business_identification TEXT,
	business_rel TEXT,
    client_id TEXT NOT NULL REFERENCES oauth_clients(id) ON DELETE CASCADE,
    rejection JSONB,
    is_linked BOOLEAN,
    link_id TEXT,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_personal_identifications (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_personal_qualifications (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_personal_complimentary_informations (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_business_identifications (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_business_qualifications (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE customer_business_complimentary_informations (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES mock_users(id) NOT NULL,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_auto_policies (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_auto_policies (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL REFERENCES insurance_auto_policies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_auto_policies PRIMARY KEY (consent_id, policy_id)
);

CREATE TABLE insurance_auto_policy_claims (
    id UUID PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES insurance_auto_policies(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_capitalization_title_plans (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_capitalization_title_plans (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES insurance_capitalization_title_plans(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_capitalization_title_plans PRIMARY KEY (consent_id, plan_id)
);

CREATE TABLE insurance_capitalization_title_events (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL REFERENCES insurance_capitalization_title_plans(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_capitalization_title_settlements (
    id UUID PRIMARY KEY,
    plan_id UUID NOT NULL REFERENCES insurance_capitalization_title_plans(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_financial_assistance_contracts (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_financial_assistance_contracts (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    contract_id TEXT NOT NULL REFERENCES insurance_financial_assistance_contracts(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_financial_assistance_contracts PRIMARY KEY (consent_id, contract_id)
);

CREATE TABLE insurance_financial_assistance_movements (
    id UUID PRIMARY KEY,
    contract_id TEXT NOT NULL REFERENCES insurance_financial_assistance_contracts(id) ON DELETE CASCADE,
    movement_data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_acceptance_and_branches_abroad_policies (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_acceptance_and_branches_abroad_policies (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL REFERENCES insurance_acceptance_and_branches_abroad_policies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_acceptance_and_branches_abroad_policies PRIMARY KEY (consent_id, policy_id)
);

CREATE TABLE insurance_acceptance_and_branches_abroad_claims (
    id UUID PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES insurance_acceptance_and_branches_abroad_policies(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_financial_risk_policies (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_financial_risk_policies (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL REFERENCES insurance_financial_risk_policies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_financial_risk_policies PRIMARY KEY (consent_id, policy_id)
);

CREATE TABLE insurance_financial_risk_claims (
    id UUID PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES insurance_financial_risk_policies(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_housing_policies (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_housing_policies (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL REFERENCES insurance_housing_policies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_housing_policies PRIMARY KEY (consent_id, policy_id)
);

CREATE TABLE insurance_housing_claims (
    id UUID PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES insurance_housing_policies(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_life_pension_contracts (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_life_pension_contracts (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    contract_id TEXT NOT NULL REFERENCES insurance_life_pension_contracts(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_life_pension_contracts PRIMARY KEY (consent_id, contract_id)
);

CREATE TABLE insurance_life_pension_portabilities (
    id UUID PRIMARY KEY,
    contract_id TEXT NOT NULL REFERENCES insurance_life_pension_contracts(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_life_pension_withdrawals (
    id UUID PRIMARY KEY,
    contract_id TEXT NOT NULL REFERENCES insurance_life_pension_contracts(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_life_pension_claims (
    id UUID PRIMARY KEY,
    contract_id TEXT NOT NULL REFERENCES insurance_life_pension_contracts(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_patrimonial_policies (
    id TEXT PRIMARY KEY,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE consent_insurance_patrimonial_policies (
    consent_id UUID NOT NULL REFERENCES consents(id) ON DELETE CASCADE,
    policy_id TEXT NOT NULL REFERENCES insurance_patrimonial_policies(id) ON DELETE CASCADE,
    owner_id UUID NOT NULL REFERENCES mock_users(id),
    status TEXT NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL,

    CONSTRAINT pk_consent_insurance_patrimonial_policies PRIMARY KEY (consent_id, policy_id)
);

CREATE TABLE insurance_patrimonial_claims (
    id UUID PRIMARY KEY,
    policy_id TEXT NOT NULL REFERENCES insurance_patrimonial_policies(id) ON DELETE CASCADE,
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    cross_org BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE OR REPLACE VIEW consent_resources AS
    WITH authorised_consents AS (SELECT id, org_id FROM consents WHERE status = 'AUTHORISED')

    SELECT
        'DAMAGES_AND_PEOPLE_AUTO' AS resource_type,
        consent_insurance_auto_policies.consent_id,
        consent_insurance_auto_policies.policy_id AS resource_id,
        consent_insurance_auto_policies.owner_id,
        consent_insurance_auto_policies.status,
        consent_insurance_auto_policies.org_id,
        consent_insurance_auto_policies.created_at,
        consent_insurance_auto_policies.updated_at
    FROM consent_insurance_auto_policies
    JOIN authorised_consents ON consent_insurance_auto_policies.consent_id = authorised_consents.id AND consent_insurance_auto_policies.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'CAPITALIZATION_TITLES' AS resource_type,
        consent_insurance_capitalization_title_plans.consent_id,
        consent_insurance_capitalization_title_plans.plan_id::TEXT AS resource_id,
        consent_insurance_capitalization_title_plans.owner_id,
        consent_insurance_capitalization_title_plans.status,
        consent_insurance_capitalization_title_plans.org_id,
        consent_insurance_capitalization_title_plans.created_at,
        consent_insurance_capitalization_title_plans.updated_at
    FROM consent_insurance_capitalization_title_plans
    JOIN authorised_consents ON consent_insurance_capitalization_title_plans.consent_id = authorised_consents.id AND consent_insurance_capitalization_title_plans.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'FINANCIAL_ASSISTANCE' AS resource_type,
        consent_insurance_financial_assistance_contracts.consent_id,
        consent_insurance_financial_assistance_contracts.contract_id AS resource_id,
        consent_insurance_financial_assistance_contracts.owner_id,
        consent_insurance_financial_assistance_contracts.status,
        consent_insurance_financial_assistance_contracts.org_id,
        consent_insurance_financial_assistance_contracts.created_at,
        consent_insurance_financial_assistance_contracts.updated_at
    FROM consent_insurance_financial_assistance_contracts
    JOIN authorised_consents ON consent_insurance_financial_assistance_contracts.consent_id = authorised_consents.id AND consent_insurance_financial_assistance_contracts.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'DAMAGES_AND_PEOPLE_ACCEPTANCE_AND_BRANCHES_ABROAD' AS resource_type,
        consent_insurance_acceptance_and_branches_abroad_policies.consent_id,
        consent_insurance_acceptance_and_branches_abroad_policies.policy_id AS resource_id,
        consent_insurance_acceptance_and_branches_abroad_policies.owner_id,
        consent_insurance_acceptance_and_branches_abroad_policies.status,
        consent_insurance_acceptance_and_branches_abroad_policies.org_id,
        consent_insurance_acceptance_and_branches_abroad_policies.created_at,
        consent_insurance_acceptance_and_branches_abroad_policies.updated_at
    FROM consent_insurance_acceptance_and_branches_abroad_policies
    JOIN authorised_consents ON consent_insurance_acceptance_and_branches_abroad_policies.consent_id = authorised_consents.id AND consent_insurance_acceptance_and_branches_abroad_policies.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'DAMAGES_AND_PEOPLE_FINANCIAL_RISKS' AS resource_type,
        consent_insurance_financial_risk_policies.consent_id,
        consent_insurance_financial_risk_policies.policy_id AS resource_id,
        consent_insurance_financial_risk_policies.owner_id,
        consent_insurance_financial_risk_policies.status,
        consent_insurance_financial_risk_policies.org_id,
        consent_insurance_financial_risk_policies.created_at,
        consent_insurance_financial_risk_policies.updated_at
    FROM consent_insurance_financial_risk_policies
    JOIN authorised_consents ON consent_insurance_financial_risk_policies.consent_id = authorised_consents.id AND consent_insurance_financial_risk_policies.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'DAMAGES_AND_PEOPLE_HOUSING' AS resource_type,
        consent_insurance_housing_policies.consent_id,
        consent_insurance_housing_policies.policy_id AS resource_id,
        consent_insurance_housing_policies.owner_id,
        consent_insurance_housing_policies.status,
        consent_insurance_housing_policies.org_id,
        consent_insurance_housing_policies.created_at,
        consent_insurance_housing_policies.updated_at
    FROM consent_insurance_housing_policies
    JOIN authorised_consents ON consent_insurance_housing_policies.consent_id = authorised_consents.id AND consent_insurance_housing_policies.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'LIFE_PENSION' AS resource_type,
        consent_insurance_life_pension_contracts.consent_id,
        consent_insurance_life_pension_contracts.contract_id AS resource_id,
        consent_insurance_life_pension_contracts.owner_id,
        consent_insurance_life_pension_contracts.status,
        consent_insurance_life_pension_contracts.org_id,
        consent_insurance_life_pension_contracts.created_at,
        consent_insurance_life_pension_contracts.updated_at
    FROM consent_insurance_life_pension_contracts
    JOIN authorised_consents ON consent_insurance_life_pension_contracts.consent_id = authorised_consents.id AND consent_insurance_life_pension_contracts.org_id = authorised_consents.org_id

    UNION ALL

    SELECT
        'DAMAGES_AND_PEOPLE_PATRIMONIAL' AS resource_type,
        consent_insurance_patrimonial_policies.consent_id,
        consent_insurance_patrimonial_policies.policy_id AS resource_id,
        consent_insurance_patrimonial_policies.owner_id,
        consent_insurance_patrimonial_policies.status,
        consent_insurance_patrimonial_policies.org_id,
        consent_insurance_patrimonial_policies.created_at,
        consent_insurance_patrimonial_policies.updated_at
    FROM consent_insurance_patrimonial_policies
    JOIN authorised_consents ON consent_insurance_patrimonial_policies.consent_id = authorised_consents.id AND consent_insurance_patrimonial_policies.org_id = authorised_consents.org_id;

CREATE TABLE insurance_auto_quotes (
    id TEXT PRIMARY KEY,
    consent_id TEXT NOT NULL,
    status TEXT NOT NULL,
    status_updated_at TIMESTAMPTZ DEFAULT now(),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE insurance_auto_quote_leads (
    id TEXT PRIMARY KEY,
    consent_id TEXT NOT NULL,
    status TEXT NOT NULL,
    status_updated_at TIMESTAMPTZ DEFAULT now(),
    data JSONB NOT NULL,
    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);

CREATE TABLE idempotency_records (
    id TEXT PRIMARY KEY,
    status_code INTEGER NOT NULL,
    request TEXT NOT NULL,
    response TEXT NOT NULL,

    org_id TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT now() NOT NULL
);
