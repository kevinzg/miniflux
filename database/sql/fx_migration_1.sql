alter table feeds add column if not exists score_extractor text default '';
alter table entries add column if not exists score int default 0;
alter table entries add column if not exists original_id text default '';
