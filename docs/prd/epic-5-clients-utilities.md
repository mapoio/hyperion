# Epic 5: Clients & Utilities

**Priority**: ⭐⭐⭐⭐ (High)
**Estimated Duration**: 1 week
**Status**: Not Started
**Dependencies**: Epic 1 (Core Foundation)

---

## Overview

Implement HTTP client with automatic tracing, object storage integration, and encryption utilities.

---

## Goals

- Provide HTTP client with automatic trace context propagation
- Support S3-compatible object storage
- Provide secure encryption/decryption utilities
- Ensure all client operations are automatically traced

---

## User Stories

### Story 5.1: HTTP Client (hyperhttp)

**As a** framework user
**I want** an HTTP client with automatic tracing
**So that** I can make external API calls with distributed tracing support

**Acceptance Criteria**:
- [ ] Can make HTTP requests (GET, POST, PUT, DELETE)
- [ ] Trace context automatically injected into headers
- [ ] Automatically creates `http.client.*` spans
- [ ] Support for timeouts and retries
- [ ] Request/response logging

**Tasks**:
- [ ] Define `Client` struct
- [ ] Implement based on `go-resty/resty`
- [ ] Implement automatic trace context injection
- [ ] Implement automatic span creation for each request
- [ ] Implement GET/POST/PUT/DELETE/PATCH methods
- [ ] Implement request/response middleware
- [ ] Implement retry mechanism with exponential backoff
- [ ] Implement timeout configuration
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with mock HTTP server
- [ ] Write godoc documentation

**Technical Details**:
```go
// HTTP client with tracing
func (c *Client) Get(ctx hyperctx.Context, url string) (*resty.Response, error) {
    // Create span
    ctx, end := ctx.StartSpan("http.client", "GET", url)
    defer end()

    req := c.client.R().SetContext(ctx)

    // Auto-inject trace context
    otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))

    resp, err := req.Get(url)
    if err != nil {
        ctx.RecordError(err)
        return nil, hypererror.InternalWrap("http request failed", err)
    }

    // Add span attributes
    ctx.SetAttributes(
        attribute.Int("http.status_code", resp.StatusCode()),
        attribute.Int64("http.response_size", resp.Size()),
    )

    return resp, nil
}
```

**Estimated**: 2 days

---

### Story 5.2: Object Storage (hyperstore)

**As a** framework user
**I want** S3-compatible object storage support
**So that** I can store and retrieve files in cloud storage

**Acceptance Criteria**:
- [ ] Can upload/download objects
- [ ] Can list objects with prefix
- [ ] Can delete objects
- [ ] All operations automatically traced
- [ ] Support for presigned URLs

**Tasks**:
- [ ] Define `ObjectStorage` interface
- [ ] Implement S3 client based on AWS SDK
- [ ] Implement Put/Get/Delete/List operations
- [ ] Implement presigned URL generation
- [ ] Implement automatic tracing for all operations
- [ ] Implement multipart upload for large files
- [ ] Write unit tests (>80% coverage)
- [ ] Write integration tests with MinIO
- [ ] Write godoc documentation

**Technical Details**:
```go
type ObjectStorage interface {
    Put(ctx hyperctx.Context, bucket, key string, data io.Reader) error
    Get(ctx hyperctx.Context, bucket, key string) (io.ReadCloser, error)
    Delete(ctx hyperctx.Context, bucket, key string) error
    List(ctx hyperctx.Context, bucket, prefix string) ([]string, error)
    PresignedURL(ctx hyperctx.Context, bucket, key string, ttl time.Duration) (string, error)
}
```

**Estimated**: 2 days

---

### Story 5.3: Encryption (hypercrypto)

**As a** framework user
**I want** secure encryption utilities
**So that** I can encrypt sensitive data at rest

**Acceptance Criteria**:
- [ ] Can encrypt/decrypt data using AES-GCM
- [ ] Support for key rotation
- [ ] Secure key management
- [ ] Type-safe interface

**Tasks**:
- [ ] Define `Crypter` interface
- [ ] Implement AES-256-GCM encryption
- [ ] Implement key derivation from passphrase
- [ ] Implement encryption/decryption methods
- [ ] Implement nonce generation
- [ ] Write unit tests (>80% coverage)
- [ ] Write security tests
- [ ] Write godoc documentation

**Technical Details**:
```go
type Crypter interface {
    Encrypt(plaintext []byte) (ciphertext []byte, err error)
    Decrypt(ciphertext []byte) (plaintext []byte, err error)
}

// AES-GCM encryption
type AESCrypter struct {
    key []byte // 32 bytes for AES-256
}

func (c *AESCrypter) Encrypt(plaintext []byte) ([]byte, error) {
    block, err := aes.NewCipher(c.key)
    if err != nil {
        return nil, err
    }

    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }

    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }

    ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
    return ciphertext, nil
}
```

**Estimated**: 1 day

---

## Milestone

**Deliverable**: Production-ready client libraries and utilities

**Demo Scenario**:

**HTTP Client with Tracing**:
```go
// Service calling external API
func (s *UserService) EnrichUserData(ctx hyperctx.Context, userID string) error {
    // Call external API - automatically traced
    resp, err := s.httpClient.Get(ctx, fmt.Sprintf("https://api.example.com/users/%s", userID))
    if err != nil {
        return err
    }

    var enrichedData map[string]any
    if err := json.Unmarshal(resp.Body(), &enrichedData); err != nil {
        return hypererror.InternalWrap("failed to parse response", err)
    }

    // Update user with enriched data
    return s.userRepo.Update(ctx, userID, enrichedData)
}
```

**Object Storage**:
```go
// Upload file to S3
func (s *DocumentService) UploadDocument(ctx hyperctx.Context, userID string, file io.Reader) error {
    key := fmt.Sprintf("users/%s/documents/%s", userID, uuid.New().String())

    if err := s.storage.Put(ctx, "my-bucket", key, file); err != nil {
        return hypererror.InternalWrap("failed to upload document", err)
    }

    ctx.Logger().Info("document uploaded", "key", key)
    return nil
}
```

**Encryption**:
```go
// Encrypt sensitive data before storing
func (s *UserService) CreateUser(ctx hyperctx.Context, req *CreateUserRequest) error {
    // Encrypt sensitive fields
    encryptedSSN, err := s.crypter.Encrypt([]byte(req.SSN))
    if err != nil {
        return hypererror.InternalWrap("failed to encrypt SSN", err)
    }

    user := &User{
        Username:     req.Username,
        EncryptedSSN: encryptedSSN,
    }

    return s.userRepo.Create(ctx, user)
}
```

---

## Technical Notes

### Architecture Decisions

- **Resty for HTTP Client**: Simple API, built-in retry, middleware support
- **AWS SDK for S3**: Official SDK, well-maintained, feature-complete
- **AES-GCM for Encryption**: Authenticated encryption, industry standard

### Dependencies

- `github.com/go-resty/resty/v2` - HTTP client
- `github.com/aws/aws-sdk-go-v2` - AWS SDK
- Standard library `crypto/aes`, `crypto/cipher` - Encryption

### Configuration

```yaml
http_client:
  timeout: 30s
  retry:
    max_attempts: 3
    wait_time: 1s
    max_wait_time: 30s

object_storage:
  provider: s3  # s3, minio, gcs
  endpoint: "s3.amazonaws.com"
  region: "us-east-1"
  access_key: "${AWS_ACCESS_KEY}"
  secret_key: "${AWS_SECRET_KEY}"

crypto:
  algorithm: aes-256-gcm
  key: "${ENCRYPTION_KEY}"  # 32 bytes for AES-256
```

### Security Considerations

**Encryption**:
- Use AES-256-GCM (Galois/Counter Mode) for authenticated encryption
- Generate unique nonce for each encryption
- Never reuse nonces
- Store encryption keys securely (environment variables, vault)
- Implement key rotation mechanism

**HTTP Client**:
- Validate TLS certificates
- Set reasonable timeouts
- Sanitize URLs in logs
- Don't log sensitive headers

**Object Storage**:
- Use presigned URLs for temporary access
- Implement bucket policies
- Encrypt objects at rest
- Enable versioning for critical data

### Testing Strategy

- **Unit Tests**:
  - HTTP client retry logic
  - Encryption/decryption correctness
  - S3 operations with mocked client
- **Integration Tests**:
  - HTTP client with mock server (httptest)
  - S3 operations with MinIO
- **Security Tests**:
  - Encryption strength validation
  - Nonce uniqueness

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| HTTP client timeout issues | Medium | Configurable timeouts, retry |
| Encryption key leakage | High | Secure key storage, rotation |
| S3 costs | Low | Lifecycle policies, monitoring |

---

## Related Documentation

- [Architecture - hyperhttp](../architecture.md#58-hyperhttp---http-client)
- [Tech Stack - HTTP Client](../architecture/tech-stack.md#http-client-resty)

---

**Last Updated**: 2025-01-XX
