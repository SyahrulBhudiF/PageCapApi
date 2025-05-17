# Doc-Management

This project is a document management system built with Go. It follows a layered architecture pattern, separating concerns into distinct layers to improve maintainability, testability, and scalability. The structure incorporates principles of Domain-Driven Design (DDD), focusing on organizing the codebase around the core business domain, and utilizes Hexagonal Architecture as an implementation strategy.

## Architecture Overview

The project structure reflects a hexagonal architecture approach, organized into the following main directories within the `internal/` folder:

- **`domain/`**: Contains the core business logic, entities, and interfaces (contracts) that define the application's domain. This layer is independent of any infrastructure details and is the heart of the application, reflecting the business concepts. This aligns with the core focus of DDD.
- **`application/`**: Houses the application's use cases or interactors. These orchestrate the domain entities and interact with infrastructure through interfaces defined in the domain layer. This layer defines the application's capabilities.
- **`infrastructure/`**: Provides the concrete implementations for the interfaces defined in the domain layer. This includes database access (PostgreSQL), external services (Redis, JWT, Mail), and persistence logic. These are the "adapters" that connect the core logic to the outside world, implementing the "ports" defined in the domain layer.
- **`interfaces/`**: Contains the entry points into the application, such as HTTP handlers and routes. This layer translates external requests into calls to the application layer and formats responses. These also act as adapters to the core.
- **`shared/`**: Includes common utilities, base entities, and generic repository interfaces used across different layers.

This structure, combining DDD principles with Hexagonal Architecture, ensures that the core business logic (`domain` and `application`) remains independent of external frameworks and databases. This makes it easier to change or replace infrastructure components without affecting the core functionality, improving the system's resilience and maintainability.

## Purpose

The purpose of this project is to provide a system for managing documents. Based on the presence of authentication-related components, it includes features for user authentication and likely provides secure access to document-related functionalities.
