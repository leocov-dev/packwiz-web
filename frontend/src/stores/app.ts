import {defineStore} from 'pinia';
import {apiClient} from "@/services/api.service.ts";


interface AppState {
  version: string
  curseforgeAvailable: boolean
}

interface AppActions {
  initialize(): Promise<void>
}

interface AppGetters {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}

interface StaticData {
  version: string
  curseforgeAvailable: boolean
}

export const useAppStore = defineStore<'app', AppState, AppGetters, AppActions>('app', {
  state: () => ({
    version: "",
    curseforgeAvailable: false,
  }),

  actions: {
    async initialize() {
      const staticData = await getStaticData();

      this.version = staticData.version
      this.curseforgeAvailable = staticData.curseforgeAvailable
    }
  }

});

const getStaticData = async () => {
  const response = await apiClient.get<StaticData>('/v1/static-data')
  return response.data
}

export async function initializeAppStore() {
  const store = useAppStore();
  await store.initialize();
}
