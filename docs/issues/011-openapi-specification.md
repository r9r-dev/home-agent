# Add OpenAPI specification

**Priority:** P3 (Low)
**Type:** Enhancement
**Component:** Backend
**Estimated Effort:** High

## Summary

Document the REST API with an OpenAPI 3.0 specification to improve API discoverability and enable client generation.

## Endpoints to Document

### Sessions
- `GET /api/sessions` - List all sessions
- `GET /api/sessions/:id` - Get session details
- `GET /api/sessions/:id/messages` - Get session messages
- `PATCH /api/sessions/:id/model` - Update session model
- `DELETE /api/sessions/:id` - Delete session
- `GET /api/sessions/:id/tool-calls` - List tool calls

### Memory
- `GET /api/memory` - List all entries
- `POST /api/memory` - Create entry
- `GET /api/memory/:id` - Get entry
- `PUT /api/memory/:id` - Update entry
- `DELETE /api/memory/:id` - Delete entry
- `GET /api/memory/export` - Export as JSON
- `POST /api/memory/import` - Import from JSON

### Files
- `POST /api/upload` - Upload file
- `GET /api/uploads/:filename` - Serve file
- `DELETE /api/uploads/:id` - Delete file

### Machines
- `GET /api/machines` - List machines
- `POST /api/machines` - Create machine
- `GET /api/machines/:id` - Get machine
- `PUT /api/machines/:id` - Update machine
- `DELETE /api/machines/:id` - Delete machine
- `POST /api/machines/:id/test` - Test connection

### Search
- `GET /api/search` - Search messages

### Settings
- `GET /api/settings` - Get settings
- `PUT /api/settings/:key` - Update setting

### System
- `GET /health` - Health check
- `GET /api/info` - API info

## Example OpenAPI Spec

```yaml
# api/openapi.yaml
openapi: 3.0.0
info:
  title: Home Agent API
  version: 1.0.0
  description: REST API for Home Agent chat interface

servers:
  - url: http://localhost:8080
    description: Development server

paths:
  /api/sessions:
    get:
      summary: List all sessions
      tags: [Sessions]
      responses:
        '200':
          description: List of sessions
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Session'

  /api/sessions/{id}:
    get:
      summary: Get session by ID
      tags: [Sessions]
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      responses:
        '200':
          description: Session details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Session'
        '404':
          $ref: '#/components/responses/NotFound'

components:
  schemas:
    Session:
      type: object
      properties:
        id:
          type: integer
        session_id:
          type: string
          format: uuid
        title:
          type: string
        model:
          type: string
          enum: [haiku, sonnet, opus]
        created_at:
          type: string
          format: date-time
        last_activity:
          type: string
          format: date-time

    Message:
      type: object
      properties:
        id:
          type: integer
        session_id:
          type: string
        role:
          type: string
          enum: [user, assistant, thinking]
        content:
          type: string
        created_at:
          type: string
          format: date-time

    MemoryEntry:
      type: object
      properties:
        id:
          type: string
        title:
          type: string
        content:
          type: string
        enabled:
          type: boolean

  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema:
            type: object
            properties:
              error:
                type: string
              message:
                type: string
```

## Tasks

- [ ] Create `api/openapi.yaml`
- [ ] Document all endpoints with request/response
- [ ] Define all schemas
- [ ] Add authentication documentation
- [ ] Set up Swagger UI endpoint at `/api/docs`
- [ ] Add request validation middleware
- [ ] Generate TypeScript types from spec
- [ ] Add WebSocket protocol documentation

## Acceptance Criteria

- [ ] All REST endpoints documented
- [ ] Request/response examples included
- [ ] Swagger UI accessible at `/api/docs`
- [ ] TypeScript types can be generated from spec
- [ ] WebSocket protocol documented separately

## References

- `ARCHITECTURE_REVIEW.md` section "1. Create Shared API Contract"
- OpenAPI Specification: https://swagger.io/specification/

## Labels

```
priority: P3
type: enhancement
component: backend
```
