CREATE TYPE order_items_status AS ENUM (
    'pending', 'cancelled', 'returned', 'delivered', 'failed', 'confirmed', 'shipped'
);

CREATE TYPE order_status AS ENUM (
    'pending', 'confirmed', 'cancelled', 'returned', 'failed', 'shipped', 'delivered'
);

CREATE TYPE payment_status AS ENUM (
    'pending', 'paid', 'failed', 'refunded', 'cancelled'
);

CREATE TYPE return_item_status AS ENUM (
    'requested', 'approved', 'rejected', 'received', 'refunded', 'store_credit', 'damaged'
);

CREATE TYPE return_status AS ENUM (
    'requested', 'partially_approved', 'approved', 'partially_received', 'received',
    'partially_refunded', 'refunded', 'rejected', 'cancelled'
);

CREATE TYPE shipment_status AS ENUM (
    'pending', 'shipped', 'in_transit', 'delivered', 'returned', 'cancelled'
);

CREATE TYPE shipment_type AS ENUM (
    'normal', 'express'
);

CREATE TYPE user_role AS ENUM (
    'user', 'admin'
);

CREATE TYPE user_status AS ENUM (
    'active', 'blocked', 'deleted'
);

CREATE TYPE wallet_transaction_type AS ENUM (
    'purchase', 'refund', 'topup'
);


CREATE TABLE attributes (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    data_type VARCHAR(10) NOT NULL
        CHECK (data_type IN ('number', 'text', 'boolean', 'date')),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE attribute_values (
    id BIGSERIAL PRIMARY KEY,
    attribute_id BIGINT NOT NULL
        REFERENCES attributes(id) ON DELETE CASCADE,
    value VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE brands (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    logo_url VARCHAR(500),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    parent_category_id BIGINT
        REFERENCES categories(id) ON DELETE CASCADE,
    image_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(320) NOT NULL UNIQUE,
    phone VARCHAR(15) UNIQUE,
    password TEXT NOT NULL,
    first_name VARCHAR(200) NOT NULL,
    last_name VARCHAR(200) NOT NULL,
    role user_role DEFAULT 'user' NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE NOT NULL,
    status user_status DEFAULT 'active' NOT NULL,
    provider VARCHAR(100) DEFAULT 'local',
    default_address_id BIGINT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ,
    profile_img_url VARCHAR(255),
    referral_code VARCHAR UNIQUE NOT NULL,
    referred_by_user_id BIGINT NULL
);


CREATE TABLE user_addresses (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    label VARCHAR(100),
    address_line VARCHAR(400) NOT NULL,
    address_line_2 VARCHAR(400),
    pincode VARCHAR(20) NOT NULL,
    city VARCHAR(200) NOT NULL,
    state VARCHAR(200),
    country VARCHAR(200) NOT NULL,
    district VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

ALTER TABLE users
    ADD CONSTRAINT fk_default_address
    FOREIGN KEY (default_address_id)
    REFERENCES user_addresses(id)
    ON DELETE SET NULL;


CREATE TABLE carts (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE cart_items (
    id BIGSERIAL PRIMARY KEY,
    cart_id BIGINT NOT NULL
        REFERENCES carts(id) ON DELETE CASCADE,
    product_variant_id BIGINT NOT NULL
        REFERENCES product_variants(id) ON DELETE RESTRICT,
    quantity INTEGER DEFAULT 1 NOT NULL CHECK (quantity > 0),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT ux_cart_items_cart_variant UNIQUE (cart_id, product_variant_id)
);


CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(500) NOT NULL,
    brand_id BIGINT
        REFERENCES brands(id) ON DELETE SET NULL,
    description TEXT,
    category_id BIGINT
        REFERENCES categories(id) ON DELETE SET NULL,
    image_url TEXT NOT NULL,
    min_price NUMERIC(12,2) NOT NULL CHECK (min_price >= 0),
    max_price NUMERIC(12,2) NOT NULL CHECK (max_price >= 0),
    is_digital BOOLEAN DEFAULT FALSE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE product_variants (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,
    sku VARCHAR(128),
    original_price NUMERIC(12,2) NOT NULL CHECK (original_price >= 0),
    sale_price NUMERIC(12,2) CHECK (sale_price >= 0),
    stock INTEGER NOT NULL CHECK (stock >= 0),
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ
);


CREATE TABLE product_variant_images (
    id BIGSERIAL PRIMARY KEY,
    product_variant_id BIGINT NOT NULL
        REFERENCES product_variants(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE product_variant_attributes (
    id BIGSERIAL PRIMARY KEY,
    product_variant_id BIGINT NOT NULL
        REFERENCES product_variants(id) ON DELETE CASCADE,
    attribute_id BIGINT NOT NULL
        REFERENCES attributes(id) ON DELETE CASCADE,
    attribute_value_id BIGINT NOT NULL
        REFERENCES attribute_values(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE coupons (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL
        CHECK (discount_type IN ('fixed', 'percentage')),
    discount_value NUMERIC(10,2) NOT NULL
        CHECK (discount_value >= 0),
    min_order_amount NUMERIC(12,2) NOT NULL DEFAULT 0
        CHECK (min_order_amount >= 0),
    max_discount_amount NUMERIC(12,2)
        CHECK (max_discount_amount >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (valid_to IS NULL OR valid_to > valid_from)
);


CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE RESTRICT,
    total_amount NUMERIC(14,2) NOT NULL CHECK (total_amount >= 0),
    tax_amount NUMERIC(14,2) NOT NULL CHECK (tax_amount >= 0),
    status order_status DEFAULT 'pending' NOT NULL,
    shipping_address_id BIGINT
        REFERENCES user_addresses(id) ON DELETE SET NULL,
    billing_address_id BIGINT
        REFERENCES user_addresses(id) ON DELETE SET NULL,
    public_order_id VARCHAR(100) NOT NULL UNIQUE,
    estimated_delivery_date DATE,
    return_status return_status,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL
        REFERENCES orders(id) ON DELETE CASCADE,
    product_variant_id BIGINT
        REFERENCES product_variants(id) ON DELETE SET NULL,
    sku_at_purchase VARCHAR(128),
    product_name_at_purchase VARCHAR(500),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL CHECK (unit_price >= 0),
    total_price NUMERIC(14,2) NOT NULL CHECK (total_price >= 0),
    status order_items_status DEFAULT 'pending' NOT NULL,
    reason TEXT,
    return_status VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE order_coupons (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE
        REFERENCES orders(id) ON DELETE CASCADE,
    coupon_id BIGINT NOT NULL
        REFERENCES coupons(id) ON DELETE RESTRICT,
    discount_applied NUMERIC(12,2) NOT NULL
        CHECK (discount_applied >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT
        REFERENCES orders(id) ON DELETE SET NULL,
    user_id BIGINT
        REFERENCES users(id) ON DELETE SET NULL,
    amount NUMERIC(14,2) NOT NULL CHECK (amount >= 0),
    collected_amount NUMERIC(14,2),
    refund_amount NUMERIC(14,2),
    currency VARCHAR(3) DEFAULT 'INR' NOT NULL,
    provider VARCHAR(50) NOT NULL,
    provider_order_id VARCHAR(255),
    status payment_status DEFAULT 'pending' NOT NULL,
    failure_reason TEXT,
    idempotency_key VARCHAR(255),
    paid_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT ux_payments_provider_pid UNIQUE (provider, provider_order_id)
);


CREATE TABLE shipments (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL
        REFERENCES orders(id) ON DELETE CASCADE,
    carrier VARCHAR(255),
    tracking_number VARCHAR(255),
    status shipment_status DEFAULT 'pending' NOT NULL,
    type shipment_type NOT NULL,
    shipped_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE order_returns (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL UNIQUE
        REFERENCES orders(id),
    user_id BIGINT NOT NULL
        REFERENCES users(id),
    refunded_amount NUMERIC(12,2) NOT NULL CHECK (refunded_amount >= 0),
    status return_status DEFAULT 'requested' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE return_items (
    id BIGSERIAL PRIMARY KEY,
    return_id BIGINT NOT NULL
        REFERENCES order_returns(id),
    order_item_id BIGINT NOT NULL
        REFERENCES order_items(id),
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    status return_item_status DEFAULT 'requested' NOT NULL,
    reason TEXT NOT NULL,
    requested_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    approved_at TIMESTAMPTZ,
    order_item_price NUMERIC(12,2) NOT NULL
);


CREATE TABLE reviews (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,
    user_id BIGINT
        REFERENCES users(id) ON DELETE SET NULL,
    rating SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title VARCHAR(500),
    body TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE wishlists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE wishlist_items (
    id BIGSERIAL PRIMARY KEY,
    wishlist_id BIGINT NOT NULL
        REFERENCES wishlists(id) ON DELETE CASCADE,
    product_id BIGINT NOT NULL
        REFERENCES products(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_wishlist_id_product_id UNIQUE (wishlist_id, product_id)
);


CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,
    balance NUMERIC(12,2) NOT NULL CHECK (balance >= 0),
    is_admin BOOLEAN DEFAULT FALSE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL
        REFERENCES wallets(id),
    amount NUMERIC(12,2) NOT NULL CHECK (amount > 0),
    transaction_type wallet_transaction_type NOT NULL,
    related_order BIGINT
        REFERENCES orders(id),
    remarks TEXT,
    balance_before NUMERIC(12,2) NOT NULL CHECK (balance_before >= 0),
    balance_after NUMERIC(12,2) NOT NULL CHECK (balance_after >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE otps (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(100) NOT NULL
        REFERENCES users(email) ON DELETE CASCADE,
    otp VARCHAR(50) NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,
    expiry_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE password_reset_tokens (
    id BIGSERIAL PRIMARY KEY,
    token VARCHAR(255) NOT NULL,
    is_used BOOLEAN DEFAULT FALSE NOT NULL,
    expiry_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);


CREATE TABLE refresh_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE
        REFERENCES users(id),
    token VARCHAR(255) NOT NULL,
    revoked BOOLEAN DEFAULT FALSE NOT NULL,
    issued_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    expiry_at TIMESTAMPTZ NOT NULL
);

-- Main offer details
CREATE TABLE offers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL CHECK (discount_type IN ('fixed', 'percentage')),
    discount_value NUMERIC(10,2) NOT NULL CHECK (discount_value >= 0),
    start_date TIMESTAMPTZ NOT NULL,
    end_date TIMESTAMPTZ NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Links an offer to specific products
CREATE TABLE product_offers (
    id BIGSERIAL PRIMARY KEY,
    offer_id BIGINT REFERENCES offers(id) ON DELETE CASCADE,
    product_id BIGINT REFERENCES products(id) ON DELETE CASCADE,
    UNIQUE(offer_id, product_id)
);

-- Links an offer to specific categories
CREATE TABLE category_offers (
    id BIGSERIAL PRIMARY KEY,
    offer_id BIGINT REFERENCES offers(id) ON DELETE CASCADE,
    category_id BIGINT REFERENCES categories(id) ON DELETE CASCADE,
    UNIQUE(offer_id, category_id)
);
