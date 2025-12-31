# Mock Insurer

Mock Insurer is a mock implementation of the Open Insurance Brasil API specifications. It serves as a reference platform for ecosystem participants to develop, test, and validate their applications in a controlled environment, without depending on real institutions.

## Endpoints & Hosts

| URL | Description | mTLS |
|-----|-------------|------|
| `https://auth.mockinsurer.{host}` | Authorization Server | No |
| `https://matls-auth.mockinsurer.{host}` | Authorization Server (mTLS) | **Required** |
| `https://matls-api.mockinsurer.{host}` | Bank Backend (mTLS) | **Required** |

## Mock Users

Mock Insurer comes with predefined users preloaded with test data to facilitate development and testing across all APIs.

> **Default Password:** All users share the password: `P@ssword01`

| Username | CPF | CNPJ | Description |
|----------|-----|------|-------------|
| `usuario1@seguradoramodelo.com.br` | `761.092.776-73` | `50.685.362/0006-73` | Primary test user with resources in all APIs |

## Getting Started

### Prerequisites
- Go 1.24+ (For development only)
- Docker and Docker Compose
- Git

Add the entries below to `/etc/hosts` (or `C:\Windows\System32\drivers\etc\hosts` on Windows):

```bash
127.0.0.1 auth.mockinsurer.local
127.0.0.1 matls-auth.mockinsurer.local
127.0.0.1 matls-api.mockinsurer.local
127.0.0.1 directory.local
127.0.0.1 keystore.local
127.0.0.1 database.local
```

### Quick Start

1. **Clone and setup**:
   ```bash
   git clone https://github.com/luikyv/mock-insurer
   cd mock-insurer
   make setup
   ```

2. **Run the application**:
   ```bash
   make run
   ```

The application will be available at:
- Bank Server: https://matls-api.mockinsurer.local
- Authorization Server: https://auth.mockinsurer.local

### Development Setup

For development with additional tools:
```bash
make setup-dev
```

### Running with Conformance Suite

To run Mock Insurer with the Open Finance Conformance Suite:

1. **Setup the Conformance Suite**:
   ```bash
   make setup-cs
   ```

2. **Run with Conformance Suite**:
   ```bash
   make run-with-cs
   ```

## TODO
- Drop all data.
- Migration as a container?
- DB indexes.
- Add doc.go's.
- Remove descriptions.
- Improve error handling.
- Add logs.
