terraform {
  required_version = ">= 1.6.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 4.57.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6.0"
    }
  }
}

provider "azurerm" {
  features {}
  subscription_id = var.subscription_id
}

locals {
  vnet_cidr       = "10.20.0.0/16"
  aca_subnet_cidr = "10.20.0.0/23" # ACA consumption-only env needs /23 or larger subnet (Azure requirement).
  pg_subnet_cidr  = "10.20.2.0/24"
  agw_subnet_cidr = "10.20.3.0/24"
  pg_admin_user   = "pgadmin"
  pg_db_name      = "mockinsurer"
  db_credentials  = jsonencode({
    username = local.pg_admin_user
    password = random_password.pg.result
    host     = azurerm_postgresql_flexible_server.pg.fqdn
    port     = 5432
    dbname   = local.pg_db_name
    sslmode  = "require"
  })
  dns_a_records = toset(["matls-auth.mockinsurer", "matls-api.mockinsurer", "auth.mockinsurer"])
}

resource "azurerm_resource_group" "rg" {
  name     = "${var.project}-${var.env}-rg"
  location = var.location
}

resource "azurerm_dns_zone" "root" {
  name                = var.domain
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_dns_a_record" "agw" {
  for_each            = local.dns_a_records
  name                = each.value
  zone_name           = azurerm_dns_zone.root.name
  resource_group_name = azurerm_resource_group.rg.name
  ttl                 = 300
  records             = [azurerm_public_ip.agw.ip_address]
}

resource "azurerm_virtual_network" "vnet" {
  name                = "${var.project}-${var.env}-vnet"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  address_space       = [local.vnet_cidr]
}

resource "azurerm_subnet" "aca" {
  name                 = "${var.project}-${var.env}-aca-snet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [local.aca_subnet_cidr]

  delegation {
    name = "aca-delegation"

    service_delegation {
      name    = "Microsoft.App/environments"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "pg" {
  name                 = "${var.project}-${var.env}-pg-snet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [local.pg_subnet_cidr]

  delegation {
    name = "postgres-delegation"

    service_delegation {
      name    = "Microsoft.DBforPostgreSQL/flexibleServers"
      actions = ["Microsoft.Network/virtualNetworks/subnets/join/action"]
    }
  }
}

resource "azurerm_subnet" "agw" {
  name                 = "${var.project}-${var.env}-agw-snet"
  resource_group_name  = azurerm_resource_group.rg.name
  virtual_network_name = azurerm_virtual_network.vnet.name
  address_prefixes     = [local.agw_subnet_cidr]
}

resource "azurerm_private_dns_zone" "pg" {
  name                = "private.postgres.database.azure.com"
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_private_dns_zone_virtual_network_link" "pg_link" {
  name                  = "${var.project}-${var.env}-pgdns-link"
  resource_group_name   = azurerm_resource_group.rg.name
  private_dns_zone_name = azurerm_private_dns_zone.pg.name
  virtual_network_id    = azurerm_virtual_network.vnet.id
}

resource "random_password" "pg" {
  length  = 32
  special = true
}

resource "azurerm_postgresql_flexible_server" "pg" {
  name                = "${var.project}-${var.env}-pg"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  administrator_login    = local.pg_admin_user
  administrator_password = random_password.pg.result

  version = "16"

  sku_name   = "B_Standard_B1ms"
  storage_mb = 32768

  delegated_subnet_id = azurerm_subnet.pg.id
  private_dns_zone_id = azurerm_private_dns_zone.pg.id

  public_network_access_enabled = false

  lifecycle {
    ignore_changes = [zone]
  }
  depends_on = [azurerm_private_dns_zone_virtual_network_link.pg_link]
}

resource "azurerm_postgresql_flexible_server_database" "db" {
  name      = local.pg_db_name
  server_id = azurerm_postgresql_flexible_server.pg.id
  charset   = "UTF8"
  collation = "en_US.utf8"
}

resource "azurerm_container_registry" "acr" {
  name                = "mockinsurer"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  sku                 = "Basic"
  admin_enabled       = false
}

resource "azurerm_log_analytics_workspace" "law" {
  name                = "${var.project}-${var.env}-law"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name

  sku               = "PerGB2018"
  retention_in_days = 30
}

resource "azurerm_container_app_environment" "env" {
  name                       = "${var.project}-${var.env}-aca-env"
  location                   = azurerm_resource_group.rg.location
  resource_group_name        = azurerm_resource_group.rg.name
  log_analytics_workspace_id = azurerm_log_analytics_workspace.law.id
  infrastructure_subnet_id   = azurerm_subnet.aca.id
  internal_load_balancer_enabled = false

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      infrastructure_resource_group_name,
      workload_profile,
      tags,
    ]
  }
}

resource "azurerm_user_assigned_identity" "aca_pull" {
  name                = "${var.project}-${var.env}-aca-pull"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_role_assignment" "acr_pull" {
  scope                = azurerm_container_registry.acr.id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_user_assigned_identity.aca_pull.principal_id
}

resource "azurerm_container_app" "mock" {
  name                         = "${var.project}-${var.env}"
  resource_group_name          = azurerm_resource_group.rg.name
  container_app_environment_id = azurerm_container_app_environment.env.id
  revision_mode                = "Single"

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aca_pull.id]
  }

  registry {
    server   = azurerm_container_registry.acr.login_server
    identity = azurerm_user_assigned_identity.aca_pull.id
  }

  secret {
    name  = "db-credentials"
    value = local.db_credentials
  }

  secret {
    name  = "server-transport-cert"
    value = file("${path.module}/keys/server_transport.crt")
  }

  secret {
    name  = "server-transport-key"
    value = file("${path.module}/keys/server_transport.key")
  }

  template {
    volume {
      name         = "keys"
      storage_type = "Secret"
    }

    container {
      name   = "mock"
      image  = "${azurerm_container_registry.acr.login_server}/mockinsurer:${var.image_tag}"
      cpu    = 0.5
      memory = "1Gi"
      
      volume_mounts {
        name = "keys"
        path = "/mnt/secrets/keys"
      }

      env {
        name        = "ENV"
        value       = upper(var.env)
      }

      env {
        name        = "DB_CREDENTIALS"
        secret_name = "db-credentials"
      }

      env {
        name = "PORT"
        value = "80"
      }

      env {
        name  = "BASE_DOMAIN"
        value = "mockinsurer.${var.domain}"
      }

      env {
        name  = "ORG_ID"
        value = var.org_id
      }

      env {
        name  = "KEYSTORE_HOST"
        value = "https://keystore.sandbox.directory.opinbrasil.com.br"
      }

      env {
        name  = "TRANSPORT_CERT_PATH"
        value = "/mnt/secrets/keys/server-transport-cert"
      }

      env {
        name  = "TRANSPORT_KEY_PATH"
        value = "/mnt/secrets/keys/server-transport-key"
      }
    }

    min_replicas = 1
    max_replicas = 3
  }

  ingress {
    external_enabled = true
    target_port      = 80
    transport        = "auto"
    
    traffic_weight {
      percentage      = 100
      latest_revision = true
    }
  }

  lifecycle {
    ignore_changes = [
      workload_profile_name,
    ]
  }
}

data "azurerm_client_config" "current" {}

resource "azurerm_user_assigned_identity" "agw" {
  name                = "${var.project}-${var.env}-agw-identity"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
}

resource "azurerm_public_ip" "agw" {
  name                = "${var.project}-${var.env}-agw-pip"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location
  allocation_method   = "Static"
  sku                 = "Standard"
}

resource "azurerm_storage_account" "static" {
  name                     = "${replace(var.project, "-", "")}${var.env}static"
  resource_group_name      = azurerm_resource_group.rg.name
  location                 = azurerm_resource_group.rg.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  account_kind             = "StorageV2"
  min_tls_version          = "TLS1_2"
}

resource "azurerm_storage_account_static_website" "static" {
  storage_account_id = azurerm_storage_account.static.id
  index_document     = "index.html"
}

resource "azurerm_application_gateway" "agw" {
  name                = "${var.project}-${var.env}-agw"
  resource_group_name = azurerm_resource_group.rg.name
  location            = azurerm_resource_group.rg.location

  sku {
    name     = "Standard_v2"
    tier     = "Standard_v2"
    capacity = 2
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.agw.id]
  }

  gateway_ip_configuration {
    name      = "agw-ip-config"
    subnet_id = azurerm_subnet.agw.id
  }

  frontend_ip_configuration {
    name                 = "agw-frontend-ip"
    public_ip_address_id = azurerm_public_ip.agw.id
  }

  frontend_port {
    name = "https-port"
    port = 443
  }

  ssl_certificate {
    name     = "mtls-server-cert"
    data     = filebase64("${path.module}/keys/server_transport.pfx")
    password = ""
  }

  ssl_certificate {
    name     = "auth-server-cert"
    data     = filebase64("${path.module}/keys/auth_server.pfx")
    password = ""
  }

  trusted_client_certificate {
    name = "client-ca-cert"
    data = filebase64("${path.module}/keys/client_ca.crt")
  }

  ssl_profile {
    name                             = "mtls-ssl-profile"
    trusted_client_certificate_names = ["client-ca-cert"]
    verify_client_cert_issuer_dn     = true
  }

  backend_address_pool {
    name = "container-app-backend"
    fqdns = [azurerm_container_app.mock.ingress[0].fqdn]
  }

  backend_address_pool {
    name = "static-storage-backend"
    fqdns = [azurerm_storage_account.static.primary_web_host]
  }
  
  rewrite_rule_set {
    name = "remove-protected-headers"

    rewrite_rule {
      name          = "remove-client-cert-header"
      rule_sequence = 100

      condition {
        variable    = "var_request_header_X-Client-Cert"
        pattern     = ".*"
        ignore_case = true
      }

      request_header_configuration {
        header_name  = "X-Client-Cert"
        header_value = ""
      }
    }
  }

  rewrite_rule_set {
    name = "forward-client-cert-header"

    rewrite_rule {
      name          = "remove-client-cert-header"
      rule_sequence = 100

      condition {
        variable    = "var_request_header_X-Client-Cert"
        pattern     = ".*"
        ignore_case = true
      }

      request_header_configuration {
        header_name  = "X-Client-Cert"
        header_value = ""
      }
    }

    rewrite_rule {
      name          = "add-client-cert-header"
      rule_sequence = 200

      condition {
        variable    = "var_client_certificate_verification"
        pattern     = "SUCCESS"
        ignore_case = true
      }

      request_header_configuration {
        header_name  = "X-Client-Cert"
        header_value = "{var_client_certificate}"
      }
    }
  }
  
  backend_http_settings {
    name                                = "backend-http-settings"
    cookie_based_affinity               = "Disabled"
    port                                = 443
    protocol                            = "Https"
    request_timeout                     = 20
    pick_host_name_from_backend_address = true
    probe_name                          = "backend-health-probe"
  }

  backend_http_settings {
    name                  = "static-storage-http-settings"
    cookie_based_affinity = "Disabled"
    port                  = 443
    protocol              = "Https"
    request_timeout       = 20
    host_name             = azurerm_storage_account.static.primary_web_host
  }

  probe {
    name                = "backend-health-probe"
    protocol            = "Https"
    path                = "/.well-known/openid-configuration"
    interval            = 10
    timeout             = 10
    unhealthy_threshold = 3
    host                = azurerm_container_app.mock.ingress[0].fqdn
    match {
      status_code = ["200"]
    }
  }

  http_listener {
    name                           = "matls-auth-listener"
    frontend_ip_configuration_name = "agw-frontend-ip"
    frontend_port_name             = "https-port"
    protocol                       = "Https"
    ssl_certificate_name           = "mtls-server-cert"
    ssl_profile_name               = "mtls-ssl-profile"
    host_name                      = "matls-auth.mockinsurer.${var.domain}"
    require_sni                    = true
  }

  http_listener {
    name                           = "matls-api-listener"
    frontend_ip_configuration_name = "agw-frontend-ip"
    frontend_port_name             = "https-port"
    protocol                       = "Https"
    ssl_certificate_name           = "mtls-server-cert"
    ssl_profile_name               = "mtls-ssl-profile"
    host_name                      = "matls-api.mockinsurer.${var.domain}"
    require_sni                    = true
  }

  http_listener {
    name                           = "auth-listener"
    frontend_ip_configuration_name = "agw-frontend-ip"
    frontend_port_name             = "https-port"
    protocol                       = "Https"
    ssl_certificate_name           = "auth-server-cert"
    host_name                      = "auth.mockinsurer.${var.domain}"
    require_sni                    = true
  }

  request_routing_rule {
    name                       = "matls-auth-rule"
    rule_type                  = "Basic"
    http_listener_name         = "matls-auth-listener"
    backend_address_pool_name  = "container-app-backend"
    backend_http_settings_name = "backend-http-settings"
    rewrite_rule_set_name      = "forward-client-cert-header"
    priority                   = 100
  }

  request_routing_rule {
    name                       = "matls-api-rule"
    rule_type                  = "Basic"
    http_listener_name         = "matls-api-listener"
    backend_address_pool_name  = "container-app-backend"
    backend_http_settings_name = "backend-http-settings"
    rewrite_rule_set_name      = "forward-client-cert-header"
    priority                   = 200
  }

  url_path_map {
    name                               = "auth-path-map"
    default_backend_address_pool_name  = "container-app-backend"
    default_backend_http_settings_name = "backend-http-settings"
    default_rewrite_rule_set_name      = "remove-protected-headers"

    path_rule {
      name                       = "static-path-rule"
      paths                      = ["/static/*"]
      backend_address_pool_name  = "static-storage-backend"
      backend_http_settings_name = "static-storage-http-settings"
    }
  }

  request_routing_rule {
    name                       = "auth-rule"
    rule_type                  = "PathBasedRouting"
    http_listener_name         = "auth-listener"
    url_path_map_name          = "auth-path-map"
    priority                   = 300
  }

}

resource "azurerm_container_app_job" "db_migrate" {
  name                         = "${var.project}-${var.env}-db-migrate"
  resource_group_name          = azurerm_resource_group.rg.name
  container_app_environment_id = azurerm_container_app_environment.env.id
  location                     = azurerm_resource_group.rg.location

  replica_timeout_in_seconds = 300
  replica_retry_limit        = 1

  template {
    container {
      name   = "migrate"
      image  = "${azurerm_container_registry.acr.login_server}/mockinsurer-migrate:${var.image_tag}"
      cpu    = 0.5
      memory = "1Gi"

      env {
        name        = "ENV"
        value       = upper(var.env)
      }

      env {
        name        = "DB_CREDENTIALS"
        secret_name = "db-credentials"
      }

      env {
        name        = "ORG_ID"
        value       = var.org_id
      }
    }
  }

  secret {
    name  = "db-credentials"
    value = local.db_credentials
  }

  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aca_pull.id]
  }

  registry {
    server   = azurerm_container_registry.acr.login_server
    identity = azurerm_user_assigned_identity.aca_pull.id
  }

  manual_trigger_config {
    parallelism             = 1
    replica_completion_count = 1
  }

  lifecycle {
    ignore_changes = [
      workload_profile_name,
    ]
  }
}
