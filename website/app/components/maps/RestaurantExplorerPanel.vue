<script setup lang="ts">
import type { Coordinate, CoordinateSource } from "~/types/location";
import type {
  RestaurantDetail,
  RestaurantListFilter,
  RestaurantPin,
  RestaurantSortOption,
} from "~/types/restaurant";
import type { QueryPreset } from "~/config/places";

const props = defineProps<{
  coordinate: Coordinate;
  source: CoordinateSource;
  label: string;
  status: string;
  loading: boolean;
  restaurantsLoading: boolean;
  restaurantsError: string;
  restaurants: RestaurantPin[];
  restaurantCount: number;
  sawEnabled: boolean;
  activeFilter: RestaurantListFilter;
  sortBy: RestaurantSortOption;
  activeQuery: string;
  queryPresets: QueryPreset[];
  selectedRestaurantId: string | null;
  detail: RestaurantDetail | null;
  detailLoading: boolean;
  detailError: string;
  isDetailOpen: boolean;
}>();

const emit = defineEmits<{
  locateMe: [];
  openDetail: [value: RestaurantPin];
  closeDetail: [];
  toggleSidebar: [];
  toggleSaw: [];
  changeFilter: [value: RestaurantListFilter];
  changeSort: [value: RestaurantSortOption];
  changeQuery: [value: string];
  prefetchDetail: [value: RestaurantPin];
}>();

const { t } = useAppI18n();

const sourceLabel = computed(() => {
  if (props.source === "browser") {
    return t("source.browser");
  } else if (props.source === "ip") {
    return t("source.ip");
  } else if (props.source === "search") {
    return t("source.search");
  }

  return t("source.default");
});

const activeQueryLabel = computed(() => {
  const preset = props.queryPresets.find((item) => item.value === props.activeQuery);
  if (preset?.label) {
    return preset.label;
  }

  const fallback = props.activeQuery.trim();
  if (fallback.length > 0) {
    return fallback;
  }

  return t("panel.restaurants");
});

const headline = computed(() => {
  const base = props.label.split(",")[0]?.trim() || t("panel.thisArea");
  return t("panel.headline", {
    count: Math.max(props.restaurantCount, 1),
    area: base,
    query: activeQueryLabel.value,
  });
});

const chips = computed(() => {
  return [
    {
      key: "all" as const,
      label: t("chip.all"),
    },
    {
      key: "price" as const,
      label: t("chip.price"),
    },
    {
      key: "open_now" as const,
      label: t("chip.openNow"),
    },
    {
      key: "delivery" as const,
      label: t("chip.delivery"),
    },
    {
      key: "takeout" as const,
      label: t("chip.takeout"),
    },
    {
      key: "halal" as const,
      label: t("chip.halal"),
    },
    {
      key: "alcohol" as const,
      label: t("chip.alcohol"),
    },
  ];
});

const sortOptions = computed(() => {
  return [
    {
      key: "recommended" as const,
      label: t("sortOption.recommended"),
    },
    {
      key: "rating_desc" as const,
      label: t("sortOption.ratingDesc"),
    },
    {
      key: "reviews_desc" as const,
      label: t("sortOption.reviewsDesc"),
    },
    {
      key: "distance_asc" as const,
      label: t("sortOption.distanceAsc"),
    },
  ];
});

const sawLabel = computed(() => {
  return props.sawEnabled ? t("panel.sawEnabled") : t("panel.sawDisabled");
});

let cardObserver: IntersectionObserver | null = null;
const observedCards = new WeakMap<HTMLElement, RestaurantPin>();

function observeCard(el: HTMLElement | null, restaurant: RestaurantPin) {
  if (!cardObserver || !el) {
    return;
  }
  observedCards.set(el, restaurant);
  cardObserver.observe(el);
}

onMounted(() => {
  if (!import.meta.client || typeof IntersectionObserver === 'undefined') {
    return;
  }
  cardObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (!entry.isIntersecting) {
          continue;
        }
        const restaurant = observedCards.get(entry.target as HTMLElement);
        if (restaurant) {
          emit('prefetchDetail', restaurant);
        }
        cardObserver?.unobserve(entry.target);
      }
    },
    { threshold: 0.25 }
  );
});

onUnmounted(() => {
  cardObserver?.disconnect();
  cardObserver = null;
});

const detailSummary = computed(() => props.detail?.summary ?? null);

const detailRaw = computed(() => asRecord(props.detail?.raw));

const detailHeroThumbnail = computed(() => detailSummary.value?.thumbnail || null);

const favoriteRestaurantIds = useState<string[]>("favorite-restaurant-ids", () => []);

const detailTagline = computed(() => {
  if (!detailSummary.value) {
    return "";
  }

  return [detailSummary.value.type, detailSummary.value.price]
    .filter((item): item is string => Boolean(item))
    .join(", ");
});

const detailDirectionLink = computed(() => {
  if (!detailSummary.value) {
    return null;
  }

  return `https://www.google.com/maps/search/?api=1&query=${detailSummary.value.lat},${detailSummary.value.lng}`;
});

const detailIsClosed = computed(() => {
  const value = detailSummary.value?.openState?.toLowerCase() || "";
  return value.includes("closed") || value.includes("tutup");
});

const detailStatusClass = computed(() => {
  if (!detailSummary.value?.openState) {
    return "bg-slate-700/70 text-white";
  }

  return detailIsClosed.value
    ? "bg-red-500/90 text-white"
    : "bg-emerald-500/90 text-white";
});

const detailServiceOptions = computed(() => {
  const serviceOptions = asRecord(detailRaw.value.service_options);
  return Object.entries(serviceOptions)
    .filter(([, enabled]) => Boolean(enabled))
    .map(([key]) => mapServiceOptionLabel(key));
});

const detailExtensionGroups = computed(() => {
  return mergeExtensionGroups(
    detailRaw.value.extensions ?? detailRaw.value.extension ?? detailRaw.value.feature_extensions,
    detailRaw.value.unsupported_extensions
      ?? detailRaw.value.unsupportedExtensions
      ?? detailRaw.value.unavailable_extensions
  );
});

const detailOperatingHours = computed(() => {
  return normalizeHoursList(
    detailRaw.value.operating_hours
      ?? detailRaw.value.operatingHours
      ?? detailRaw.value.opening_hours
      ?? detailRaw.value.openingHours
  );
});

const detailReviewSnippets = computed(() => {
  const values = asArray(detailRaw.value.review_snippets);
  return values
    .map((item) => toText(item))
    .filter((item): item is string => Boolean(item))
    .slice(0, 6);
});

const detailGallery = computed(() => {
  const photos = asRecord(detailRaw.value.photos);
  const items = asArray(photos.items);

  const collected = items
    .map((item) => asRecord(item))
    .map((item) => {
      const url = toText(item.url);
      if (!url) {
        return null;
      }

      return {
        url,
        caption: toText(item.caption),
      };
    })
    .filter((item): item is { url: string; caption: string | null } => Boolean(item));

  if (
    detailSummary.value?.thumbnail &&
    !collected.some((item) => item.url === detailSummary.value?.thumbnail)
  ) {
    collected.unshift({
      url: detailSummary.value.thumbnail,
      caption: detailSummary.value.title,
    });
  }

  return collected.slice(0, 12);
});

function mapServiceOptionLabel(key: string): string {
  if (key === "dine_in") {
    return t("panel.serviceDineIn");
  }
  if (key === "takeout") {
    return t("chip.takeout");
  }
  if (key === "delivery") {
    return t("chip.delivery");
  }

  return key.replaceAll("_", " ");
}

function mergeExtensionGroups(
  supportedInput: unknown,
  unsupportedInput: unknown
): Array<{ key: string; label: string; supported: string[]; unsupported: string[] }> {
  const supportedMap = extractExtensionCategoryMap(supportedInput);
  const unsupportedMap = extractExtensionCategoryMap(unsupportedInput);

  const orderedKeys: string[] = [];

  Object.keys(supportedMap).forEach((key) => {
    if (!orderedKeys.includes(key)) {
      orderedKeys.push(key);
    }
  });

  Object.keys(unsupportedMap).forEach((key) => {
    if (!orderedKeys.includes(key)) {
      orderedKeys.push(key);
    }
  });

  return orderedKeys
    .map((key) => ({
      key,
      label: humanizeToken(key),
      supported: supportedMap[key] || [],
      unsupported: unsupportedMap[key] || [],
    }))
    .filter((group) => group.supported.length > 0 || group.unsupported.length > 0);
}

function extractExtensionCategoryMap(input: unknown): Record<string, string[]> {
  const map: Record<string, string[]> = {};

  if (!input) {
    return map;
  }

  const addValues = (categoryKey: string, values: string[]) => {
    if (!values.length) {
      return;
    }

    if (!map[categoryKey]) {
      map[categoryKey] = [];
    }

    values.forEach((value) => {
      if (!map[categoryKey]?.includes(value)) {
        map[categoryKey]?.push(value);
      }
    });
  };

  const parseEntry = (entry: unknown) => {
    if (!entry) {
      return;
    }

    if (Array.isArray(entry)) {
      entry.forEach((innerEntry) => parseEntry(innerEntry));
      return;
    }

    if (typeof entry === "object") {
      const record = asRecord(entry);
      Object.entries(record).forEach(([categoryKey, value]) => {
        addValues(categoryKey, normalizeDetailValueList(value));
      });
      return;
    }

    const text = toText(entry);
    if (text) {
      addValues("general", [text]);
    }
  };

  parseEntry(input);

  return map;
}

function normalizeDetailValueList(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value
      .map((entry) => toText(entry))
      .filter((entry): entry is string => Boolean(entry));
  }

  const text = toText(value);
  if (text) {
    return [text];
  }

  if (value && typeof value === "object") {
    const record = asRecord(value);
    return Object.values(record)
      .map((entry) => toText(entry))
      .filter((entry): entry is string => Boolean(entry));
  }

  return [];
}

function normalizeHoursList(input: unknown): string[] {
  if (!input) {
    return [];
  }

  if (Array.isArray(input)) {
    return input
      .map((item) => normalizeHourItem(item))
      .filter((item): item is string => Boolean(item));
  }

  if (typeof input === "object") {
    const record = asRecord(input);
    return Object.entries(record)
      .map(([key, value]) => {
        const valueText = normalizeHourItem(value) || toText(value);
        if (!valueText) {
          return null;
        }

        return `${humanizeToken(key)}: ${valueText}`;
      })
      .filter((item): item is string => Boolean(item));
  }

  return [];
}

function normalizeHourItem(item: unknown): string | null {
  const directText = toText(item);
  if (directText) {
    return directText;
  }

  if (Array.isArray(item)) {
    const [day, hours] = item;
    const dayText = toText(day);
    const hoursText = toText(hours);

    if (dayText && hoursText) {
      return `${dayText}: ${hoursText}`;
    }

    return dayText || hoursText;
  }

  if (item && typeof item === "object") {
    const record = asRecord(item);

    const day = toText(record.day) || toText(record.label) || toText(record.name);
    const hours =
      toText(record.hours)
      || toText(record.value)
      || toText(record.open)
      || toText(record.close);

    if (day && hours) {
      return `${day}: ${hours}`;
    }

    return day || hours || null;
  }

  return null;
}

function humanizeToken(token: string): string {
  return token
    .replaceAll("_", " ")
    .replace(/\b\w/g, (char) => char.toUpperCase());
}

function toggleDetailFavorite() {
  const id = detailSummary.value?.id;
  if (!id) {
    return;
  }

  if (favoriteRestaurantIds.value.includes(id)) {
    favoriteRestaurantIds.value = favoriteRestaurantIds.value.filter(
      (item) => item !== id
    );
    return;
  }

  favoriteRestaurantIds.value = [...favoriteRestaurantIds.value, id];
}

async function shareDetail() {
  if (!import.meta.client || !detailSummary.value) {
    return;
  }

  const shareUrl = detailDirectionLink.value || window.location.href;

  if (typeof navigator !== "undefined" && typeof navigator.share === "function") {
    try {
      await navigator.share({
        title: detailSummary.value.title,
        text: detailSummary.value.address || detailSummary.value.title,
        url: shareUrl,
      });
      return;
    } catch {
      // Fallback to clipboard when share is cancelled or unsupported by browser.
    }
  }

  if (typeof navigator !== "undefined" && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(shareUrl);
    } catch {
      // No-op when clipboard permission is unavailable.
    }
  }
}

function asRecord(input: unknown): Record<string, unknown> {
  if (input && typeof input === "object" && !Array.isArray(input)) {
    return input as Record<string, unknown>;
  }

  return {};
}

function asArray(input: unknown): unknown[] {
  if (Array.isArray(input)) {
    return input;
  }

  return [];
}

function toText(input: unknown): string | null {
  if (typeof input === "string") {
    const value = input.trim();
    return value.length ? value : null;
  }

  if (typeof input === "number" && Number.isFinite(input)) {
    return String(input);
  }

  return null;
}

function handleSortChange(event: Event) {
  const nextSort = (event.target as HTMLSelectElement).value;

  if (isSortOption(nextSort)) {
    emit("changeSort", nextSort);
  }
}

function isSortOption(value: string): value is RestaurantSortOption {
  return ["recommended", "rating_desc", "reviews_desc", "distance_asc"].includes(value);
}

function formatRating(restaurant: RestaurantPin): string {
  return restaurant.rating !== null ? restaurant.rating.toFixed(1) : "-";
}

function formatReviewCount(restaurant: RestaurantPin): string {
  return restaurant.reviews !== null
    ? t("panel.reviews", { count: restaurant.reviews })
    : t("panel.noReviews");
}

function formatPlaceMeta(restaurant: RestaurantPin): string {
  const pieces = [
    restaurant.address,
    restaurant.price,
    restaurant.openState,
  ].filter((item): item is string => Boolean(item));
  return pieces.length
    ? pieces.join(" • ")
    : `${restaurant.lat.toFixed(5)}, ${restaurant.lng.toFixed(5)}`;
}
</script>

<template>
  <aside
    class="relative grid min-h-0 grid-rows-[auto_minmax(0,1fr)] gap-4 overflow-hidden overscroll-contain border-r border-slate-200 bg-white px-6 pb-4 pt-7 max-[900px]:h-screen max-[900px]:min-h-screen max-[900px]:border-b max-[900px]:border-r-0 max-[900px]:pb-[calc(env(safe-area-inset-bottom)+12px)] max-[900px]:pt-[calc(env(safe-area-inset-top)+14px)] max-[640px]:px-3.5 max-[640px]:pb-[calc(env(safe-area-inset-bottom)+10px)] max-[640px]:pt-[calc(env(safe-area-inset-top)+10px)] lg:h-screen lg:min-h-screen"
  >
    <header class="grid gap-3">
      <div class="flex items-center justify-between gap-3">
        <p class="m-0 text-[0.95rem] font-medium text-slate-500">
          {{ activeQueryLabel }}
        </p>

        <UButton
          color="neutral"
          variant="soft"
          icon="i-lucide-x"
          class="lg:hidden"
          aria-label="Close sidebar"
          @click="emit('toggleSidebar')"
        >
          Close
        </UButton>
      </div>

      <div>
        <h1 class="m-0 text-[2.4rem] font-semibold tracking-[-0.02em] text-slate-900 max-[900px]:text-[1.9rem] max-[640px]:text-[1.45rem] max-[640px]:leading-[1.15]">
          {{ headline }}
        </h1>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3">
        <div class="flex items-center gap-2">
          <p class="m-0 text-[1.02rem] text-slate-700">
            {{ t("panel.sort") }}
          </p>

          <select
            :value="sortBy"
            class="rounded-full border border-slate-300 bg-white px-3 py-1.5 text-[0.88rem] font-medium text-slate-800 outline-none ring-sky-500 transition focus:ring-2"
            @change="handleSortChange"
          >
            <option
              v-for="option in sortOptions"
              :key="option.key"
              :value="option.key"
            >
              {{ option.label }}
            </option>
          </select>
        </div>

        <div class="flex items-center gap-2">
          <UButton
            color="neutral"
            variant="soft"
            icon="i-lucide-locate-fixed"
            :loading="loading"
            @click="emit('locateMe')"
          >
            {{ t("action.useMyLocation") }}
          </UButton>
        </div>
      </div>

      <div class="flex flex-wrap gap-2">
        <UButton
          color="neutral"
          :variant="sawEnabled ? 'solid' : 'soft'"
          class="rounded-full border border-slate-300 px-3 py-1.5 text-[0.82rem] font-semibold"
          @click="emit('toggleSaw')"
        >
          {{ sawLabel }}
        </UButton>

        <UButton
          v-for="chip in chips"
          :key="chip.key"
          color="neutral"
          :variant="activeFilter === chip.key ? 'solid' : 'soft'"
          class="rounded-full border border-slate-300 px-3 py-1.5 text-[0.82rem]"
          @click="emit('changeFilter', chip.key)"
        >
          {{ chip.label }}
        </UButton>
      </div>

      <div class="-mx-1 flex gap-2 overflow-x-auto px-1 pb-1 [scrollbar-width:none]">
        <UButton
          v-for="preset in queryPresets"
          :key="preset.value"
          color="primary"
          :variant="activeQuery === preset.value ? 'solid' : 'outline'"
          class="shrink-0 rounded-full px-3 py-1 text-[0.8rem] font-medium"
          @click="emit('changeQuery', preset.value)"
        >
          {{ preset.label }}
        </UButton>
      </div>

      <p class="m-0 text-[0.86rem] text-slate-500">
        {{ status }}
      </p>
      <p class="m-0 text-[0.86rem] text-slate-500">
        {{ t("panel.source") }}: {{ sourceLabel }} • {{ coordinate.lat.toFixed(5) }},
        {{ coordinate.lng.toFixed(5) }}
      </p>

      <p v-if="restaurantsError" class="m-0 text-[0.86rem] text-red-700">
        {{ restaurantsError }}
      </p>
    </header>

    <div
      class="scrollbar-soft grid min-h-0 gap-4 overflow-y-auto overscroll-contain pr-1"
    >
      <div v-if="restaurantsLoading" class="grid gap-3">
        <USkeleton class="h-32 w-full rounded-2xl" />
        <USkeleton class="h-32 w-full rounded-2xl" />
        <USkeleton class="h-32 w-full rounded-2xl" />
      </div>

      <p
        v-else-if="restaurants.length === 0"
        class="m-0 rounded-xl border border-dashed border-slate-200 p-4 text-slate-500"
      >
        {{ t("panel.empty", { query: activeQueryLabel }) }}
      </p>

      <button
        v-for="(restaurant, index) in restaurants"
        v-else
        :key="restaurant.id"
        :ref="(el: HTMLElement | null) => observeCard(el, restaurant)"
        type="button"
        class="grid w-full cursor-pointer appearance-none grid-cols-[140px_minmax(0,1fr)] gap-4 rounded-2xl border border-slate-200 bg-white p-3 text-left transition duration-200 hover:-translate-y-px hover:border-sky-300 hover:shadow-[0_18px_38px_rgba(15,23,42,0.08)] max-[640px]:grid-cols-1"
        :class="
          selectedRestaurantId === restaurant.id
            ? 'border-sky-500 ring-2 ring-sky-500/15'
            : ''
        "
        @click="emit('openDetail', restaurant)"
      >
        <div
          class="h-28 overflow-hidden rounded-xl bg-linear-to-br from-slate-100 to-blue-100 max-[640px]:h-39"
        >
          <img
            v-if="restaurant.thumbnail"
            :src="restaurant.thumbnail"
            :alt="restaurant.title"
            class="h-full w-full object-cover"
            loading="lazy"
          />
          <div
            v-else
            class="flex h-full w-full items-center justify-center text-[1.6rem] font-bold text-blue-700"
          >
            {{ index + 1 }}
          </div>
        </div>

        <div class="grid gap-1.5">
          <div class="flex items-start justify-between gap-2.5">
            <h2 class="m-0 text-[1.22rem] leading-[1.2] tracking-[-0.02em]">
              {{ restaurant.title }}
            </h2>
            <span
              class="inline-flex h-7 w-7 items-center justify-center rounded-full bg-red-100 text-[0.82rem] font-bold text-red-600"
            >
              {{ index + 1 }}
            </span>
          </div>

          <p class="m-0 text-[0.9rem] text-slate-500">
            {{ formatPlaceMeta(restaurant) }}
          </p>

          <p
            class="m-0 flex items-center gap-1.5 text-[0.92rem] font-semibold text-slate-900"
          >
            <UIcon name="i-lucide-star" class="h-4 w-4 text-amber-500" />
            <span>{{ formatRating(restaurant) }}</span>
            <span class="font-normal text-slate-500"
              >({{ formatReviewCount(restaurant) }})</span
            >
          </p>

          <p v-if="restaurant.type" class="m-0 text-[0.9rem] text-slate-500">
            {{ restaurant.type }}
          </p>
        </div>
      </button>
    </div>

    <div
      v-if="isDetailOpen"
      class="absolute inset-0 z-50 grid grid-rows-[auto_minmax(0,1fr)] border-l border-slate-200 bg-white"
    >
      <div class="scrollbar-soft min-h-0 overflow-y-auto">
        <div v-if="detailLoading" class="grid gap-3">
          <USkeleton class="h-56 w-full rounded-2xl" />
          <USkeleton class="h-8 w-3/5 rounded-lg" />
          <USkeleton class="h-24 w-full rounded-xl" />
          <USkeleton class="h-32 w-full rounded-xl" />
        </div>

        <p
          v-else-if="detailError"
          class="m-0 rounded-xl border border-red-200 bg-red-50 p-3 text-[0.9rem] text-red-700"
        >
          {{ detailError }}
        </p>

        <div v-if="detailSummary" class="grid gap-4">
          <section
            class="relative overflow-hidden border border-slate-200 bg-slate-900"
          >
            <img
              v-if="detailHeroThumbnail"
              :src="detailHeroThumbnail"
              :alt="detailSummary.title"
              class="h-56 w-full object-cover"
              loading="lazy"
            />
            <div
              v-else
              class="flex h-56 w-full items-center justify-center text-sm font-semibold text-white"
            >
              {{ t("panel.detailNoPhoto") }}
            </div>

            <div
              class="pointer-events-none absolute inset-0 bg-linear-to-b from-slate-900/70 via-slate-900/20 to-slate-900/90"
            />

            <div
              class="absolute inset-x-0 top-0 flex items-start justify-between gap-2 px-3 py-3"
            >
              <div class="flex items-center gap-2">
                <UButton
                  icon="i-lucide-arrow-left"
                  color="neutral"
                  variant="ghost"
                  size="sm"
                  aria-label="Back to list"
                  @click="emit('closeDetail')"
                >
                  Back
                </UButton>
              </div>
            </div>

            <div class="absolute inset-x-0 bottom-0 grid gap-2 px-3 py-3 text-white">
              <h3 class="m-0 capitalize text-[1.35rem] leading-[1.12] tracking-[-0.02em]">
                {{ detailSummary.title }}
              </h3>

              <div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-[0.86rem]">
                <span class="inline-flex items-center gap-1.5 font-semibold">
                  <UIcon name="i-lucide-star" class="h-4 w-4 text-amber-300" />
                  {{ formatRating(detailSummary) }}
                </span>
                <span>{{ formatReviewCount(detailSummary) }}</span>
                <span v-if="detailTagline">{{ detailTagline }}</span>
              </div>

              <p class="m-0 line-clamp-2 text-[0.84rem] text-white/90">
                {{ detailSummary.address || "-" }}
              </p>
            </div>
          </section>

          <div class="grid gap-4 px-4 pb-5 pt-4">
            <section class="grid content-start gap-3">
              <section
                class="grid gap-2 rounded-2xl border border-slate-200 bg-slate-50 p-3"
              >
                <p
                  class="m-0 text-xs font-semibold uppercase tracking-[0.12em] text-slate-500"
                >
                  {{ t("panel.detailInformation") }}
                </p>
                <p class="m-0 text-[0.92rem] text-slate-700">
                  {{ detailSummary.phone || t("panel.noPhone") }}
                </p>
                <a
                  v-if="detailSummary.website"
                  :href="detailSummary.website"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="inline-flex w-fit text-[0.86rem] font-semibold text-sky-700"
                >
                  {{ t("panel.openWebsite") }}
                </a>
              </section>

              <section v-if="detailServiceOptions.length" class="grid gap-2">
                <p
                  class="m-0 text-xs font-semibold uppercase tracking-[0.12em] text-slate-500"
                >
                  {{ t("panel.detailServices") }}
                </p>
                <div class="flex flex-wrap gap-2">
                  <UBadge
                    v-for="service in detailServiceOptions"
                    :key="service"
                    color="primary"
                    variant="soft"
                    class="rounded-full"
                  >
                    {{ service }}
                  </UBadge>
                </div>
              </section>

              <section class="grid gap-2">
                <p
                  class="m-0 text-xs font-semibold uppercase tracking-[0.12em] text-slate-500"
                >
                  {{ t("panel.detailExtensions") }}
                </p>
                <div
                  v-if="detailExtensionGroups.length"
                  class="grid gap-2"
                >
                  <div
                    v-for="group in detailExtensionGroups"
                    :key="group.key"
                    class="grid gap-2 rounded-xl border border-slate-200 bg-white p-3"
                  >
                    <p class="m-0 text-[0.86rem] font-semibold text-slate-900">
                      {{ group.label }}
                    </p>

                    <div
                      v-if="group.supported.length"
                      class="flex flex-wrap items-center gap-2"
                    >
                      <UBadge
                        v-for="item in group.supported"
                        :key="`${group.key}-supported-${item}`"
                        color="primary"
                        variant="soft"
                        class="rounded-full"
                      >
                        {{ item }}
                      </UBadge>
                    </div>

                    <div
                      v-if="group.unsupported.length"
                      class="flex flex-wrap items-center gap-2"
                    >
                      <UBadge
                        v-for="item in group.unsupported"
                        :key="`${group.key}-unsupported-${item}`"
                        color="error"
                        variant="soft"
                        class="rounded-full"
                      >
                        {{ item }}
                      </UBadge>
                    </div>
                  </div>
                </div>
                <p
                  v-else
                  class="m-0 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-3 py-2 text-[0.9rem] text-slate-600"
                >
                  {{ t("panel.noData") }}
                </p>
              </section>

              <section class="grid gap-2">
                <p
                  class="m-0 text-xs font-semibold uppercase tracking-[0.12em] text-slate-500"
                >
                  {{ t("panel.detailOperatingHours") }}
                </p>
                <ul
                  v-if="detailOperatingHours.length"
                  class="m-0 grid gap-2 p-0"
                >
                  <li
                    v-for="hour in detailOperatingHours"
                    :key="hour"
                    class="list-none rounded-lg border border-slate-200 bg-white px-3 py-2 text-[0.9rem] text-slate-700"
                  >
                    {{ hour }}
                  </li>
                </ul>
                <p
                  v-else
                  class="m-0 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-3 py-2 text-[0.9rem] text-slate-600"
                >
                  {{ t("panel.noData") }}
                </p>
              </section>

              <section v-if="detailReviewSnippets.length" class="grid gap-2">
                <p
                  class="m-0 text-xs font-semibold uppercase tracking-[0.12em] text-slate-500"
                >
                  {{ t("panel.detailReviewHighlights") }}
                </p>
                <ul class="m-0 grid gap-2.5 p-0">
                  <li
                    v-for="snippet in detailReviewSnippets"
                    :key="snippet"
                    class="list-none rounded-xl border border-slate-200 bg-white px-3 py-2 text-[0.92rem] text-slate-700"
                  >
                    {{ snippet }}
                  </li>
                </ul>
              </section>

              <section class="grid gap-2">
                <p
                  class="m-0 text-[1.1rem] font-semibold tracking-[-0.01em] text-slate-900"
                >
                  {{ t("panel.detailGallery") }}
                </p>
                <div
                  v-if="detailGallery.length"
                  class="grid grid-cols-2 gap-2 sm:grid-cols-3"
                >
                  <a
                    v-for="image in detailGallery"
                    :key="image.url"
                    :href="image.url"
                    target="_blank"
                    rel="noopener noreferrer"
                    class="group overflow-hidden rounded-xl border border-slate-200 bg-slate-100"
                  >
                    <img
                      :src="image.url"
                      :alt="image.caption || detailSummary.title"
                      class="h-24 w-full object-cover transition duration-200 group-hover:scale-[1.04]"
                      loading="lazy"
                    />
                  </a>
                </div>
                <p
                  v-else
                  class="m-0 rounded-xl border border-dashed border-slate-300 bg-slate-50 px-3 py-2 text-[0.9rem] text-slate-600"
                >
                  {{ t("panel.detailNoPhoto") }}
                </p>
              </section>
            </section>
          </div>
        </div>
      </div>
    </div>
  </aside>
</template>
