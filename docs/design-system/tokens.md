---
type: design-system-tokens
last_updated: 2026-05-06
source: docs/mockups/_template.md + 10 mockup HTML 인라인 스타일 추출
---

# 디자인 토큰 (Tokens)

> 색상·타이포그래피·스페이싱·디바이스 사양의 단일 정의. mockup과 향후 구현 코드는 이 토큰만 사용하고 임의 값을 지양한다.

본 문서는 기존 mockup의 인라인 `<style>`에서 사용된 값을 추출·정리한 결과다.
값을 변경할 때는 본 문서가 단일 진실 원천이며, 변경 시 모든 mockup HTML과 `_index.md`의 영향 범위를 함께 갱신한다.

---

## 1. 영역별 색상 (Area Palette)

PocketAide의 가장 중요한 토큰은 **영역(area)별 색상 세트**다. 각 영역은 그 영역만의 surface(배경), ink(본문), accent(강조), rule(보더 라인), soft(부드러운 강조 배경)을 가진다. 영역 색상은 V3(영역 분리)을 시각적으로 보장하는 핵심 장치이므로, 영역 간 색상을 섞지 않는다.

### 1.1 개인 (Personal) — 따뜻한 점토 (Warm Clay)
- `--bg` 표면: `#FBF1EA`
- `--ink` 본문: `#3D2A22` (헤더 라벨에서는 `#2A1714`도 허용)
- `--clay` 강조: `#B65A3C`
- `--rule` 보더: `#EBD9CB`
- `--soft` 부드러운 강조: `#F4E2D4`

### 1.2 회사 (Work) — 차가운 슬레이트 (Cool Slate)
- `--bg` 표면: `#EEF2F8`
- `--ink` 본문: `#1E2A3A` (또는 `#0E1A2B`)
- `--slate` 강조: `#355577`
- `--rule` 보더: `#D6DEE9`
- `--soft` 부드러운 강조: `#DDE5F0`

### 1.3 AI 채팅 — 차콜 + 세이지
- `--paper` 표면: `#FAFAF7`
- `--ink` 본문: `#1C2624`
- `--sage` 강조: `#5E8B73`
- 보조: `--clay #B65A3C` (사용자 액션 보조 강조)
- 채팅 버블 보더: `#ECEAE3`

### 1.4 임시공간 (Scratchpad) — 종이 + 따뜻한 갈색
- `--paper` 표면: `#F5EFE0`
- `--ink` 본문: `#2A2723`
- `--warm` 강조: `#B6855E`
- `--rule` 보더: `#E0D8C2`
- 카드 배경: `#FBF7EC`

### 1.5 루틴 (Routines) — 숲
- `--bg` 표면: `#F0F2EC`
- `--ink` 본문: `#243329` (또는 `#1C2A22`)
- `--forest` 강조: `#4F6E5C`
- `--rule` 보더: `#D6DDD2`
- `--soft` 부드러운 강조: `#E0E7DA`

### 1.6 다짐 (Affirmations) — 모래/탠
- `--bg` 표면: `#F4EBDD`
- `--ink` 본문: `#2E251A` (또는 `#3A2E1E`)
- `--tan` 강조: `#8B6F47`
- `--rule` 보더: `#E5D7C0`
- `--soft` 부드러운 강조: `#EADCC2`

### 1.7 음성 모드 (다크 변형, AI 채팅의 카운터파트)
- `--bg` 다크 표면: `#0F1614`
- `--ink` 본문: `#FFFFFF`
- `--sage` 강조: `#5E8B73` (라이트와 동일)

### 1.8 시스템 통합 영역 (System-Integrated)
잠금 화면 단축어, iOS 위젯, 키보드 확장은 **iOS 시스템 컨벤션**을 차용하므로 본 디자인 시스템의 영역 토큰을 따르지 않는다. 단, 앱이 노출하는 강조색은 AI 채팅의 `--sage`나 해당 데이터 출처 영역의 강조를 사용한다.

키보드 확장 보조 토큰:
- `--kbd` 키보드 배경: `#D8D3C7`
- `--key` 키 캡: `#FBFAF6`

---

## 2. 중립 토큰 (Neutral)

영역과 무관하게 쓰이는 회색·검정 계열. iOS 컨벤션과 Tailwind `stone` 팔레트에 의존한다.

- 다이나믹 아일랜드 / 홈 인디케이터: `#0a0a0a`
- 디바이스 베젤 outer: `#2a2a2a` / inner ring: `#0a0a0a`
- 페이지 배경(목업 갤러리): `bg-stone-200/70`
- 보조 텍스트: `text-stone-500` / `text-stone-400`
- 흐린 보더: `border-stone-200/70`, `border-stone-300/80`
- 흰색 카드 배경: `#FFFFFF` (영역에 따라 약간의 톤 가미는 영역의 카드 배경 토큰 사용)

영역 토큰이 정의된 곳에서는 stone 팔레트보다 영역 토큰을 우선한다. stone은 어디에도 속하지 않는 메타 영역(목업 외부 nav, 페이지 배경 등)에서만 쓴다.

---

## 3. 타이포그래피

### 3.1 폰트 패밀리

```
'Apple SD Gothic Neo', 'SF Pro Text', -apple-system, BlinkMacSystemFont, system-ui, sans-serif
```

다짐(Affirmations) 영역만 serif 변형 사용을 허용한다 (V4 의도된 반복 노출의 정서적 톤).

### 3.2 전역 타이포 규칙
- `letter-spacing: -0.01em` (전체 화면 기본)
- `-webkit-font-smoothing: antialiased`
- 숫자 정렬이 필요한 곳: `tabular-nums` (예: 진행률 %)

### 3.3 사이즈 스케일

mockup에서 실제 사용된 사이즈만 등재한다. 새로운 사이즈가 필요하면 가까운 토큰으로 우선 매핑한다.

| 토큰 | 크기 | 용도 |
|------|------|------|
| caption-2xs | 10 / 10.5 | 탭바 라벨, 디스클레이머, 미세 메타 |
| caption-xs | 11 | 메타 텍스트, 상단 영역 라벨, 카드 부속 정보 |
| caption-sm | 12 | 칩, 작은 버튼 라벨 |
| body-sm | 13 | 보조 본문 (italic 안내 등) |
| body | 14 | 리스트 부속 본문 |
| body-lg | 15 | 메인 본문 (채팅 메시지, 투두 항목) |
| title-md | 16 | 카드 타이틀 (루틴 카드 등) |
| status | 15 (semibold) | 상태바 시간 |
| h2 | 22 | 화면 H1 (채팅 화면) |
| h1 | 26~28 (bold) | 영역 메인 타이틀 |

### 3.4 가중치
- `font-medium` (500), `font-semibold` (600), `font-bold` (700)
- 영역 라벨(예: "PERSONAL")은 항상 `font-bold` + `tracking-[0.18em~0.22em]` + `uppercase`

---

## 4. 스페이싱 / 라운드

### 4.1 컨테이너 패딩
- 화면 가로 패딩: `px-5` (20px)
- 헤더 위/아래: `pt-3 pb-3`
- 메인 영역과 헤더 사이: 헤더 높이 = 54(상태바) + 약 86(헤더) = `top-[140px]` 부근

### 4.2 라운드
| 용도 | 값 |
|------|-----|
| 디바이스 frame | 56px |
| 다이나믹 아일랜드 | 20px |
| 홈 인디케이터 | 3px |
| 카드 (큰) | 24px (`rounded-3xl`) |
| 카드 (기본 task) | 16px (`rounded-2xl`) |
| 채팅 버블 | 20px (꼬리 쪽 코너만 `rounded-br-md`/`rounded-bl-md`) |
| 칩/필 | full (`rounded-full`) |
| 키 캡 | 6px |
| 입력 필드 | 24px (`rounded-3xl`) |

### 4.3 갭 / 간격
- 카드 간격(리스트): `space-y-1.5`
- 채팅 버블 간격: `space-y-3.5`
- 섹션 간격: `space-y-4`

---

## 5. 그림자 / 스트로크

### 5.1 디바이스 frame 그림자 (고정)
```
0 30px 80px -20px rgba(28,38,36,.25),
0 0 0 12px #0a0a0a,
0 0 0 13px #2a2a2a
```
디바이스 베젤은 12px 검정 + 1px 진회색 outer ring. 절대 변경하지 않는다.

### 5.2 강조 링 (focused)
```
.ringy { box-shadow: 0 0 0 4px rgba(94,139,115,.16); }
```
음성 진입 버튼 등에 사용. 영역 강조색의 16% 알파를 4px 링으로.

---

## 6. 디바이스 / 레이아웃 사양

| 요소 | 사양 |
|------|------|
| Frame | 393 × 852, radius 56px |
| Status bar | 높이 54px, 시간 `9:41` 좌측 / 신호·와이파이·배터리 우측 |
| Dynamic Island | 118 × 35, radius 20, 상단 중앙 (`top-3`) |
| Home indicator | 140 × 5, radius 3, 하단 중앙 (`bottom-2`) |
| Tab bar | 높이 88px (포함 `pb-7` 안전영역), 6탭 균등 배치 |
| Header | 상단 status bar 바로 아래 시작, 영역 라벨 + 제목 + 액션 |

---

## 7. 탭바 구성 (전역 고정)

탭 순서·아이콘은 모든 영역 화면에서 동일하다.

| # | 탭 | 아이콘 키 | 활성 시 색 |
|---|-----|----------|-----------|
| 1 | 채팅 | message bubble (filled when active) | 영역의 ink |
| 2 | 임시공간 | document/file | 영역의 ink |
| 3 | 개인 | person | 영역의 강조색 (개인 탭일 때) |
| 4 | 회사 | briefcase | 영역의 강조색 (회사 탭일 때) |
| 5 | 루틴 | refresh loop | 영역의 강조색 (루틴 탭일 때) |
| 6 | 다짐 | heart | 영역의 강조색 (다짐 탭일 때) |

**규칙**: 활성 탭은 해당 영역의 강조색 또는 ink를 사용하고, 비활성 탭은 `text-stone-400` (또는 영역 ink의 50% 투명도)을 사용한다.

---

## 8. 사용 규칙 (간단)

- mockup에서 임의의 hex 값을 직접 쓰지 않는다. 영역 토큰 또는 중립 토큰 안에서 고른다.
- 영역 화면에서는 stone 팔레트를 본문 텍스트의 흐린 색에만 한정한다.
- 새 색이 필요하다고 느끼면 먼저 "이게 정말 새 색인가, 기존 영역의 변형인가"를 묻는다.
- 변경은 본 문서를 먼저 수정한 뒤 mockup에 반영한다.
