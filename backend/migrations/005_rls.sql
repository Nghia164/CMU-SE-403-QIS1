-- Supabase tự công khai mọi bảng trong schema public qua API REST (PostgREST) bằng anon key.
-- Bật RLS và KHÔNG tạo policy nào: anon/authenticated không đọc hay ghi được gì.
-- Backend kết nối bằng role chủ sở hữu (postgres), bỏ qua RLS, nên không bị ảnh hưởng.
do $$
declare t text;
begin
  for t in select tablename from pg_tables where schemaname = 'public' loop
    execute format('alter table public.%I enable row level security', t);
  end loop;
end $$;

-- Bảng tạo sau này cũng phải bật RLS: thêm vào migration của bảng đó (alter table ... enable row level security).