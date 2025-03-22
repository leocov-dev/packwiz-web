import {defineStore} from 'pinia';
import {apiClient} from "@/services/api.service.ts";
import {toTitleCase} from "@/services/utils.ts";


interface CacheState {
  minecraftLatest: string
  minecraftSnapshot: string
  minecraftVersions: string[]
  loaders: string[]
  loaderVersions: LoaderVersions
}

interface CacheActions {
  initializeVersions(): Promise<void>
}

interface CacheGetters {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any;
}

export interface LoaderVersions {
  fabric: string[]
  forge: Record<string, string[]>
  liteloader: string[]
  quilt: string[]
  neoforge: Record<string, string[]>
}

export interface VersionData {
  minecraft: {
    latest: string,
    snapshot: string,
    versions: string[]
  }
  loaders: LoaderVersions
}


export const useCacheStore = defineStore<'cache', CacheState, CacheGetters, CacheActions>('cache', {
  state: () => ({
    minecraftLatest: "",
    minecraftSnapshot: "",
    minecraftVersions: [],
    loaders: [],
    loaderVersions: {
      fabric: [],
      forge: {},
      neoforge: {},
      quilt: [],
      liteloader: [],
    }
  }),

  actions: {
    async initializeVersions() {
      const versionData = await getVersionData();

      this.minecraftLatest = versionData.minecraft.latest
      this.minecraftSnapshot = versionData.minecraft.snapshot
      this.minecraftVersions = versionData.minecraft.versions
      this.loaderVersions = versionData.loaders
      this.loaders = Object.keys(this.loaderVersions).map(loader => toTitleCase(loader));
    }
  }

});

const getVersionData = async () => {
  const response = await apiClient.get<VersionData>('/v1/packwiz/loaders')
  return response.data
}

export async function initializeCacheStore() {
  const store = useCacheStore();
  await store.initializeVersions();
}
