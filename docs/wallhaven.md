# Wallhaven API — observed behavior and fixtures

Checked 2026-09-08 on macOS 26.6.2 against the live endpoint and `https://wallhaven.cc/help/api`.
Fixtures: `internal/source/testdata/`.

## Endpoint and parameters (observed)

`GET https://wallhaven.cc/api/v1/search` with `q`, `categories` (3-bit `general|anime|people`,
e.g. `010`), `purity` (`sfw|sketchy|nsfw`, e.g. `100`; NSFW needs a valid key), `sorting`
(`relevance|date_added|random|views|favorites|toplist`), `order`, `atleast`, `ratios`, `page`,
`seed`, `colors`. No params → latest SFW, 24 per page.

## Response shape (observed)

Envelope `{ "data": [...], "meta": {...} }`. `data[]`: `id`, `url`, `short_url`, `views`,
`favorites`, `source`, `purity`, `category`, `dimension_x`, `dimension_y`, `resolution`, `ratio`,
`file_size`, `file_type`, `created_at`, `colors[]` (5 `#RRGGBB`), `path` (full image on
`w.wallhaven.cc`), `thumbs` (`large`, `original`, `small` on `th.wallhaven.cc`). `meta`:
`current_page`, `last_page`, `per_page`, `total`, `query`, `seed` (nullable).

## Keyless path (observed)

A search with no `apikey` parameter works for public/SFW listings (HTTP 200). This is the MVP
default; a key only lifts NSFW access / limits.

## Rate limiting and errors

- Documented (help page): **"API calls are currently limited to 45 per minute... a 429 - Too
  many requests error."** Observed headers: `x-ratelimit-limit: 45`, `x-ratelimit-remaining: 44`.
- **Correction:** the explore.md "per-30s download limit" is NOT confirmed; the help page
  documents only 45 API calls/min, and image hosts return no `x-ratelimit-*` headers.
- Auth failure (observed): invalid `apikey` → HTTP 401 `{"error":"Unauthorized"}`.
- Rate-limit body fixture is synthesized `{"error":"Rate limit reached"}` (429 body not
  live-triggered; HTTP 429 + rate headers are the contract).
- Empty results (observed): HTTP 200, `{"data":[],"meta":{...,"total":0,...}}`.

## Download (observed)

Thumb `https://th.wallhaven.cc/small/p2/p29ke3.jpg` → 200, 22432 bytes, `image/jpeg`, 300x200
(checked in as `wallhaven_download_small.jpg`).

## Attribution

Wallhaven requires credit for downloaded art; surface at least `url` in results. Local images
need no Wallhaven attribution.

## Fixture map

`wallhaven_search.json` (live `q=nature`, 2 items) · `wallhaven_search_nokey.json` (live keyless
toplist, 2 items) · `wallhaven_search_empty.json` (live no-results) · `wallhaven_error_auth.json`
(live 401) · `wallhaven_error_ratelimit.json` (synthesized 429) · `wallhaven_error_malformed.json`
(synthesized truncated) · `wallhaven_download_small.jpg` (live thumb).
