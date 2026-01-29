# 🛒 E-Commerce Backend API (Go)

A full-featured **E-Commerce Backend API** built using **Golang**, designed for scalability, security, and real-world use cases. This project includes authentication, product management, cart, wishlist, and checkout functionality with support for **Google OAuth**, **JWT authentication**, and **product variants**.

---

## 🚀 Features

### 🔐 Authentication & Authorization
- User Authentication (Register / Login)
- Admin Authentication
- Google OAuth 2.0 Login
- JWT Access & Refresh Tokens
- Role-based Access Control (User / Admin)

### 🗂️ Product Management
- Categories
- Brands
- Products
- Product Variants (size, color, etc.)
- Product Attributes (dynamic attributes)

### 🛍️ Shopping Features
- Shopping Cart
- Wishlist
- Checkout from:
  - Cart
  - Direct Product Variant
- Order Management
---

## 🧱 Tech Stack

- **Language:** Golang (Go)
- **Framework:** Gin 
- **Database:** PostgreSQL / Redis 
- **Driver:** pgx 
- **Authentication:** JWT + Google OAuth
- **API Type:** RESTful API

---

## 📦 Order & Post-Checkout Features

### 🧾 Order Management
- Order Creation from:
  - Cart
  - Direct Product Variant Checkout
- Order Status Lifecycle Tracking
- Order Item–level breakdown
- User & Admin Order Views
- Support for:
  - Full Order Cancellation
  - Partial Order Cancellation (per item)

---

### 🔄 Returns & Refunds
- Order Return Requests
- Full & Partial Returns
- Refund Initiation & Processing
- Refund Status Tracking
- Refunds mapped to original payment method
- Separate flows for:
  - COD Refunds
  - Online Payment Refunds

---

### 💳 Payments & Checkout Enhancements
- Cash on Delivery (COD)
- Razorpay Payment Gateway Integration
  - Payment Link Creation
  - Payment Capture & Verification
  - Webhook Handling
  - Secure Signature Validation
- Payment Status Tracking
- Safe Order Confirmation after Payment Success
- Handling Payment Failures & Retries

---

### 🧠 Business Logic Enhancements
- Transaction-safe Order & Payment Processing
- Inventory Validation during Checkout
- Item-level price calculation
- Idempotent Payment & Webhook APIs
- Support for real-world checkout edge cases

---
