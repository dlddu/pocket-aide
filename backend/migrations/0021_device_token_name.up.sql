-- 푸시 기기 이름(사용자가 등록 때 붙이는 별명). 이름 없이 등록된 기존 row 는 빈 문자열.
ALTER TABLE device_tokens ADD COLUMN name TEXT NOT NULL DEFAULT '';
