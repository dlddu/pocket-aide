-- PRD-10 AC13: 그룹핑 키로 사용할 head_sha 보존. 기존 row는 빈 문자열로 채워지며
-- 클라이언트가 "PR도 head_sha도 없는 행" 케이스에서 단독 그룹으로 흡수한다.
ALTER TABLE notification_history ADD COLUMN head_sha TEXT NOT NULL DEFAULT '';
