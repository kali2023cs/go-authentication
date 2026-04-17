# 🚀 Gin Auth Pro: The Ultimate Advanced Authentication Guide

Welcome to the **Gin Auth Pro** project! This repository contains a state-of-the-art, production-ready authentication system built with Go. This is not just a basic login system; it integrates advanced security "habits" used by world-class engineering teams.

---

## 🏗️ Integrated Technologies & Concepts

### 🔐 Advanced Security Habits
-   **Dual-Token Architecture**: Uses short-lived **Access Tokens** (15 min) and long-lived **Refresh Tokens** (7 days).
-   **Token Rotation & Reuse Detection**: Every time a token is refreshed, both tokens are rotated. If a stolen refresh token is reused, the system automatically revokes all sessions for that user for safety.
-   **Secure Cookie Storage**: Implements **HTTP-only, SameSite** cookies for token storage, providing robust protection against XSS and CSRF attacks.
-   **Session Tracking**: A dedicated database table tracks active sessions, allowing for immediate global logout and revocation.
-   **Multi-Strategy Auth**: Supports both standard password-based login and **Google OAuth 2.0**.

### 🛠️ Core Stack
-   **Gin Gonic**: High-performance HTTP web framework.
-   **GORM**: Modern ORM for Go mapping structs to MySQL tables.
-   **Bcrypt**: Industry-standard salted password hashing.
-   **SMTP Integration**: Real-world communication via Gmail servers for OTP delivery.
-   **Middleware Pattern**: Clean interception of requests for authorization and session validation.

---

## 🔄 The Full Flow: Advanced Logic

### 1. Secure Login & Session Initialization
-   **Logic**:
    1.  **Validation**: Verifies password or Google account.
    2.  **Dual Issuance**: Generates an Access Token and a Refresh Token.
    3.  **Persistence**: The Refresh Token is hashed and stored in the `refresh_tokens` table.
    4.  **Cookie Injection**: Both tokens are set as **HTTP-only** cookies.

### 2. Silent Token Refresh (The "Graceful" Rotation)
-   **Endpoint**: `POST /refresh`
-   **Logic**:
    1.  **Cookie Extraction**: Reads the refresh token from the secure cookie.
    2.  **Identity Verification**: Validates the token and checks if it exists in the database.
    3.  **Rotation Logic**: If valid, the old token is deleted, and a brand new pair is issued and stored.
    4.  **Breach Detection**: If a valid JWT is provided but it's **missing** from the DB, the system assumes it was previously used (stolen) and clears **all** sessions for that User ID.

### 3. Google OAuth 2.0 Flow
-   **Logic**:
    1.  **Redirect**: User is sent to Google's consent screen.
    2.  **Callback**: Server receives an authorization code and exchanges it for a Google Identity token.
    3.  **Sync**: The system finds the user by email or Google ID. If they don't exist, it creates a verified account automatically.
    4.  **Token Issue**: Standard dual-token session is initialized.

---

## 🎓 The Senior Developer Interview Prep

### 🟢 Core Concepts (Access & Refresh)
1.  **Why use both Access and Refresh tokens instead of one long-lived token?**
    *   *Answer:* It minimizes the "Window of Opportunity" for an attacker. If an access token is stolen, it only works for a few minutes. The long-lived refresh token is stored more securely (HTTP-only) and can be revoked by the server if a breach is detected.
2.  **What is the "stateless" trade-off when adding a session table?**
    *   *Answer:* Pure JWT is stateless but impossible to revoke. By adding a small `refresh_tokens` table, we gain the ability to "logout" or "revoke" sessions while keeping the `access_token` validation stateless and fast in the middleware.

### 🟡 Security Masterclass (Cookies & CSRF)
3.  **Why are HTTP-only cookies better than LocalStorage for tokens?**
    *   *Answer:* LocalStorage is accessible by JavaScript. If your site has an XSS vulnerability, an attacker can steal the token. HTTP-only cookies are invisible to JavaScript, making them immune to XSS theft.
4.  **How does Token Rotation prevent "Replay Attacks"?**
    *   *Answer:* Since each refresh token is "one-time use," an attacker cannot use an old token they captured. If they try, the server sees it's missing from the DB and triggers a security alert (revoking all sessions).

### 🔴 Architectural Deep Dive (OAuth & GORM)
5.  **How do you handle "Account Linking" between Email and Google login?**
    *   *Answer:* We use the **Email as the unique identifier**. If a user registers with email/pass first and later logs in with Google, we link them via the email address and update their `google_id` in the database.
6.  **Explain GORM's `gorm.DeletedAt` field in the Session model.**
    *   *Answer:* It enables **Soft Deletes**. When you "delete" a session, GORM simply updates the `deleted_at` timestamp. This allows us to keep an audit trail of sessions for security analysis without them appearing in normal queries.

---

## 🛠️ Production Checklist
- [x] **Secure Secrets**: All JWT secrets moved to `.env`.
- [ ] **HTTPS**: Enable `Secure: true` on cookies in production.
- [ ] **Rate Limiting**: Protect the `/login` and `/refresh` endpoints from brute force.
- [ ] **Redis Migration**: Move Refresh Tokens to Redis for sub-millisecond session validation.

> [!IMPORTANT]
> This project follows the **Clean Architecture** principles. Every utility, middleware, and controller is built to be modular and testable.
