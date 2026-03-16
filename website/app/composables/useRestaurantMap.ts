import { storeToRefs } from 'pinia'
import { useRestaurantMapStore } from '~/stores/restaurantMap'

export function useRestaurantMap() {
  const store = useRestaurantMapStore()

  return {
    ...storeToRefs(store),
    fetchRestaurantsByCoordinate: store.fetchRestaurantsByCoordinate,
    openRestaurantDetail: store.openRestaurantDetail,
    closeRestaurantDetail: store.closeRestaurantDetail,
    clearDetailCache: store.clearDetailCache
  }
}
