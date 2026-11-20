# Learning Journal

## Lesson 001: Writing a Value to the Cache

### Goal

Implement `PUT /cache/{key}` so a client can store a JSON value under a cache key. The HTTP layer should enforce the request contract, the application layer should express the caching operation, and the infrastructure layer should translate it to Redis.

### Exact Request Example

```http
PUT /cache/user:42 HTTP/1.1
Host: localhost:8080
Content-Type: application/json

{"name":"Ada","role":"admin"}
```

Equivalent `curl` request:

```bash
curl -X PUT http://localhost:8080/cache/user:42 \
  -H 'Content-Type: application/json' \
  -d '{"name":"Ada","role":"admin"}'
```

### Request Trace

1. `cmd/api` registers the cache route and connects the HTTP server to the application dependencies while owning process lifecycle.
2. The controller receives the request and adapts path, headers, and body into the application input.
3. The controller strictly decodes the JSON body, rejecting unknown fields and malformed JSON.
4. Custom validation checks the key and decoded value against application rules.
5. The use case coordinates the write without depending on Redis details.
6. The use case calls the Redis port defined by the application boundary.
7. The Redis adapter translates the port call into a Redis RESP command and sends it to Redis.
8. The adapter maps Redis success or failure back to the port result.
9. The controller adapts the result into the HTTP response.

### Contracts

#### Request

- Method: `PUT`
- Path: `/cache/{key}`
- Header: `Content-Type: application/json`
- Body: one JSON object representing the value to store
- Key: supplied by the `{key}` path parameter
- Unknown JSON fields: rejected
- Malformed JSON or invalid application data: rejected

#### Success

- Status: `201 Created`
- Body: empty
- Meaning: the value was written successfully

#### Error

- `400 Bad Request`: malformed JSON, unsupported fields, or failed validation
- `500 Internal Server Error`: application or Redis failure that is not a client input error
- Error responses should use the service's standard JSON error shape.

### Proposed Debugging Breakpoint Order

1. Route registration in `cmd/api`: verify the request reaches the intended handler.
2. Controller entry: inspect the path key, content type, and raw body.
3. Strict decoder: confirm malformed JSON and unknown fields fail here.
4. Custom validation: inspect the exact validation input and returned errors.
5. Use-case entry: verify the application command is complete.
6. Redis port invocation: verify the key and serialized value passed across the boundary.
7. RESP adapter: inspect the generated Redis command and Redis reply.
8. Controller response mapping: verify status code and error translation.

### Observed Edge Case

**Whitespace-only keys can be written before later response validation rejects the request.** This indicates that key validation or response validation is occurring after the Redis write. Validation that protects the operation should happen before the use case invokes the Redis port; response construction should not be the first place an invalid key is detected.

### Next Questions

1. Where should key normalization occur, and should surrounding whitespace be trimmed or rejected?
2. What exact JSON value types and maximum payload size should the cache contract permit?
3. How should Redis timeouts, unavailable connections, and command errors map to HTTP responses?
