create table notifications (
  id text primary key,
  user_id text not null references users(id) on delete cascade,
  kind text not null,                 -- order_paid | order_shipped | order_delivered | listing_approved | listing_rejected | ai_done
  title text not null,
  body text not null default '',
  link text not null default '',      -- đường dẫn trong ứng dụng, ví dụ /orders/DH-1
  read_at timestamptz,
  created_at timestamptz not null default now()
);
create index notifications_user on notifications (user_id, created_at desc);