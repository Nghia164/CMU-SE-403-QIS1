-- Mỗi lần đổi/đặt lại mật khẩu tăng token_version; access token mang phiên bản cũ bị từ chối.
alter table users add column token_version int not null default 0;