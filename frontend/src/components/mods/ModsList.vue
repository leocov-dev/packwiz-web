<script setup lang="ts">
import type {ModData} from "@/interfaces/pack.ts";

const {mods, canEdit, disabled} = defineProps<{
  mods: ModData[],
  canEdit: boolean,
  disabled: boolean,
}>()

defineEmits(['add-mod'])

const search = ref<string>('')
</script>

<template>
  <v-data-iterator
    :items="mods"
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
          v-if="canEdit && !disabled"
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
          v-for="item in items"
          :key="item.raw.name"
        >
          {{ item.raw.name }}
        </v-list-item>
      </v-list>
    </template>
  </v-data-iterator>
</template>
