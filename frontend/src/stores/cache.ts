import { defineStore } from 'pinia';
import axios from "axios";


interface CacheState {
  minecraftVersions: string[]
}

interface CacheActions {
  initializeVersions(): Promise<void>
}

interface CacheGetters {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}


export const useCacheStore = defineStore<'cache', CacheState, CacheGetters, CacheActions>('cache', {
  state: () => ({
    minecraftVersions: [],
  }),

  actions: {
    async initializeVersions() {
      this.minecraftVersions = await getAvailableVersions();
    }
  }

});


interface VersionResponse {
  versions: {
    id: string,
    type: string,
  }[]
}

const getAvailableVersions = async () => {
  const response = await axios.get<VersionResponse>('https://launchermeta.mojang.com/mc/game/version_manifest.json')
  return response.data.versions.filter(v => v.type == "release").map(v => v.id)
}

export async function initializeCacheStore() {
  const store = useCacheStore();
  await store.initializeVersions();
}
