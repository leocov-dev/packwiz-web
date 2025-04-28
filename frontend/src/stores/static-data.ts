import {defineStore} from 'pinia';
import {apiClient} from "@/services/api.service.ts";


interface StaticDataState {
  version: string
}

interface StaticDataActions {
  initializeData(): Promise<void>
}

interface StaticDataGetters {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}


export const useStaticDataStore = defineStore<'staticData', StaticDataState, StaticDataGetters, StaticDataActions>(
  'staticData',
  {
    state: () => ({
      version: "",
    }),

    actions: {
      async initializeData() {
        const data = await getStaticData();

        this.version = data.version
      },
    },
  },
);

const getStaticData = async () => {
  const response = await apiClient.get<StaticDataState>('/v1/static')
  return response.data
}

export async function initializeStaticData() {
  const store = useStaticDataStore();
  await store.initializeData();
}
