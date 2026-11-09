-- Lược đồ Postgres (Supabase) theo diagrams/06-erd.drawio.
-- ID là text do ứng dụng sinh (tiền tố u_, l_, DH-, BC-, job_) nên đọc được trong log và URL.
create extension if not exists vector;

create table users (
  id text primary key,
  name text not null,
  email text not null,
  password_hash text not null,
  role text not null check (role in ('admin','seller','buyer')),
  status text not null default 'active' check (status in ('active','locked','pending')),
  failed_logins int not null default 0,
  locked_until timestamptz,
  created_at timestamptz not null default now()
);
create unique index users_email_lower on users (lower(email));

create table refresh_tokens (
  id text primary key,
  user_id text not null references users(id) on delete cascade,
  token_hash text not null unique,
  expires_at timestamptz not null,
  revoked_at timestamptz
);

create table categories (
  value text primary key,
  label text not null,
  icon text not null
);
insert into categories (value, label, icon) values
  ('dien-tu','Điện tử','smartphone'), ('xe','Xe cộ','bike'), ('nha','Nhà cửa','sofa'),
  ('thoi-trang','Thời trang','shirt'), ('sach','Sách','book-open'), ('khac','Khác','package');

create table listings (
  id text primary key,
  seller_id text not null references users(id),
  seller_name text not null,
  category text not null references categories(value),
  icon text not null default 'package',
  title text not null check (char_length(title) between 5 and 120),
  description text not null default '' check (char_length(description) <= 2000),
  search_text text not null default '',          -- tiêu đề + mô tả đã bỏ dấu, để tìm không dấu
  price bigint not null check (price > 0),
  original_price bigint not null default 0,
  condition text not null check (condition in ('Mới','Như mới','Tốt','Khá','Cũ')),
  location text not null default '',
  hashtags text[] not null default '{}',
  status text not null default 'pending_review'
    check (status in ('pending_review','active','rejected','sold','hidden')),
  ai_verdict text not null default '',
  ai_score int not null default 0,
  ai_generated boolean not null default false,
  embedding vector(768),
  created_at timestamptz not null default now()
);
create index listings_browse on listings (status, category, created_at desc);

create table media_assets (
  id text primary key,
  listing_id text references listings(id) on delete cascade,
  owner_id text not null references users(id),
  r2_key text not null,
  position int not null default 0
);

create table ai_jobs (
  id text primary key,
  seller_id text not null references users(id),
  status text not null check (status in ('QUEUED','RUNNING','SUCCEEDED','FAILED')),
  progress int not null default 0,
  steps jsonb not null default '[]',
  result jsonb,
  error text not null default '',
  created_at timestamptz not null default now()
);

create table comparables (
  id text primary key,
  ai_job_id text not null references ai_jobs(id) on delete cascade,
  source text not null,
  title text not null,
  price bigint not null,
  condition text
);

create table suggestion_history (
  id text primary key,
  ai_job_id text not null references ai_jobs(id),
  listing_id text references listings(id),
  suggested jsonb not null,
  final jsonb,                       -- nội dung người bán thực sự đăng
  created_at timestamptz not null default now()
);

create table orders (
  id text primary key,
  listing_id text not null references listings(id),
  listing_title text not null,
  buyer_id text not null references users(id),
  seller_id text not null references users(id),
  amount bigint not null check (amount > 0),
  method text not null check (method in ('vnpay','momo')),
  status text not null check (status in ('pending_payment','paid','shipping','delivered','cancelled','refunded')),
  receiver_name text not null,
  phone text not null,
  address text not null,
  events jsonb not null default '{}',
  payment_ref text not null default '',
  created_at timestamptz not null default now()
);
-- Mỗi tin chỉ có tối đa một đơn đã thanh toán (chống bán trùng ngay ở tầng CSDL).
create unique index one_paid_order_per_listing on orders (listing_id)
  where status in ('paid','shipping','delivered');
create index orders_buyer on orders (buyer_id, created_at desc);
create index orders_seller on orders (seller_id, created_at desc);

create table payment_transactions (
  id text primary key,
  order_id text not null references orders(id),
  provider text not null check (provider in ('vnpay','momo')),
  provider_txn_id text not null,
  amount bigint not null,
  status text not null,
  raw jsonb not null,
  created_at timestamptz not null default now(),
  unique (provider, provider_txn_id)  -- idempotent khi cổng gọi IPN lại
);

create table reports (
  id text primary key,
  reporter_id text not null references users(id),
  target_type text not null check (target_type in ('listing','user')),
  target_id text not null,
  target_title text not null default '',
  reason text not null,
  status text not null default 'new' check (status in ('new','progress','done')),
  action text not null default '',
  note text not null default '',
  created_at timestamptz not null default now()
);

create table audit_logs (
  id text primary key,
  actor_id text not null references users(id),
  action text not null,
  target text not null,
  created_at timestamptz not null default now()
);
