-- Dùng chung bảng refresh_tokens cho cả token đặt lại mật khẩu (kind = 'reset').
alter table refresh_tokens add column kind text not null default 'refresh' check (kind in ('refresh','reset'));
create index refresh_tokens_user on refresh_tokens (user_id, kind);