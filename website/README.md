# Rasanta Frontend (Nuxt 4)

Frontend Rasanta adalah web app untuk eksplorasi tempat/restaurant berbasis peta, filter, sorting, dan detail place.

## Stack

- Nuxt 4
- Vue 3 + TypeScript
- Pinia
- Nuxt UI
- Leaflet

## Prasyarat

- Node.js 20+
- pnpm (repo ini menggunakan `pnpm@10`)
- Backend Rasanta aktif di `http://localhost:8000`

## Instalasi

```bash
pnpm install
```

## Menjalankan Frontend

Development server (default port frontend: `3000`):

```bash
pnpm dev
```

Setelah jalan, buka:

- `http://localhost:3000`

## Konfigurasi API Backend

Secara default frontend memanggil backend ke path relatif yang sama domain (`/api/v1/...`).

Untuk local development, set base URL backend via env:

```bash
cp .env.example .env
```

## Script yang Tersedia

```bash
pnpm dev        # jalankan development server
pnpm build      # build production
pnpm preview    # preview hasil build production
pnpm lint       # linting
pnpm typecheck  # type checking (nuxt typecheck)
```

## Fitur Utama

- Pemilihan lokasi dari browser geolocation dan fallback IP location
- Query preset (Restaurant, Cafe, Bakso, Mie, dan lainnya)
- Smart Ranking (SAW) toggle
- Filter chips (price, open now, delivery, takeout, halal, alcohol)
- Sorting (recommended, rating, reviews, distance)
- Lazy prefetch detail:
  - Saat card restoran masuk viewport list
  - Saat pin di map di-hover

## Struktur Folder Inti

- `app/pages/index.vue` - halaman utama (state orchestration)
- `app/stores/restaurantMap.ts` - fetch list/detail dan cache
- `app/components/maps/LeafletMap.vue` - peta dan pin interaksi
- `app/components/maps/RestaurantExplorerPanel.vue` - list, filter, sort, detail panel
- `app/composables/useUserLocation.ts` - geolocation + fallback IP
- `i18n/locales` - terjemahan Indonesia dan English

## Catatan Pengembangan

- Frontend memanggil endpoint backend:
  - `GET /api/v1/places`
  - `GET /api/v1/places/detail`
- Proxy lokal IP location tersedia di route frontend:
  - `GET /api/ip-location`

## License

Proyek ini dilisensikan di bawah [MIT License](../LICENSE).
