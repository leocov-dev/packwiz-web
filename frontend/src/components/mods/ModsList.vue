<script setup lang="ts">
import type {Mod} from "@/interfaces/pack.ts";

const {packId, mods, canEdit} = defineProps<{
  packId: number,
  mods: Mod[],
  canEdit: boolean,
}>()

defineEmits(['add-mod'])

const search = ref<string>('')

const sortedMods = computed(() => {
  const regularMods = mods.filter(mod => !mod.isDependency)
    .sort((a, b) => a.name.localeCompare(b.name));

  const dependencyMods = mods.filter(mod => mod.isDependency)
    .sort((a, b) => a.name.localeCompare(b.name));

  return [...regularMods, ...dependencyMods];
})

const isFirstDependency = (mod: Mod, items: readonly Mod[], index: number) => {
  if (!mod.isDependency) return false;
  return index === 0 || !items[index - 1].isDependency;
};

</script>

<template>
  <v-data-iterator
    :items="sortedMods"
    :search="search"
    items-per-page="0"
  >
    <template #header>
      <v-toolbar class="d-flex flex-wrap">
        <v-toolbar-title>Mods</v-toolbar-title>
        <v-text-field
          v-model="search"
          max-width="300"
          class="me-3"
          density="compact"
          placeholder="Search"
          prepend-inner-icon="mdi-magnify"
          variant="solo"
          clearable
          hide-details
        />
        <v-btn
          v-if="canEdit"
          class="me-3"
          color="primary"
          variant="flat"
          prepend-icon="mdi-plus"
          text="Add Mod"
          @click="$emit('add-mod')"
        />
      </v-toolbar>
    </template>

    <template #default="{items}">
      <v-list>
        <v-list-item
          v-for="(item, index) in items"
          :key="item.raw.name"
          :class="{'first-dependency': isFirstDependency(item.raw, items.map(i => i.raw), index)}"
        >
          <ModCard
            :pack-id="packId"
            :mod="item.raw"
          />
        </v-list-item>
      </v-list>
    </template>
  </v-data-iterator>
</template>

<style scoped>
.first-dependency {
  margin-top: 24px !important;
}
</style>
