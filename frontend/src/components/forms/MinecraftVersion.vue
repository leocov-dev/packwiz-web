<script lang="ts">
// Sentinel values sent as the request's `minecraft.version` when "Latest"/"Latest
// Snapshot" is selected; PackEditForm.vue and PackMigrateDialog.vue match against
// these exact strings to set `latest`/`snapshot` flags instead. Keeping the model
// value equal to the sentinel (with the real version only in the displayed title)
// is what makes that equality check reachable.
export const LATEST_SENTINEL = "Latest"
export const LATEST_SNAPSHOT_SENTINEL = "Latest Snapshot"
</script>

<script setup lang="ts">
import {useCacheStore} from "@/stores/cache.ts";

const cacheStore = useCacheStore()

const {includeLatest} = defineProps({includeLatest: Boolean})

type VersionItem = string | { title: string, value: string }

let versionsList: VersionItem[] = cacheStore.minecraftVersions
if (includeLatest) {
  versionsList = [
    {title: `Latest (${cacheStore.minecraftLatest})`, value: LATEST_SENTINEL},
    {title: `Latest Snapshot (${cacheStore.minecraftSnapshot})`, value: LATEST_SNAPSHOT_SENTINEL},
    ...versionsList
  ]
}

const version = defineModel<string | undefined>('version', {required: true})

const itemValue = (item: VersionItem) => typeof item === "string" ? item : item.value

if (!!version.value && !versionsList.some(item => itemValue(item) === version.value)) {
  versionsList = [
    version.value,
    ...versionsList,
  ]
}

const versions = ref<VersionItem[]>(versionsList)

const rules = {
  versionRequired: (value: string) => !!value || "Minecraft Version is required",
}

</script>

<template>
  <v-select
    v-model="version"
    :rules="[rules.versionRequired]"
    :items="versions"
    label="Minecraft Version"
    hint="Minecraft server version compatible with this pack."
    persistent-hint
    persistent-placeholder
  />
</template>
